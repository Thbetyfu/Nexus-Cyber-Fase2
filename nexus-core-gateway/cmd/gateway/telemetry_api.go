package main

import (
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/nexus-cyber/nexus-core-gateway/internal/ai"
	"github.com/nexus-cyber/nexus-core-gateway/internal/mtd"
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

// telemetryHandler serves live MTD and traffic stats to the internal dashboard.
func telemetryHandler(shuffler *mtd.TopologyShuffler, telemetry *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// SECURITY: Strict Localhost Internal Access Only
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		if ip != "127.0.0.1" && ip != "::1" && ip != "localhost" {
			http.Error(w, `{"error":"Forbidden. Dashboard API is internal only."}`, http.StatusForbidden)
			return
		}

		// CORS for local Next.js dev server on port 3000
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		port, remain := shuffler.GetStatus()
		logs := telemetry.GetRecentLogs()

		resp := TelemetryResponse{}
		resp.MTD.ActivePort = port
		resp.MTD.NextShuffle = remain
		resp.MTD.Status = "ACTIVE"
		resp.RecentLogs = logs
		resp.Stats.Allowed = telemetry.TotalAllowed
		resp.Stats.Blocked = telemetry.TotalBlocked
		resp.Stats.Honeypot = telemetry.TotalHoneypot
		resp.Stats.Panics = telemetry.TotalPanic

		json.NewEncoder(w).Encode(resp)
	}
}

// panicHandler triggers an Emergency Rescue Protocol (MTD Shuffle)
func panicHandler(shuffler *mtd.TopologyShuffler, telemetry *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

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
		shuffler.ManualShuffle()

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
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
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
			Query string `json:"query"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		// RAG: Inject memory
		logs := telemetry.GetRecentLogs()

		// Call AI
		reply, err := nechat.Chat(logs, payload.Query)
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
