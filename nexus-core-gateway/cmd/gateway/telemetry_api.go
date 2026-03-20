package main

import (
	"encoding/json"
	"net"
	"net/http"

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
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Content-Type", "application/json")

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

		json.NewEncoder(w).Encode(resp)
	}
}
