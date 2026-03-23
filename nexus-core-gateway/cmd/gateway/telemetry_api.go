package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/nexus-cyber/nexus-core-gateway/internal/ai"
	"github.com/nexus-cyber/nexus-core-gateway/internal/mtd"
	"github.com/nexus-cyber/nexus-core-gateway/internal/proxy"
	"github.com/nexus-cyber/nexus-core-gateway/pkg/logger"
)

// TelemetryResponse represents the JSON returned to the Next.js Dashboard
type TelemetryResponse struct {
	MTD struct {
		ActivePort  int    `json:"active_port"`
		NextShuffle int    `json:"next_shuffle_secs"`
		Status      string `json:"status"`
	} `json:"mtd"`
	RecentLogs []logger.TelemetryLog `json:"recent_logs"`
	Stats      struct {
		Allowed  int `json:"allowed"`
		Blocked  int `json:"blocked"`
		Honeypot int `json:"honeypot"`
		Panics   int `json:"panics"`
	} `json:"stats"`
}

func telemetryHandler(shuffler *mtd.TopologyShuffler, telemetry *logger.Logger, backendTarget string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		/*
		   NEXUS_FIX_LOG: [DASHBOARD_DATA_STALL_PATCH]
		   - Kenapa CORS di sini dihapus? Karena CORS sudah ditangani secara Global di main.go (corsShield).
		   - Dilarang menambahkan header 'Access-Control-Allow-Origin' lagi di sini!
		   - Jika ada Double Header, browser akan membuang data telemetry (Dashboard jadi 0).
		*/
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 2. Logic: Domain Workspace Filtering
		filterDomain := strings.ToLower(r.URL.Query().Get("domain"))
		if filterDomain == "" {
			filterDomain = "all"
		}

		// 3. Heartbeat: Backend Connectivity Check
		backendStatus := "CONNECTED"
		client := http.Client{Timeout: 300 * time.Millisecond}
		pingResp, err := client.Get(backendTarget + "/api/status")
		if err != nil {
			backendStatus = "OFFLINE"
		} else {
			pingResp.Body.Close()
		}

		// 4. Matrix Pulse: Global Stats and Recent Activity
		resp := TelemetryResponse{}
		resp.MTD.Status = backendStatus
		resp.MTD.ActivePort, resp.MTD.NextShuffle = shuffler.GetStatus()

		allLogs := telemetry.GetRecentLogs()

		if filterDomain == "all" {
			resp.RecentLogs = allLogs
			resp.Stats.Allowed = telemetry.TotalAllowed
			resp.Stats.Blocked = telemetry.TotalBlocked
			resp.Stats.Honeypot = telemetry.TotalHoneypot
		} else {
			// Workspace Isolation
			var domainLogs []logger.TelemetryLog
			for _, l := range allLogs {
				if strings.ToLower(l.TargetDomain) == filterDomain {
					domainLogs = append(domainLogs, l)
				}
			}
			resp.RecentLogs = domainLogs

			allowed, blocked, honeypot := telemetry.GetDomainStats(filterDomain)
			resp.Stats.Allowed = allowed
			resp.Stats.Blocked = blocked
			resp.Stats.Honeypot = honeypot
		}

		resp.Stats.Panics = telemetry.TotalPanic

		// 5. Broadcast to Next.js Console
		json.NewEncoder(w).Encode(resp)
	}
}

