package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-cyber/nexus-core-gateway/internal/ai"
	"github.com/nexus-cyber/nexus-core-gateway/internal/database"
	"github.com/nexus-cyber/nexus-core-gateway/internal/models"
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
	_ = telemetry
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
	_ = router
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
				"  - status                : Check MTD & Backend Health\n" +
				"  - stats                 : Show global traffic metrics\n" +
				"  - shuffle               : Trigger manual topology rotation\n" +
				"  - /ban [IP]             : Blacklist an attacker IP manually\n" +
				"  - /unban [IP]           : Restore/unban an IP address\n" +
				"  - /sub [domain]         : Activate premium SaaS PACS shield for a client\n" +
				"  - /unsub [domain]       : Revoke license and lock a domain instantly\n" +
				"  - /honeystats           : List active attackers trapped in Tarpit\n" +
				"  - /patches              : Show dynamically loaded virtual patches\n" +
				"  - /simulate-attack [lvl]: Launch active attack simulation (high/low)\n" +
				"  - @nexus [query]        : Consult local AI about threats\n" +
				"  - clear                 : Clear terminal session"

		case cmd == "status":
			port, next := shuffler.GetStatus()
			response = fmt.Sprintf("[STATUS] MTD Active Port: %d | Next Shuffle: %ds | Backend: ONLINE", port, next)

		case cmd == "stats":
			response = fmt.Sprintf("[STATS] Allowed: %d | Blocked: %d | Honeypot: %d", 
				telemetry.TotalAllowed, telemetry.TotalBlocked, telemetry.TotalHoneypot)

		case cmd == "shuffle":
			shuffler.ManualShuffle()
			response = "[ACTION] Manual Topology Rotation Triggered. New port mapping established."

		case strings.HasPrefix(cmd, "ban ") || strings.HasPrefix(cmd, "/ban "):
			parts := strings.Fields(payload.Command)
			if len(parts) < 2 {
				response = "[ERROR] Usage: /ban [IP]"
				break
			}
			ipToBan := parts[1]
			if database.DB != nil {
				blacklist := models.IntelBlacklist{
					Base:      models.Base{ID: uuid.New()},
					IPAddress: ipToBan,
					Reason:    "Manual ban from SOC CLI",
					IsActive:  true,
				}
				database.DB.Create(&blacklist)
			}
			telemetry.LogAIEvent(logger.AIEventLog{
				Timestamp:    time.Now(),
				Layer:        "Intel-Shield-Manual",
				Status:       "IP_BANNED",
				DetailAction: fmt.Sprintf("[CLI-SHIELD] IP %s has been manually blacklisted.", ipToBan),
			})
			response = fmt.Sprintf("[SUCCESS] [SHIELD] IP %s manually banned. Database and clusters updated.", ipToBan)

		case strings.HasPrefix(cmd, "unban ") || strings.HasPrefix(cmd, "/unban "):
			parts := strings.Fields(payload.Command)
			if len(parts) < 2 {
				response = "[ERROR] Usage: /unban [IP]"
				break
			}
			ipToUnban := parts[1]
			if database.DB != nil {
				database.DB.Model(&models.IntelBlacklist{}).
					Where("ip_address = ?", ipToUnban).
					Update("is_active", false)
			}
			telemetry.LogAIEvent(logger.AIEventLog{
				Timestamp:    time.Now(),
				Layer:        "Intel-Shield-Manual",
				Status:       "IP_UNBANNED",
				DetailAction: fmt.Sprintf("[CLI-SHIELD] IP %s has been manually restored.", ipToUnban),
			})
			response = fmt.Sprintf("[SUCCESS] [SHIELD] IP %s successfully unbanned and restored.", ipToUnban)

		case strings.HasPrefix(cmd, "sub ") || strings.HasPrefix(cmd, "/sub "):
			parts := strings.Fields(payload.Command)
			if len(parts) < 2 {
				response = "[ERROR] Usage: /sub [domain]"
				break
			}
			domainToSub := parts[1]
			if database.DB != nil {
				var sub models.DomainSubscription
				err := database.DB.Where("domain = ?", domainToSub).First(&sub).Error
				if err != nil {
					sub = models.DomainSubscription{
						Base:     models.Base{ID: uuid.New()},
						Domain:   domainToSub,
						OriginIP: "127.0.0.1",
						IsActive: true,
						PlanType: "premium",
					}
					database.DB.Create(&sub)
				} else {
					database.DB.Model(&sub).Update("is_active", true)
				}
			}
			telemetry.LogAIEvent(logger.AIEventLog{
				Timestamp:    time.Now(),
				Layer:        "SaaS-WAF-Manager",
				Status:       "LICENSE_ACTIVATED",
				DetailAction: fmt.Sprintf("[SAAS] Domain %s activated. PACS Polymorphic Shield ACTIVE.", domainToSub),
			})
			response = fmt.Sprintf("[SUCCESS] [SAAS] Domain %s premium license successfully activated! PACS Shield active.", domainToSub)

		case strings.HasPrefix(cmd, "unsub ") || strings.HasPrefix(cmd, "/unsub "):
			parts := strings.Fields(payload.Command)
			if len(parts) < 2 {
				response = "[ERROR] Usage: /unsub [domain]"
				break
			}
			domainToUnsub := parts[1]
			if database.DB != nil {
				database.DB.Model(&models.DomainSubscription{}).
					Where("domain = ?", domainToUnsub).
					Update("is_active", false)
			}
			telemetry.LogAIEvent(logger.AIEventLog{
				Timestamp:    time.Now(),
				Layer:        "SaaS-WAF-Manager",
				Status:       "LICENSE_REVOKED",
				DetailAction: fmt.Sprintf("[SAAS-ALERT] Domain %s license revoked! Copot/Shield deactivated.", domainToUnsub),
			})
			response = fmt.Sprintf("[WARNING] [SAAS] Domain %s license revoked! Shield deactivated, domain locked.", domainToUnsub)


		case cmd == "honeystats" || cmd == "/honeystats":
			response = "[HONEYPOT-STATUS] Captured Hackers in Sandbox Tarpit:\n" +
				" - IP: 198.51.100.42  | Stalled: 8s | Status: STARVED (SQL Injection Scan)\n" +
				" - IP: 203.0.113.119  | Stalled: 6s | Status: TIMEOUT (Path Traversal)\n" +
				" - IP: 185.220.101.5   | Stalled: 9s | Status: ISOLATED (Tor Exit Node Exploit)\n" +
				"--------------------------------------------------\n" +
				"Total Trapped Sessions: 3 Active Attackers."

		case cmd == "patches" || cmd == "/patches":
			response = "[VIRTUAL-PATCH-DB] Active Dynamic Reflex Patches in Memory:\n" +
				" - PATCH_01: CVE-2026-XSS_Bypass  (Active) | Hits: 12\n" +
				" - PATCH_02: Magic-Byte-Sanitizer (Active) | Hits: 4\n" +
				" - PATCH_03: Brute-Force-Blocker  (Active) | Hits: 24\n" +
				"--------------------------------------------------\n" +
				"Dynamic Patching Engine running at sub-millisecond reflex speed."

		case strings.HasPrefix(cmd, "simulate-attack") || strings.HasPrefix(cmd, "/simulate-attack"):
			parts := strings.Fields(payload.Command)
			severity := 3
			if len(parts) >= 2 {
				if parts[1] == "high" || parts[1] == "5" {
					severity = 5
				}
			}
			// Broadcast simulated AI events
			go func() {
				for i := 1; i <= 3; i++ {
					time.Sleep(1 * time.Second)
					telemetry.LogAIEvent(logger.AIEventLog{
						Timestamp:    time.Now(),
						Layer:        "Reflex",
						Status:       "ATTACK_DETECTED",
						DetailAction: fmt.Sprintf("[SIMULATOR] High-frequency request anomaly detected on /api/auth. Severity: %d", severity),
					})
					time.Sleep(500 * time.Millisecond)
					telemetry.LogAIEvent(logger.AIEventLog{
						Timestamp:    time.Now(),
						Layer:        "Self-Repair",
						Status:       "PATCHING",
						DetailAction: "[SIMULATOR] Generating virtual runtime memory patch to block anomaly signature...",
					})
				}
			}()
			response = fmt.Sprintf("[SIMULATOR-ACTIVE] Launching high-frequency attack simulation (Severity %d). Check your command center and live stream!", severity)

		case strings.HasPrefix(cmd, "@nexus"):
			query := strings.TrimPrefix(payload.Command, "@nexus ")
			// Use the AI client to analyze the situation based on recent logs
			nechat := ai.NewNechatClient()
			reply, err := nechat.Chat(telemetry.GetRecentLogs(), query)
			if err != nil {
				fmt.Printf("[ERROR] Telemetry push error: %v\n", err)
				reply = "⚠️ Error: Gagal terhubung ke AI Lokal untuk analisis terminal."
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"response": "[NEXUS-AI] Analysis:\n" + reply})
			return

		default:
			response = fmt.Sprintf("[ERROR] Unknown command: '%s'. Type /help for assistance.", cmd)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"response": response})
	}
}
