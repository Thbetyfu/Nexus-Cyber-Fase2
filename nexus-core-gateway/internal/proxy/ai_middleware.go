package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nexus-cyber/nexus-core-gateway/internal/mtd"
	"github.com/nexus-cyber/nexus-core-gateway/pkg/logger"
)

// AIMiddleware provides the 'Dual-Brain' protection for ALL entry points,
// including internal APIs. It triggers the OS Firewall for critical threats.
func (np *NexusProxy) AIMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip for Honeypot itself to avoid loops
		if r.URL.Path == "/api/verify-session" {
			next.ServeHTTP(w, r)
			return
		}

		// Safety: Allow SOC Dashboard Diagnostic APIs without AI filtering
		if strings.HasPrefix(r.URL.Path, "/api/telemetry") ||
			strings.HasPrefix(r.URL.Path, "/api/logs") ||
			strings.HasPrefix(r.URL.Path, "/api/domains") ||
			strings.HasPrefix(r.URL.Path, "/api/ai-events") {
			next.ServeHTTP(w, r)
			return
		}

		// 1. Capture Payload
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(body)) // Reset for next handler

		query, _ := url.QueryUnescape(r.URL.RawQuery)
		analysisData := query + " " + string(body)

		if len(analysisData) > 5 {
			fmt.Printf("[AI-INSPECT] Path: %s | Payload: %s\n", r.URL.Path, analysisData)
		}

		// [NEW: VIRTUAL PATCHING] Layer 0: Immunity Check (Local Antibodies Memory)
		// O(1) time complexity — Instant Drop for already discovered malicious patterns.
		isPatched := false
		np.Patches.Range(func(key, value interface{}) bool {
			pattern := key.(string)
			if strings.Contains(analysisData, pattern) {
				isPatched = true
				return false // stop iteration
			}
			return true
		})

		if isPatched {
			// LOG VIRTUAL PATCH HIT
			np.Logger.LogAIEvent(logger.AIEventLog{
				Timestamp:    time.Now(),
				Layer:        "Virtual-Patch",
				Status:       "IMMUNE",
				DetailAction: "[VIRTUAL PATCH] Instant Drop - Antibody Signature Match.",
			})

			// LOG TRAFFIC INCIDENT
			tLog := logger.TelemetryLog{
				Timestamp:    time.Now(),
				SourceIP:     r.RemoteAddr,
				Endpoint:     r.URL.Path,
				Method:       r.Method,
				Status:       "INSTANT_DROP_PATCH",
				TargetDomain: r.Host,
				ThreatDetail: "VIRTUAL_PATCH_MATCH",
			}
			np.Logger.LogTraffic(tLog)

			// MTD: Digital Hallucination (Honeypot Redirect)
			np.routeToHoneypot(w, r)
			return
		}

		// 2. Dual-Brain Layer 1: Reflex
		isThreat, threatType := np.Filter.InspectRequest(analysisData)

		if isThreat {
			// [NEW: VIRTUAL PATCHING] Generate new antibody for this pattern
			if len(analysisData) > 10 {
				np.AddAntibody(analysisData)
				np.Logger.LogAIEvent(logger.AIEventLog{
					Timestamp:    time.Now(),
					Layer:        "Virtual-Patch",
					Status:       "ANTIBODY_GEN",
					DetailAction: "[VIRTUAL PATCH] New signature generated and cached for future immunity.",
				})
			}

			// LOG SECURITY INCIDENT
			np.Logger.LogAIEvent(logger.AIEventLog{
				Timestamp:    time.Now(),
				Layer:        "Reflex",
				Status:       "MITIGATING",
				DetailAction: fmt.Sprintf("Critical Threat [%s] from %s. Triggering OS-Firewall...", threatType, r.RemoteAddr),
			})

			// LOG TRAFFIC INCIDENT
			tLog := logger.TelemetryLog{
				Timestamp:    time.Now(),
				SourceIP:     r.RemoteAddr,
				Endpoint:     r.URL.Path,
				Method:       r.Method,
				Status:       "DIVERTED_TO_HONEYPOT",
				TargetDomain: r.Host,
				ThreatDetail: threatType,
			}
			np.Logger.LogTraffic(tLog)

			// TRIGGER OS FIREWALL (iptables) - Skip Dev Loopback for Safety
			if !strings.HasPrefix(r.RemoteAddr, "127.0.0.1") && !strings.HasPrefix(r.RemoteAddr, "[::1]") {
				if strings.Contains(threatType, "SQL_INJECTION") || strings.Contains(threatType, "DETECTED") {
					go mtd.BlockIPAtOSLevel(r.RemoteAddr)
				}
			} else {
				fmt.Printf("[ALER] Threat detected from LOCALHOST (%s). Kernel bypass enabled for Safety.\n", r.RemoteAddr)
			}

			// MTD: Digital Hallucination (Honeypot Redirect)
			np.routeToHoneypot(w, r)
			return
		}

		// 3. Dual-Brain Layer 2: Reasoning (Async Background Audit)
		// We pass high-risk requests to the deep cognitive engine without blocking.
		if len(analysisData) > 10 {
			go func(data string, source string) {
				isMalicious, _ := np.Reasoning.AnalyzeIntent(data)
				if isMalicious {
					mtd.BlockIPAtOSLevel(source)
				}
			}(analysisData, r.RemoteAddr)
		}

		next.ServeHTTP(w, r)
	})
}