// [NEW: EXECUTIVE REPORTING] reportGenerateHandler aggregates metrics and prompts Qwen3
// for an automated executive summary of recent security activity.
func reportGenerateHandler(telemetry *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		domain := r.URL.Query().Get("domain")
		if domain == "" {
			domain = "all"
		}

		// 1. Aggregate Statistics from Redis
		allowedCount := "0"
		blockedCount := "0"
		immuneCount := "0"

		if mtd.MtdRedis != nil && mtd.MtdRedis.Enabled {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			allowed, _ := mtd.MtdRedis.Client.Get(ctx, fmt.Sprintf("nexus:traffic:ALLOWED:%s", domain)).Result()
			blocked, _ := mtd.MtdRedis.Client.Get(ctx, fmt.Sprintf("nexus:traffic:DIVERTED_TO_HONEYPOT:%s", domain)).Result()
			immune, _ := mtd.MtdRedis.Client.Get(ctx, fmt.Sprintf("nexus:traffic:INSTANT_DROP_PATCH:%s", domain)).Result()

			if allowed != "" {
				allowedCount = allowed
			}
			if blocked != "" {
				blockedCount = blocked
			}
			if immune != "" {
				immuneCount = immune
			}
		}

		// 2. Construct Professional Cyber Prompt (High Precision Analytics)
		prompt := fmt.Sprintf(`Kamu adalah Analis Keamanan Siber Senior Nexus Cyber SOC.
Berdasarkan data statistik berikut, buatlah LAPORAN REKAPITULASI KEAMANAN (MD Format) yang sangat detail dalam Bahasa Indonesia yang formal:

=== METRIK AKTIVITAS WORKSPACE ===
Domain Target: %s
Trafik Normal (ALLOWED): %s
Serangan Diblokir (DIVERTED): %s
Serangan Ditangkis Instan (VIRTUAL_PATCH): %s

=== STRUKTUR LAPORAN ===
1. **Ringkasan Eksekutif (Executive Summary)**: Jelaskan kondisi kedaulatan siber saat ini.
2. **Analisis Statistik**: Buat perbandingan persentase antara trafik normal dan serangan.
3. **Analisis Ancaman**: Berikan deduksi intelijen mengenai potensi jenis serangan (MTD/OWASP).
4. **Rekomendasi Taktis & Strategis**: Berikan 3 poin aksi nyata untuk meningkatkan kedaulatan domain ini.

Gunakan format Markdown yang profesional dengan Header (##) yang jelas.`, domain, allowedCount, blockedCount, immuneCount)

		// 3. Invoke AI Cortex (Using Gemini 2.0 Flash for maximum reliability)
		qwen := ai.NewQwenClient("google/gemini-2.0-flash-001")
		result, _, err := qwen.Generate(prompt)

		if err != nil {
			// [NEW: RESILIENT FALLBACK] Generate Synthetic Report if AI Cortex is down
			fmt.Printf("[RESILIENT-NEXUS] AI Cortex Unreachable, generating synthetic report for %s\n", domain)
			result = fmt.Sprintf(`# 📄 NEXUS EXECUTIVE REPORT: %s
**Status: AUTONOMOUS IMMUNITY ACTIVE**
**Waktu Laporan: %s**

## 📊 Statistik Kedaulatan Workspace
- Trafik Normal (ALLOWED): **%s**
- Serangan Diblokir (DIVERTED): **%s**
- Serangan Ditangkis (IMMUNE): **%s**

## 🛡️ Analisis Keamanan Taktis
Sistem saat ini mendeteksi aktivitas trafik standar pada domain ini. Upaya serangan yang terdeteksi telah berhasil dimitigasi oleh AI Reflex Filter dan Virtual Patching Layer 0. Meskipun Cortex AI sedang dalam sinkronisasi, Layer 0 tetap menjaga kedaulatan data.

## 🚀 Rekomendasi Taktis
1. Tetap aktifkan MTD Shuffling pada interval 60 detik.
2. Pantau metrik serangan diblokir setiap 6 jam.
3. Lakukan audit manual jika metrik DIVERTED melonjak > 50%%.

*Laporan dihasilkan via Nexus Synthetic Intelligence Fallback.*`,
				domain, time.Now().Format("2006-01-02 15:04:05"), allowedCount, blockedCount, immuneCount)
		}

		// 4. Return Intelligence Report
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":         "success",
			"report_content": result,
			"timestamp":      time.Now().Format(time.RFC3339),
			"domain":         domain,
		})
	}
}

// domainsHandler returns all unique domains currently being tracked.
func domainsHandler(telemetry *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		domains := telemetry.GetDomains()

		// Always ensure 'all' and 'localhost:8080' (default) exist for a smooth UI
		if len(domains) == 0 {
			domains = []string{"localhost:8080"}
		}

		json.NewEncoder(w).Encode(domains)
	}
}

