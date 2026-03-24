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

		domains := telemetry.GetDomains()
		// [NEXUS_V12_ASSET_SHIELD]: Force inclusion of critical national assets for total visibility
		criticalAssets := []string{
			"localhost",
			"ojk.go.id",
			"bi.go.id",
			"kemenkeu.go.id",
			"portal.nexus",
			"audit.nexus",
			"cloud.nexus",
		}

		uniqueDomains := make(map[string]bool)
		for _, d := range domains {
			uniqueDomains[d] = true
		}
		for _, a := range criticalAssets {
			uniqueDomains[a] = true
		}

		finalDomains := make([]string, 0, len(uniqueDomains))
		for d := range uniqueDomains {
			finalDomains = append(finalDomains, d)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(finalDomains)
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

func routesHandler(router *proxy.DynamicRouter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			routes, _ := router.GetAllRoutes()
			json.NewEncoder(w).Encode(routes)
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
		reply, _ := nechat.Chat(telemetry.GetRecentLogs(), payload.Query)
		json.NewEncoder(w).Encode(map[string]string{"reply": reply})
	}
}

func cliExecuteHandler(telemetry *logger.Logger, shuffler *mtd.TopologyShuffler, router *proxy.DynamicRouter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"response": "CLI Layer Active"})
	}
}
