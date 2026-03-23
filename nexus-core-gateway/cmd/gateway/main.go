package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/nexus-cyber/nexus-core-gateway/internal/ai"
	"github.com/nexus-cyber/nexus-core-gateway/internal/mtd"
	"github.com/nexus-cyber/nexus-core-gateway/internal/proxy"
	"github.com/nexus-cyber/nexus-core-gateway/pkg/logger"
)

// Minimal .env loader for Zero-Dependency Native Nexus Architecture
func loadEnv() {
	file, err := os.Open(".env")
	if err != nil {
		return // Silently fallback to os.Getenv
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			os.Setenv(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}
}

func main() {
	loadEnv()
	fmt.Println("[NEXUS] NEXUS CYBER GATEWAY - ENTERPRISE PRODUCTION INITIALIZING...")

	// 0. Initialize Distributed State (Redis via ISO-25010 Fallback)
	mtd.InitRedis()

	// 1. Initialize Intelligence Components
	filter := ai.NewReflexFilter()
	reasoning := ai.NewReasoningEngine("http://localhost:11434", "llama3")
	telemetry, err := logger.NewLogger()
	if err != nil {
		log.Fatalf("[NEXUS] Failed to initiate logger: %v", err)
	}
	defer telemetry.Close()

	// Register Real-time AI Event Streaming (Powering the Command Center SOC Terminal)
	telemetry.OnAIEvent = func(event logger.AIEventLog) {
		if mtd.MtdRedis != nil && mtd.MtdRedis.Enabled {
			data, _ := json.Marshal(event)
			// Broadcast to all active Dashboard SSH-Tunnel-SSE sessions
			mtd.MtdRedis.Client.Publish(context.Background(), "nexus:ai_stream", data)
		}
	}

	// 2. MTD: Token Bucket Rate Limiter (closes GAP-004)
	// 5 burst capacity, 5 req/sec sustained rate to allow synchronous testing to fail properly
	rateLimiter := mtd.NewTokenBucket(5, 5)
	rateLimiter.OnRateLimit = func(r *http.Request) {
		tLog := logger.TelemetryLog{
			Timestamp:    time.Now(),
			SourceIP:     r.RemoteAddr,
			Endpoint:     r.URL.Path,
			Method:       r.Method,
			Status:       "RATE_LIMITED",
			ThreatDetail: "RATE_LIMIT_EXCEEDED",
			LatencyMS:    0,
		}
		telemetry.EnrichLog(&tLog, r) // Call EnrichLog so the Dashboard shows Forensics
		telemetry.LogTraffic(tLog)
	}

	// 3. MTD: Digital Hallucination Honeypot
	// Runs on :9090, stalls attackers for 8 seconds, fully isolated
	honeypot := mtd.NewHoneypot(":9090", 8*time.Second)
	honeypot.Start()

	// 4. Setup Initial Backend Target (Mockup OJK Data Center)
	backendHost := os.Getenv("TARGET_BACKEND_HOST")
	if backendHost == "" {
		backendHost = "host.docker.internal" // Default to Docker Desktop's host bridge
	}

	target := os.Getenv("TARGET_BACKEND")
	if target == "" {
		target = fmt.Sprintf("http://%s:3001", backendHost)
	}

	// 5. MTD: Topology Shuffler (CSPRNG port rotation)
	// We use the same backendHost to support Docker -> Host communication.
	var gateway *proxy.NexusProxy

	shuffler := mtd.NewTopologyShuffler(
		backendHost, // baseHost
		[]int{3001}, // portPool (hanya 3001 untuk test mockup OJK)
		60,          // rotate every 60 seconds
		func(newTarget mtd.TargetBackend) {
			// Graceful Handoff: Update proxy without dropping in-flight requests
			if gateway != nil {
				if err := gateway.UpdateTarget(newTarget.URL()); err != nil {
					log.Printf("[MTD] Handoff failed: %v", err)
				} else {
					log.Printf("[MTD] Graceful handoff complete -> %s", newTarget.URL())
				}
			}
		},
	)
	shuffler.Start()

	// 6. Initialize MTD-aware Proxy
	gateway, err = proxy.NewNexusProxy(target, filter, reasoning, telemetry, shuffler, honeypot)
	if err != nil {
		log.Fatalf("[NEXUS] Failed to initiate proxy: %v", err)
	}

	// 7. Chain: BrowserIntegrity -> TokenBucket -> NexusProxy (defense-in-depth)
	gatewayHandler := proxy.BrowserIntegrityCheck(rateLimiter.HTTPMiddleware(gateway))

	// 8. Start Server wrapped in top-level mux for internal APIs
	mux := http.NewServeMux()
	mux.HandleFunc("/api/routes", routesHandler(gateway.Router)) // Zero-Code Onboarding
	mux.HandleFunc("/api/telemetry", telemetryHandler(shuffler, telemetry, target))
	mux.HandleFunc("/api/ai-events", aiEventsHandler(telemetry))               // AI Cognitive Core Tracker
	mux.HandleFunc("/api/ai/stream", aiStreamHandler())                        // SSE for Live CLI
	mux.HandleFunc("/api/ai/status", aiStatusHandler())                        // Health Check
	mux.HandleFunc("/api/cli/execute", cliExecuteHandler(telemetry))           // Interactive Terminal CLI
	mux.HandleFunc("/api/logs", telemetryHandler(shuffler, telemetry, target)) // Phase 6 requirement
	mux.HandleFunc("/api/domains", domainsHandler(telemetry))                  // Multi-Tenant Workspace Switcher
	mux.HandleFunc("/api/nechat", nechatHandler(telemetry))                    // Phase 6 Nechat Assist
	mux.HandleFunc("/api/panic", panicHandler(shuffler, telemetry))            // Phase 6 Rescue Protocol
	mux.HandleFunc("/api/verify-session", gateway.VerifySessionHandler)        // CGNAT Bypass Challenge Validator
	mux.Handle("/", gatewayHandler)                                            // all other requests go to the proxy

	// 9. Root Matrix Shield: Wrap EVERYTHING in AI Intelligence
	// 3. CORS Shield (Access for Dashboard)
	corsShield := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	rootShield := gateway.AIMiddleware(corsShield(mux))

	port := ":8080"
	fmt.Printf("[NEXUS] Gateway Active on port %s -> Proxying to %s\n", port, target)
	fmt.Println("[NEXUS] MODE: Phase 5 MTD Active | Honeypot: :9090 | Rate Limiter: 50r/s")

	// NEXUS_FIX: [INITIAL_HEARTBEAT]
	// Injecting initial status to populate 'Autonomous Operations' on start.
	telemetry.LogAIEvent(logger.AIEventLog{
		Timestamp:    time.Now(),
		Layer:        "Self-Repair",
		Status:       "SYSTEM_READY",
		DetailAction: "Nexus Cyber Intelligence established. Adaptive Matrix Layer 7 active and monitoring traffic vectors.",
	})

	if err := http.ListenAndServe(port, rootShield); err != nil {
		log.Fatal(err)
	}
}
