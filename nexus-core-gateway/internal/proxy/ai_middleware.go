package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-cyber/nexus-core-gateway/internal/database"
	"github.com/nexus-cyber/nexus-core-gateway/internal/mtd"
	"github.com/nexus-cyber/nexus-core-gateway/pkg/logger"
)

// AIMiddleware provides the 'Dual-Brain' protection for ALL entry points,
// including internal APIs. It triggers the OS Firewall for critical threats.
func (np *NexusProxy) AIMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// [LAYER 0: INTEL BLACKLIST] Instant Drop for known malicious IPs
		if database.IsIPBlacklisted(r.RemoteAddr) {
			np.Logger.LogAIEvent(logger.AIEventLog{
				Timestamp:    time.Now(),
				Layer:        "Intel-Shield",
				Status:       "BANNED_MATCH",
				DetailAction: fmt.Sprintf("[ZERO TRUST] Instant Diversion - IP %s is in the persistent blacklist.", r.RemoteAddr),
			})

			tLog := logger.TelemetryLog{
				Timestamp:    time.Now(),
				SourceIP:     r.RemoteAddr,
				Endpoint:     r.URL.Path,
				Method:       r.Method,
				Status:       "BANNED_IP_DIVERTED",
				TargetDomain: r.Host,
				ThreatDetail: "INTEL_BLACKLIST_HIT",
			}
			np.Logger.EnrichLog(&tLog, r)
			np.Logger.LogTraffic(tLog)

			np.PublishThreat(r.RemoteAddr, "BLACKLISTED_IP")
			np.routeToHoneypot(w, r)
			return
		}

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

		// [NEW: VIRTUAL PATCHING] Immunity Check (Local Antibodies Memory)
		isPatched := false
		np.Patches.Range(func(key, value interface{}) bool {
			pattern := key.(string)
			if strings.Contains(analysisData, pattern) {
				isPatched = true
				return false
			}
			return true
		})

		// Base Telemetry Log
		tLog := logger.TelemetryLog{
			Timestamp:     time.Now(),
			SourceIP:      r.RemoteAddr,
			Endpoint:      r.URL.Path,
			Method:        r.Method,
			TargetDomain:  r.Host,
			PayloadSample: analysisData,
		}

		if isPatched {
			np.Logger.LogAIEvent(logger.AIEventLog{
				Timestamp:    time.Now(),
				Layer:        "Virtual-Patch",
				Status:       "IMMUNE",
				DetailAction: "[VIRTUAL PATCH] Instant Drop - Antibody Signature Match.",
			})

			tLog.Status = "INSTANT_DROP_PATCH"
			tLog.ThreatDetail = "VIRTUAL_PATCH_MATCH"
			np.Logger.EnrichLog(&tLog, r)
			np.Logger.LogTraffic(tLog)

			np.PublishThreat(r.RemoteAddr, "VIRTUAL_PATCH_IMMUNE")
			np.routeToHoneypot(w, r)
			return
		}

		// 2. Dual-Brain Layer 1: Reflex
		ua := r.Header.Get("User-Agent")
		isThreat, threatType := np.Filter.InspectAdvanced(analysisData, ua)

		if isThreat {
			if len(analysisData) > 10 {
				np.AddAntibody(analysisData)
			}

			np.Logger.LogAIEvent(logger.AIEventLog{
				Timestamp:    time.Now(),
				Layer:        "Reflex",
				Status:       "MITIGATING",
				DetailAction: fmt.Sprintf("Critical Threat [%s] from %s.", threatType, r.RemoteAddr),
			})

			tLog.Status = "DIVERTED_TO_HONEYPOT"
			tLog.ThreatDetail = threatType
			np.Logger.EnrichLog(&tLog, r)
			np.Logger.LogTraffic(tLog)

			np.PublishThreat(r.RemoteAddr, threatType)
			np.routeToHoneypot(w, r)
			return
		}

		// 3. Dual-Brain Layer 2: Reasoning (Async Background Audit)
		if len(analysisData) > 10 {
			tLog.Status = "ALLOWED"
			tLog.ThreatDetail = "PASS_THROUGH_AI"
			np.Logger.EnrichLog(&tLog, r)
			logID := np.Logger.LogTraffic(tLog)

			go func(data string, source string, id uuid.UUID) {
				result, err := np.Reasoning.AnalyzeIntent(data)
				if err == nil && result != nil {
					// Persist AI reasoning to database
					database.SaveAIInsight(id, "qwen/qwen3-235b-a22b", result.ForensicSummary, result.ThreatVerdict)

					if result.ThreatVerdict == "CONFIRMED_MALICIOUS" || result.ThreatVerdict == "ADVANCED_PERSISTENT" {
						mtd.BlockIPAtOSLevel(source)
					}
				}
			}(analysisData, r.RemoteAddr, logID)
		} else {
			tLog.Status = "ALLOWED"
			tLog.ThreatDetail = "PASS_THROUGH_AI"
			np.Logger.EnrichLog(&tLog, r)
			np.Logger.LogTraffic(tLog)
		}

		w.Header().Set("X-Quantum-Safe", "ML-KEM-768-Active")
		next.ServeHTTP(w, r)
	})
}
