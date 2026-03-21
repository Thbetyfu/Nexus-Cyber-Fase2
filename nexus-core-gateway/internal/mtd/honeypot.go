package mtd

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"time"
)

// HoneypotServer implements @skill-mtd Digital Hallucination.
// It mimics a real server (HTTP 200 OK) but with a cryptographically random
// Tarpit delay (5-10s) to exhaust attacker resources and defeat timing analysis.
type HoneypotServer struct {
	ListenAddr  string
	MinTarpit   time.Duration // min stall duration
	MaxTarpit   time.Duration // max stall duration — random range adds unpredictability
	FakeVersion string        // server version string to deceive fingerprinting
}

// NewHoneypot creates a honeypot with random tarpit delay in [5s, 10s].
// The tarpitDelay param is retained for API compatibility but ignored — range is now fixed.
func NewHoneypot(addr string, tarpitDelay time.Duration) *HoneypotServer {
	return &HoneypotServer{
		ListenAddr:  addr,
		MinTarpit:   5 * time.Second,
		MaxTarpit:   10 * time.Second,
		FakeVersion: "nginx/1.18.0 (Ubuntu)",
	}
}

// Start launches the honeypot HTTP server on a background goroutine.
// ISOLATION GUARANTEE: This server has NO access to the main backend.
// It only serves fake responses from memory.
func (h *HoneypotServer) Start() {
	mux := http.NewServeMux()

	// Mimic common API endpoints to look real
	mux.HandleFunc("/api/", h.tarpitHandler)
	mux.HandleFunc("/get", h.tarpitHandler)
	mux.HandleFunc("/post", h.tarpitHandler)
	mux.HandleFunc("/", h.tarpitHandler)

	srv := &http.Server{
		Addr:    h.ListenAddr,
		Handler: mux,
	}

	go func() {
		log.Printf("[HONEYPOT] Digital Hallucination server ACTIVE on %s (random tarpit: %v-%v)",
			h.ListenAddr, h.MinTarpit, h.MaxTarpit)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[HONEYPOT] Error: %v", err)
		}
	}()
}

// tarpitHandler is the core Digital Hallucination logic.
// Random delay [5s, 10s] defeats attacker timing calibration and drains resources.
func (h *HoneypotServer) tarpitHandler(w http.ResponseWriter, r *http.Request) {
	attackerIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		attackerIP = r.RemoteAddr
	}

	// 1. Record IP to Distributed Mem Store (Redis) with TTL
	if MtdRedis != nil && MtdRedis.Enabled {
		ctx := r.Context()
		// Lock IP globally for 24 hours
		err := MtdRedis.Client.Set(ctx, "honeypot:"+attackerIP, time.Now().String(), 24*time.Hour).Err()
		if err != nil {
			log.Printf("[HONEYPOT-REDIS] Failed to record attacker IP: %v", err)
		} else {
			log.Printf("[HONEYPOT-REDIS] Recorded attacker IP '%s' with 24h TTL.", attackerIP)
		}
	}

	// 2. Log attacker fingerprint for forensics
	log.Printf("[HONEYPOT-TRAP] Attacker caught: IP=%s | Path=%s | UA=%s",
		r.RemoteAddr, r.URL.Path, r.Header.Get("User-Agent"))

	// 3. Cryptographically random tarpit within [MinTarpit, MaxTarpit]
	delay := h.randomTarpit()
	log.Printf("[HONEYPOT-TARPIT] Stalling %s for %v...", r.RemoteAddr, delay.Round(time.Millisecond))
	time.Sleep(delay)

	// Craft convincing fake response — mimic real backend
	w.Header().Set("Server", h.FakeVersion)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", generateFakeRequestID())
	w.WriteHeader(http.StatusOK)

	// Return plausible but empty data to confuse attacker
	fmt.Fprintf(w, `{
  "status": "success",
  "data": {},
  "server_time": "%s",
  "request_id": "%s"
}`, time.Now().Format(time.RFC3339), generateFakeRequestID())
}

// randomTarpit returns a cryptographically random duration between MinTarpit and MaxTarpit.
func (h *HoneypotServer) randomTarpit() time.Duration {
	delta := h.MaxTarpit - h.MinTarpit
	if delta <= 0 {
		return h.MinTarpit
	}
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(delta)))
	if err != nil {
		return h.MinTarpit // safe fallback
	}
	return h.MinTarpit + time.Duration(nBig.Int64())
}

// generateFakeRequestID generates a cryptographically random UUID-like identifier.
func generateFakeRequestID() string {
	b := make([]byte, 16)
	rand.Read(b) //nolint:errcheck
	return fmt.Sprintf("%08x-%04x-4%03x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
