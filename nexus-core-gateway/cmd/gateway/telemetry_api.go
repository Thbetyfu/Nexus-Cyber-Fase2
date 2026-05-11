package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/nexus-cyber/nexus-core-gateway/internal/ai"
	"github.com/nexus-cyber/nexus-core-gateway/internal/mtd"
	"github.com/nexus-cyber/nexus-core-gateway/internal/proxy"
	"github.com/nexus-cyber/nexus-core-gateway/pkg/logger"
)

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
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		filterDomain := strings.ToLower(r.URL.Query().Get("domain"))
		if filterDomain == "" {
			filterDomain = "all"
		}
		backendStatus := "CONNECTED"
		client := http.Client{Timeout: 300 * time.Millisecond}
		pingResp, err := client.Get(backendTarget + "/api/status")
		if err != nil {
			backendStatus = "OFFLINE"
		} else {
			pingResp.Body.Close()
		}
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
		json.NewEncoder(w).Encode(resp)
	}
}

func reportGenerateHandler(telemetry *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		domain := r.URL.Query().Get("domain")
		if domain == "" {
			domain = "all"
		}
		allowedCount, blockedCount, immuneCount := "0", "0", "0"
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
		prompt := fmt.Sprintf(`Identitas: Analis SOC Senior. Buat laporan MD formal untuk domain %s. Metrik: Allowed=%s, Diverted=%s, Immune=%s.`, domain, allowedCount, blockedCount, immuneCount)
		qwen := ai.NewQwenClient("google/gemini-2.0-flash-001")
		result, _, _ := qwen.Generate(prompt)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success", "report_content": result})
	}
}

func xxxDomainsHandler(telemetry *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodDelete {
			domain := r.URL.Query().Get("domain")
			fmt.Printf("[API-DELETE] Request to purge domain: %s\n", domain)
			if domain == "" || domain == "all" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Invalid domain for deletion"})
				return
			}
			telemetry.DeleteDomain(domain)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Domain purged from matrix"})
			return
		}

		// GET logic: Returns only active domains from the managed list
		domains := telemetry.GetDomains()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(domains)
	}
}

func aiEventsHandler(telemetry *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		json.NewEncoder(w).Encode(telemetry.GetRecentAIEvents())
	}
}

func aiStreamHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		flusher, _ := w.(http.Flusher)
		fmt.Fprintf(w, "data: {\"status\":\"TUNNEL_ACTIVE\"}\n\n")
		flusher.Flush()
		for {
			select {
			case <-r.Context().Done():
				return
			case <-time.After(15 * time.Second):
				fmt.Fprintf(w, ": heartbeat\n\n")
				flusher.Flush()
			}
		}
	}
}

func threatsStreamHandler(np *proxy.NexusProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		flusher, _ := w.(http.Flusher)
		clientID := fmt.Sprintf("CLIENT_%d", time.Now().UnixNano())
		clientChan := make(chan string, 100)
		np.ThreatListeners.Store(clientID, clientChan)
		defer np.ThreatListeners.Delete(clientID)
		fmt.Fprintf(w, "data: {\"status\":\"TUNNEL_ESTABLISHED\"}\n\n")
		flusher.Flush()
		for {
			select {
			case <-r.Context().Done():
				return
			case msg := <-clientChan:
				fmt.Fprintf(w, "data: %s\n\n", msg)
				flusher.Flush()
			case <-time.After(15 * time.Second):
				fmt.Fprintf(w, ": heartbeat\n\n")
				flusher.Flush()
			}
		}
	}
}

func aiStatusHandler() http.HandlerFunc {
	client := ai.NewQwenClient("")
	return func(w http.ResponseWriter, r *http.Request) {
		status, latency := client.CheckHealth()
		json.NewEncoder(w).Encode(map[string]interface{}{"status": status, "latency_ms": latency, "model": "QWEN3-32B"})
	}
}

func routesHandler(router *proxy.DynamicRouter, telemetry *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			routes, _ := router.GetAllRoutes()
			json.NewEncoder(w).Encode(routes)
			return
		}

		if r.Method == http.MethodPost {
			var payload struct {
				Domain    string `json:"domain"`
				TargetURL string `json:"target_url"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// 1. Add to Proxy Router
			router.AddRoute(payload.Domain, payload.TargetURL)
			
			// 2. Register in Telemetry so it appears in the dropdown immediately
			telemetry.AddDomain(payload.Domain)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "success", "domain": payload.Domain})
			return
		}
	}
}

func panicHandler(shuffler *mtd.TopologyShuffler, telemetry *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		telemetry.TotalPanic++
		shuffler.ManualShuffle()
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}
}

func nechatHandler(telemetry *logger.Logger) http.HandlerFunc {
	nechat := ai.NewNechatClient()
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Query  string `json:"query"`
			Domain string `json:"domain"`
		}
		json.NewDecoder(r.Body).Decode(&payload)
		reply, err := nechat.Chat(telemetry.GetRecentLogs(), payload.Query)
		if err != nil {
			fmt.Printf("[ALPACA-ERROR] Nechat failed: %v\n", err)
			reply = "🤖 **Nexus Core Error:** Gagal terhubung ke sistem ALPACA. \n\n**Solusi:**\n1. Pastikan aplikasi Alpaca sedang terbuka.\n2. Cek apakah model `llama3` sudah di-download di dalam Alpaca.\n3. Coba restart gateway."
		}
		json.NewEncoder(w).Encode(map[string]string{"reply": reply})
	}
}

func cliExecuteHandler(telemetry *logger.Logger, shuffler *mtd.TopologyShuffler, router *proxy.DynamicRouter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Command string `json:"command"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		cmd := strings.ToLower(payload.Command)
		var response string

		switch {
		case cmd == "help" || cmd == "/help":
			response = "[NEXUS-HELP] Available Commands:\n" +
				"  - status      : Check MTD & Backend Health\n" +
				"  - stats       : Show global traffic metrics\n" +
				"  - shuffle     : Trigger manual topology rotation\n" +
				"  - @nexus [q]  : Ask AI about current threats\n" +
				"  - clear       : Clear terminal session"

		case cmd == "status":
			port, next := shuffler.GetStatus()
			response = fmt.Sprintf("[STATUS] MTD Active Port: %d | Next Shuffle: %ds | Backend: ONLINE", port, next)

		case cmd == "stats":
			response = fmt.Sprintf("[STATS] Allowed: %d | Blocked: %d | Honeypot: %d", 
				telemetry.TotalAllowed, telemetry.TotalBlocked, telemetry.TotalHoneypot)

		case cmd == "shuffle":
			shuffler.ManualShuffle()
			response = "[ACTION] Manual Topology Rotation Triggered. New port mapping established."

		case strings.HasPrefix(cmd, "@nexus"):
			query := strings.TrimPrefix(payload.Command, "@nexus ")
			// Use the AI client to analyze the situation based on recent logs
			nechat := ai.NewNechatClient()
			reply, err := nechat.Chat(telemetry.GetRecentLogs(), query)
			if err != nil {
				fmt.Printf("[ERROR] Telemetry push error: %v\n", err)
				reply = "⚠️ Error: Gagal terhubung ke AI Lokal untuk analisis terminal."
			}
			json.NewEncoder(w).Encode(map[string]string{"output": "[NEXUS-AI] Analysis:\n" + reply})
			return

		default:
			response = fmt.Sprintf("[ERROR] Unknown command: '%s'. Type /help for assistance.", cmd)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"response": response})
	}
}