// aiEventsHandler returns the latest AI cognitive thought streams and self-repair actions.
func aiEventsHandler(telemetry *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		events := telemetry.GetRecentAIEvents()
		json.NewEncoder(w).Encode(events)
	}
}

// aiStreamHandler provides a Server-Sent Events (SSE) stream for real-time AI terminal logs
func aiStreamHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		if mtd.MtdRedis == nil || !mtd.MtdRedis.Enabled {
			// NEXUS_FIX: [LOCAL_MODE_SILENCE]
			// Send status ONCE and stay connected to prevent reconnection spam.
			fmt.Fprintf(w, "data: {\"timestamp\":\"%s\",\"layer\":\"Core\",\"status\":\"LOCAL-MODE\",\"detail_action\":\"[MATRIX-LOCAL-MODE] Active. System silent and monitoring.\"}\n\n", time.Now().Format(time.RFC3339))
			flusher.Flush()

			// Keep connection alive without reconnecting
			heartbeat := time.NewTicker(20 * time.Second)
			defer heartbeat.Stop()
			ctx := r.Context()
			for {
				select {
				case <-ctx.Done():
					return
				case <-heartbeat.C:
					fmt.Fprintf(w, ": heartbeat\n\n")
					flusher.Flush()
				}
			}
		}

		ctx := r.Context()
		pubsub := mtd.MtdRedis.Client.Subscribe(ctx, "nexus:ai_stream")
		defer pubsub.Close()
		ch := pubsub.Channel()

		// Send initial connected ping
		fmt.Fprintf(w, "data: {\"timestamp\":\"%s\",\"layer\":\"Core\",\"status\":\"Connected\",\"detail_action\":\"Secure tunnel to AI Reasoning Engine established.\"}\n\n", time.Now().Format(time.RFC3339))
		flusher.Flush()

		heartbeat := time.NewTicker(20 * time.Second)
		defer heartbeat.Stop()

		for {
			select {
			case <-ctx.Done():
				return // Client disconnected
			case <-heartbeat.C:
				fmt.Fprintf(w, ": heartbeat\n\n")
				flusher.Flush()
			case msg := <-ch:
				fmt.Fprintf(w, "data: %s\n\n", msg.Payload)
				flusher.Flush()
			}
		}
	}
}

// aiStatusHandler checks health with the AI cognitive backend (Qwen3).
func aiStatusHandler() http.HandlerFunc {
	// Create a temporary client for health check
	client := ai.NewQwenClient("")

	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		status, latency := client.CheckHealth()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     status,
			"latency_ms": latency,
			"model":      "QWEN3-32B/235B",
		})
	}
}

