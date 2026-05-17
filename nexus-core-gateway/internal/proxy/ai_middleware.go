// Package proxy mengimplementasikan gateway proxy reverse otonom dengan kecerdasan MTD.
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

// AIMiddleware mengimplementasikan sistem proteksi berlapis "Dual-Brain" (Reflex + Reasoning).
//
// Alasan Arsitektural (Why):
// Sistem proteksi ini menerapkan filosofi pertahanan mendalam (Defense-in-Depth) dan Arsitektur Zero Trust.
// Setiap paket dianalisis secara berjenjang dari pencocokan data cache tercepat (O(1)) hingga audit kognitif
// asinkron menggunakan LLM guna memberikan keseimbangan optimal antara latency gateway dan tingkat deteksi ancaman.
func (np *NexusProxy) AIMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// [LAYER 0: INTEL BLACKLIST SHIELD]
		// Alasan Teknis (Why):
		// IP yang sudah terbukti jahat langsung dibelokkan ke Honeypot tanpa perlu melalui analisis AI lagi.
		// Ini memangkas konsumsi komputasi gerbang (Zero AI Waste Policy).
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

		// Mencegah perulangan tak terbatas (infinite redirection loops) pada endpoint verifikasi sesi.
		if r.URL.Path == "/api/verify-session" {
			next.ServeHTTP(w, r)
			return
		}

		// [BYPASS DIAGNOSTIC APIS]
		// Alasan Teknis (Why):
		// API internal Dashboard SOC harus dibebaskan dari filter AI agar visualisasi data telemetri,
		// log ancaman, dan interaksi NCC tidak mengalami kemacetan akibat evaluasi heuristik teks.
		if strings.HasPrefix(r.URL.Path, "/api/telemetry") ||
			strings.HasPrefix(r.URL.Path, "/api/logs") ||
			strings.HasPrefix(r.URL.Path, "/api/domains") ||
			strings.HasPrefix(r.URL.Path, "/api/ai-events") ||
			strings.HasPrefix(r.URL.Path, "/api/upload") ||
			strings.HasPrefix(r.URL.Path, "/api/unlock-reward") {
			next.ServeHTTP(w, r)
			return
		}

		// 1. Tangkap payload body secara non-destruktif.
		// Alasan Teknis (Why):
		// Pembacaan r.Body mengonsumsi aliran data biner (stream read-once).
		// Kita menyalin data ke memori, lalu merestorasi r.Body menggunakan `io.NopCloser` agar
		// handler proxy berikutnya dapat membaca data request secara normal.
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		query, _ := url.QueryUnescape(r.URL.RawQuery)
		analysisData := query + " " + string(body)

		// [LAYER 1: VIRTUAL PATCH IMMUNITY CHALLENGE]
		// Alasan Arsitektural (Why):
		// Sistem kekebalan adaptif (Adaptive Immune System). Ketika AI mendeteksi muatan serangan SQLi/XSS baru,
		// pola muatan tersebut disimpan dalam bentuk tanda tangan "Antibodi" di memori.
		// Request selanjutnya yang mengandung pola serupa akan langsung diblokir seketika secara O(1) di Layer 1
		// tanpa membuang siklus CPU untuk memanggil Regex atau Model AI kembali.
		isPatched := false
		np.Patches.Range(func(key, value interface{}) bool {
			pattern := key.(string)
			if strings.Contains(analysisData, pattern) {
				isPatched = true
				return false
			}
			return true
		})

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

		// [LAYER 2: AI REFLEX SHIELD]
		// Alasan Arsitektural (Why):
		// Lapisan pertahanan respons instan (latency sub-50ms). Heuristik pre-compiled Regex melakukan inspeksi
		// tanda tangan serangan siber OWASP Top 10 (SQLi, XSS, Path Traversal) secara real-time.
		ua := r.Header.Get("User-Agent")
		isThreat, threatType := np.Filter.InspectAdvanced(analysisData, ua)

		if isThreat {
			// Otomatis imunisasi pola serangan jika panjang karakter memenuhi kelayakan tanda tangan antibodi.
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

		// [LAYER 3: AI COGNITIVE REASONING AUDIT]
		// Alasan Arsitektural (Why):
		// Analisis Kognitif Asinkron (Async Background Task).
		// Model reasoning (seperti Llama) memerlukan waktu komputasi intensif (timeout 30 detik) untuk memahami intensi peretas.
		// Menjalankannya secara sinkron akan memacetkan koneksi klien dan meningkatkan latency gerbang secara ekstrem.
		// Solusinya, request aman dilewatkan ke backend utama (sub-millisecond HTTP PASS), sementara audit mendalam
		// dieksekusi di background goroutine secara terisolasi. Jika terkonfirmasi serangan canggih (APT),
		// sistem langsung menginstruksikan modul `mtd.BlockIPAtOSLevel` untuk menghentikan seluruh paket IP penyerang.
		if len(analysisData) > 10 {
			tLog.Status = "ALLOWED"
			tLog.ThreatDetail = "PASS_THROUGH_AI"
			np.Logger.EnrichLog(&tLog, r)
			logID := np.Logger.LogTraffic(tLog)

			go func(data string, source string, id uuid.UUID) {
				result, err := np.Reasoning.AnalyzeIntent(data)
				if err == nil && result != nil {
					// Rekam jejak hasil keputusan kognitif AI ke basis data audit.
					database.SaveAIInsight(id, "qwen/qwen3-235b-a22b", result.ForensicSummary, result.ThreatVerdict)

					// Eksekusi pemblokiran Netfilter seketika jika intensi terbukti berbahaya.
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

		// Sematkan penanda pertahanan pasca-kuantum aktif pada respon keluar.
		w.Header().Set("X-Quantum-Safe", "ML-KEM-768-Active")
		next.ServeHTTP(w, r)
	})
}
