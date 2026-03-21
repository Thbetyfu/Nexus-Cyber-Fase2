package mtd

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ipBucket holds per-IP token bucket state.
type ipBucket struct {
	tokens     float64
	lastRefill time.Time
}

// PerIPTokenBucket implements per-source-IP rate limiting using Token Bucket.
// Upgrade dari versi global: setiap IP memiliki bucket-nya sendiri.
// Capacity: max burst per IP (e.g. 100 requests)
// RefillRate: sustained rate per IP per second (e.g. 100 req/s)
type PerIPTokenBucket struct {
	mu         sync.Mutex
	buckets    map[string]*ipBucket
	capacity   float64
	refillRate float64
	// janitor cleans up stale IP entries periodically
	cleanupInterval time.Duration
	// optional callback for telemetry
	OnRateLimit func(r *http.Request)
}

// NewPerIPTokenBucket creates a per-IP rate limiter.
// capacity: max burst per unique source IP
// refillRate: tokens added per second per IP
func NewPerIPTokenBucket(capacity, refillRate float64) *PerIPTokenBucket {
	tb := &PerIPTokenBucket{
		buckets:         make(map[string]*ipBucket),
		capacity:        capacity,
		refillRate:      refillRate,
		cleanupInterval: 5 * time.Minute,
	}
	go tb.janitor()
	return tb
}

// getRealIP extracts the true client IP from a request.
// Priority: X-Forwarded-For (first entry) -> X-Real-IP -> RemoteAddr.
// This fix addresses FINDING-B01: localhost deployments behind a proxy or during
// testing with spoofed X-Forwarded-For headers will now correctly isolate per-IP.
func getRealIP(r *http.Request) string {
	// X-Forwarded-For may contain a comma-separated chain: "client, proxy1, proxy2"
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		for _, part := range parts {
			ip := strings.TrimSpace(part)
			if ip != "" {
				return ip
			}
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}
	// Fallback: strip port from RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// Allow checks if the given source IP has tokens remaining.
// sourceIP should be the raw string from RemoteAddr; use HTTPMiddleware for HTTP.
// Returns true if allowed, false if rate-limited.
// Allow checks if the given source IP has tokens remaining.
// sourceIP should be the raw string from RemoteAddr; use HTTPMiddleware for HTTP.
// Returns true if allowed, false if rate-limited.
func (tb *PerIPTokenBucket) Allow(sourceIP string) bool {
	// Strip port if present
	ip, _, err := net.SplitHostPort(sourceIP)
	if err != nil {
		ip = sourceIP
	}

	if MtdRedis != nil && MtdRedis.Enabled {
		return tb.allowRedis(ip)
	}
	return tb.allowLocal(ip)
}

// allowRedis uses atomic Lua scripting in Redis for Token Bucket.
func (tb *PerIPTokenBucket) allowRedis(ip string) bool {
	ctx := context.Background()
	// Using a simple fixed-window or INCR fallback for Redis simplicity in this prototype
	// For production: a correct Lua TokenBucket script. Here is a simple implementation:
	script := `
		local key = KEYS[1]
		local capacity = tonumber(ARGV[1])
		local rate = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		local requested = 1

		local info = redis.call('HMGET', key, 'tokens', 'last_refill')
		local tokens = tonumber(info[1])
		local last_refill = tonumber(info[2])

		if not tokens then
			tokens = capacity
			last_refill = now
		end

		local elapsed = now - last_refill
		tokens = math.min(capacity, tokens + elapsed * rate)

		if tokens >= requested then
			tokens = tokens - requested
			redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
			redis.call('EXPIRE', key, math.ceil(capacity/rate) + 10)
			return 1
		end
		
		redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
		redis.call('EXPIRE', key, math.ceil(capacity/rate) + 10)
		return 0
	`
	now := float64(time.Now().UnixNano()) / 1e9 // seconds
	res, err := MtdRedis.Client.Eval(ctx, script, []string{"tb:" + ip}, tb.capacity, tb.refillRate, now).Result()
	if err != nil {
		log.Printf("[MTD-REDIS] Eval Error: %v. Falling back to memory.", err)
		return tb.allowLocal(ip)
	}
	return res.(int64) == 1
}

// allowLocal is the local fallback Memory-based Token Bucket (Zero-Dependency)
func (tb *PerIPTokenBucket) allowLocal(ip string) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	b, exists := tb.buckets[ip]
	if !exists {
		b = &ipBucket{tokens: tb.capacity, lastRefill: now}
		tb.buckets[ip] = b
	}

	// Refill based on elapsed time
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.lastRefill = now
	b.tokens += elapsed * tb.refillRate
	if b.tokens > tb.capacity {
		b.tokens = tb.capacity
	}

	if b.tokens >= 1.0 {
		b.tokens--
		return true
	}
	return false
}

// HTTPMiddleware wraps an HTTP handler with per-IP Token Bucket rate limiting.
// Uses getRealIP() to correctly identify client behind proxies/CDN.
// Returns HTTP 429 with a Retry-After header when IP bucket is exhausted.
func (tb *PerIPTokenBucket) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		realIP := getRealIP(r)
		if !tb.Allow(realIP) {
			log.Printf("[MTD-RATELIMIT] IP THROTTLED (Redis Integration Active): %s (real) / %s (remote) — >%.0f req/s",
				realIP, r.RemoteAddr, tb.refillRate)
			if tb.OnRateLimit != nil {
				tb.OnRateLimit(r)
			}
			w.Header().Set("Retry-After", "1")
			http.Error(w,
				`{"error":"rate_limit_exceeded","message":"Too many requests from your IP","retry_after":"1s"}`,
				http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// janitor periodically removes stale IP entries to prevent memory growth.
func (tb *PerIPTokenBucket) janitor() {
	ticker := time.NewTicker(tb.cleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-tb.cleanupInterval)
		tb.mu.Lock()
		for ip, b := range tb.buckets {
			if b.lastRefill.Before(cutoff) {
				delete(tb.buckets, ip)
			}
		}
		tb.mu.Unlock()
	}
}

// TokenBucket is kept for backward compatibility with main.go.
// New code should use PerIPTokenBucket.
type TokenBucket = PerIPTokenBucket

// NewTokenBucket is a backward-compatible alias with global-style params.
// Internally creates a PerIPTokenBucket.
func NewTokenBucket(capacity, refillRate float64) *PerIPTokenBucket {
	return NewPerIPTokenBucket(capacity, refillRate)
}