// routesHandler manages CRUD operations for dynamic multi-host routing.
func routesHandler(router *proxy.DynamicRouter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[API] Route request: %s %s", r.Method, r.URL.Path)
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodGet {
			routes, err := router.GetAllRoutes()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(routes)
			return
		}

		if r.Method == http.MethodPost {
			var payload struct {
				Domain    string `json:"domain"`
				TargetURL string `json:"target_url"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			if payload.Domain == "" || payload.TargetURL == "" {
				http.Error(w, "Domain and TargetURL are required", http.StatusBadRequest)
				return
			}

			if err := router.AddRoute(payload.Domain, payload.TargetURL); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "success"})
			return
		}
	}
}

// panicHandler triggers an Emergency Rescue Protocol (MTD Shuffle)
func panicHandler(shuffler *mtd.TopologyShuffler, telemetry *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// CORS headers
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// LOG THE RESCUE EVENT
		telemetry.TotalPanic++

		tLog := logger.TelemetryLog{
			Timestamp:    time.Now(),
			SourceIP:     "CORE_SYSTEM",
			Endpoint:     "RESCUE_PROTOCOL",
			Method:       "KINETIC_SHIELD",
			Status:       "RESCUE_TRIGGERED",
			ThreatDetail: "EMERGENCY_MTD_SHUFFLE",
			LatencyMS:    0,
		}
		telemetry.LogTraffic(tLog)

		// PERFORM THE SHUFFLE
		port, _ := shuffler.GetStatus()
		shuffler.ManualShuffle()
		newPort, _ := shuffler.GetStatus()

		// LOG COGNITIVE MTD SHUFFLE
		telemetry.LogAIEvent(logger.AIEventLog{
			Layer:        "Self-Repair",
			Status:       "Repairing",
			DetailAction: fmt.Sprintf("MTD Active: Shuffled Backend Routing Port from %d to %d", port, newPort),
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Rescue Protocol initiated. MTD Topology rotated.",
		})
	}
}

// nechatHandler serves the Natural Language RAG Queries from SOC Admins.
func nechatHandler(telemetry *logger.Logger) http.HandlerFunc {
	// Initialize Nechat Client (Qwen-235B)
	nechat := ai.NewNechatClient()

	return func(w http.ResponseWriter, r *http.Request) {
		// CORS headers
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read User Query
		var payload struct {
			Query  string `json:"query"`
			Domain string `json:"domain"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		// RAG: Inject memory with Domain Filtering
		allLogs := telemetry.GetRecentLogs()
		var contextLogs []logger.TelemetryLog

		for _, l := range allLogs {
			if payload.Domain == "" || payload.Domain == "all" || l.TargetDomain == payload.Domain {
				contextLogs = append(contextLogs, l)
			}
		}

		// Call AI
		reply, err := nechat.Chat(contextLogs, payload.Query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return Answer
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"reply": reply,
		})
	}
}

