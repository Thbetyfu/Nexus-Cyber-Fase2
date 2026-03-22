package proxy

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/nexus-cyber/nexus-core-gateway/internal/mtd"
)

// RouteEntry stores the target URL and expiration for in-memory caching.
type RouteEntry struct {
	TargetURL string
	ExpiresAt time.Time
}

// DynamicRouter manages domain-to-backend mappings with Redis + In-Memory Cache.
// Implements ISO-25010 Efficiency & Reliability.
type DynamicRouter struct {
	cache map[string]RouteEntry
	mu    sync.RWMutex
	ttl   time.Duration
}

func NewDynamicRouter(cacheTTL time.Duration) *DynamicRouter {
	return &DynamicRouter{
		cache: make(map[string]RouteEntry),
		ttl:   cacheTTL,
	}
}

// Lookup finds the target URL for a given domain/host.
// Checks local cache first, then falls back to Redis.
func (dr *DynamicRouter) Lookup(host string) (string, bool) {
	// 1. Check Local In-Memory Cache (Performance Layer)
	dr.mu.RLock()
	entry, exists := dr.cache[host]
	dr.mu.RUnlock()

	if exists && time.Now().Before(entry.ExpiresAt) {
		return entry.TargetURL, true
	}

	// 2. Fallback to Redis (Distributed Layer)
	if mtd.MtdRedis != nil && mtd.MtdRedis.Enabled {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		target, err := mtd.MtdRedis.Client.HGet(ctx, "nexus:routes", host).Result()
		if err == nil && target != "" {
			// Update Local Cache
			dr.mu.Lock()
			dr.cache[host] = RouteEntry{
				TargetURL: target,
				ExpiresAt: time.Now().Add(dr.ttl),
			}
			dr.mu.Unlock()
			return target, true
		}
	}

	// 3. Fallback to default (if it's localhost:8080 or other known defaults)
	// For demo purposes, we keep reflecting the primary target if host is empty or unknown
	// but the requirement says 404 if not found.
	return "", false
}

// AddRoute manually injects a route into Redis and clears cache.
func (dr *DynamicRouter) AddRoute(host, target string) error {
	// 1. Update In-Memory Cache (Immediate availability)
	dr.mu.Lock()
	dr.cache[host] = RouteEntry{
		TargetURL: target,
		ExpiresAt: time.Now().Add(dr.ttl),
	}
	dr.mu.Unlock()

	// 2. Global Sync (Persistence Layer)
	if mtd.MtdRedis != nil && mtd.MtdRedis.Enabled {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := mtd.MtdRedis.Client.HSet(ctx, "nexus:routes", host, target).Err()
		if err != nil {
			return err
		}
	}

	log.Printf("[ROUTER] Mapping established: %s -> %s", host, target)
	return nil
}

// GetAllRoutes retrieves all registered routes from Redis.
func (dr *DynamicRouter) GetAllRoutes() (map[string]string, error) {
	if mtd.MtdRedis != nil && mtd.MtdRedis.Enabled {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		return mtd.MtdRedis.Client.HGetAll(ctx, "nexus:routes").Result()
	}
	return nil, nil
}