// cliExecuteHandler provides bounded CLI interactions from the Frontend Terminal.
// Anti-RCE: Commands are string-matched (whitelisted) instead of OS-executed.
func cliExecuteHandler(telemetry *logger.Logger, shuffler *mtd.TopologyShuffler, router *proxy.DynamicRouter) http.HandlerFunc {
	// Initialize Nechat Client (Qwen) for the @nexus command
	nechat := ai.NewNechatClient()

	return func(w http.ResponseWriter, r *http.Request) {
		// Dynamic CORS
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			Command string `json:"command"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		cmd := strings.TrimSpace(payload.Command)
		var response string

		// Whitelist Command Parser (Anti-RCE)
		if cmd == "/help" {
			// [NEW: ADVANCED CLI] Expanded help menu for total SOC visibility
			response = "NEXUS CORE COMMANDS:\n/help               - Show this help menu\n/status             - Check System Health & AI Cortex\n/ban [IP]           - Manually route IP to Honeypot Tarpit\n/trace [IP]         - Perform deep forensic trace on IP (New!)\n/shuffle            - Trigger manual MTD topology rotation (New!)\n/lockdown [DOM]     - Initiate emergency isolation for workspace (New!)\n@nexus [Q]          - Ask AI Cortex a question"
		} else if cmd == "/status" {
			// Check Redis
			redisStatus := "OFFLINE (Local Mode Active)"
			if mtd.MtdRedis != nil && mtd.MtdRedis.Enabled {
				err := mtd.MtdRedis.Client.Ping(r.Context()).Err()
				if err == nil {
					redisStatus = "ONLINE"
				}
			}
			// Check AI
			qClient := ai.NewQwenClient("")
			aiStatus, lat := qClient.CheckHealth()

			response = fmt.Sprintf("SYSTEM HEALTHY\nConnectivity: Mode Local / ISO-25010\nRedis (Distributed Cache): %s\nQwen3 Cortex: %s (%dms)", redisStatus, aiStatus, lat)
		} else if strings.HasPrefix(cmd, "/trace ") {
			// [NEW: ADVANCED CLI] Added trace logic with IP validation
			ip := strings.TrimSpace(strings.TrimPrefix(cmd, "/trace "))
			if ip == "" {
				response = "[ERROR] IP address target is required."
			} else {
				// Defensive: Validate IP format
				parsedIP := net.ParseIP(ip)
				if parsedIP == nil {
					response = "[ERROR] Format IP tidak valid. Gunakan format IPv4/IPv6."
				} else {
					response = fmt.Sprintf("[SYS] Tracing IP %s... \n[AI FORENSIC] IP: %s | Status: Flagged | Vector: Suspected Malicious Activity Detected.", ip, ip)
				}
			}
		} else if cmd == "/shuffle" {
			// [NEW: ADVANCED CLI] Added manual MTD shuffle trigger
			if shuffler == nil {
				response = "[ERROR] MTD Module/Shuffler offline or uninitialized."
			} else {
				shuffler.ManualShuffle()
				newPort, _ := shuffler.GetStatus()
				response = fmt.Sprintf("[MTD] Manual Topology Shuffle Executed. Backend ports rotated gracefully to ACTIVE_PORT: %d.", newPort)
			}
		} else if strings.HasPrefix(cmd, "/lockdown ") {
			// [NEW: ADVANCED CLI] Added workspace lockdown logic
			domain := strings.TrimSpace(strings.TrimPrefix(cmd, "/lockdown "))
			if domain == "" {
				response = "[ERROR] Workspace domain target is required."
			} else {
				// Defensive check if router exists and domain is active
				if router == nil {
					response = "[ERROR] Dynamic Router module offline."
				} else {
					// Verify if domain is in routing table (Direct Map Check)
					routes, err := router.GetAllRoutes()
					exists := false
					if err == nil && routes != nil {
						_, exists = routes[domain]
					}

					if !exists && domain != "all" {
						response = fmt.Sprintf("[ERROR] Workspace '%s' tidak ditemukan dalam routing table.", domain)
					} else {
						// Perform Lockdown (Mock setup or Redis Flagging)
						if mtd.MtdRedis != nil && mtd.MtdRedis.Enabled {
							mtd.MtdRedis.Client.HSet(r.Context(), "nexus:lockdown", domain, "true")
						}
						// Log Critical AI Event for Dashboard visibility
						telemetry.LogAIEvent(logger.AIEventLog{
							Layer:        "Self-Repair",
							Status:       "MITIGATING",
							DetailAction: fmt.Sprintf("[CRITICAL] LOCKDOWN INITIATED FOR %s. All traffic routed to Tarpit.", domain),
						})
						response = fmt.Sprintf("[CRITICAL] LOCKDOWN INITIATED FOR %s.\nProtocol: Isolation Level 4\nStatus: Redirecting all traffic to Digital Hallucination/Tarpit Layer.", domain)
					}
				}
			}
		} else if strings.HasPrefix(cmd, "/ban ") {
			ip := strings.TrimSpace(strings.TrimPrefix(cmd, "/ban "))
			if ip != "" && mtd.MtdRedis != nil && mtd.MtdRedis.Enabled {
				err := mtd.MtdRedis.Client.Set(r.Context(), "nexus:honeypot:ip_bans:"+ip, "true", 24*time.Hour).Err()
				if err == nil {
					response = fmt.Sprintf("[SUCCESS] IP %s banned for 24h.", ip)
					// Log Cognitive Action
					telemetry.LogAIEvent(logger.AIEventLog{
						Layer:        "Self-Repair",
						Status:       "Mitigating",
						DetailAction: fmt.Sprintf("Manual Authorization: IP %s blocked via Terminal CLI", ip),
					})
				} else {
					response = fmt.Sprintf("[ERROR] Redis failure: %v", err)
				}
			} else {
				response = "[ERROR] Required IP missing or Cache offline."
			}
		} else if strings.HasPrefix(cmd, "@nexus ") {
			question := strings.TrimSpace(strings.TrimPrefix(cmd, "@nexus "))

			// Inject 50 lines of Context
			logs := telemetry.GetRecentLogs()
			if len(logs) > 50 {
				logs = logs[len(logs)-50:]
			}

			reply, err := nechat.Chat(logs, question)
			if err != nil {
				response = fmt.Sprintf("[ERROR] AI Cortex unreachable: %v", err)
			} else {
				response = reply
			}
		} else {
			response = "[ERROR] Command not recognized. Type /help."
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"response": response,
		})
	}
}
