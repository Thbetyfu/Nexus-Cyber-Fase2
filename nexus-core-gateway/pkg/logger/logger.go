// Package logger mengimplementasikan sistem pencatatan telemetri trafik dan aktivitas kognitif AI secara real-time.
package logger

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mssola/user_agent"
	"github.com/nexus-cyber/nexus-core-gateway/internal/database"
	"github.com/nexus-cyber/nexus-core-gateway/internal/models"
)

// TelemetryLog merepresentasikan metadata trafik yang dialirkan ke dasbor visualisasi Nexus.
type TelemetryLog struct {
	Timestamp         time.Time `json:"timestamp"`
	SourceIP          string    `json:"source_ip"`
	AttackerID        string    `json:"attacker_id,omitempty"` // Identifikasi unik peretas
	GeoLocation       string    `json:"geo_location,omitempty"`
	ISP               string    `json:"isp,omitempty"`
	DeviceFingerprint string    `json:"device_fingerprint,omitempty"`
	Endpoint          string    `json:"endpoint"`
	Method            string    `json:"method"`
	Status            string    `json:"status"` // Status eksekusi: ALLOWED, BLOCKED, HONEYPOT_REDIRECTED, dll.
	ThreatDetail      string    `json:"threat_detail,omitempty"`
	TargetDomain      string    `json:"target_domain"` // Domain target untuk pelacakan Multi-Tenant
	LatencyMS         int64     `json:"latency_ms"`
	PayloadSample     string    `json:"payload_sample,omitempty"`
}

// AIEventLog merepresentasikan aktivitas internal asisten AI kognitif dan tindakan self-repair.
type AIEventLog struct {
	Timestamp    time.Time `json:"timestamp"`
	Layer        string    `json:"layer"`         // Lapisan keamanan: "Reflex", "Reasoning", "Self-Repair"
	Status       string    `json:"status"`        // Status: "Analyzing", "Mitigating", "Repairing"
	DetailAction string    `json:"detail_action"` // Penjelasan rinci tindakan otonom
}

// Logger mengelola pencatatan data trafik ke penyimpanan lokal, basis data, dan memori untuk dasbor.
//
// Alasan Arsitektural (Why):
// Sistem pencatatan ini dirancang berdasarkan standar ISO 25010 (Time-Behavior & Efficiency) dan ISO 27001 (Auditing).
// Untuk mencegah perlambatan eksekusi rute akibat operasi pencatatan, logger ini menerapkan:
// - Caching Fingerprint: Menghindari parsing berulang terhadap IP dan User-Agent yang sama.
// - Asynchronous DB Write: Penyimpanan audit ke basis data PostgreSQL dilakukan melalui goroutine latar belakang terpisah.
type Logger struct {
	file           *os.File
	aiFile         *os.File
	mu             sync.RWMutex
	recentLogs     []TelemetryLog
	recentAIEvents []AIEventLog
	OnAIEvent      func(AIEventLog)

	// Penghitung akumulatif global untuk visualisasi dasbor utama
	TotalAllowed  int
	TotalBlocked  int
	TotalHoneypot int
	TotalPanic    int

	// Statistik per domain untuk arsitektur multi-tenant
	DomainStats map[string]*DomainStatsEntry

	// Cache profil perantara untuk efisiensi komputasi ekstrem
	fingerprintCache map[string]TelemetryLog
}

// NewLogger mengonstruksi sistem pencatatan baru dan mempratata domain aset nasional kritis.
func NewLogger() (*Logger, error) {
	f, err := os.OpenFile("nexus_traffic.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	aiF, err := os.OpenFile("nexus_ai_events.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	l := &Logger{
		file:             f,
		aiFile:           aiF,
		recentLogs:       make([]TelemetryLog, 0),
		recentAIEvents:   make([]AIEventLog, 0),
		fingerprintCache: make(map[string]TelemetryLog),
		DomainStats:      make(map[string]*DomainStatsEntry),
	}

	// Inisialisasi awal aset strategis nasional demi pelacakan terpusat instan.
	criticalAssets := []string{
		"localhost",
		"ojk.go.id",
		"bi.go.id",
		"kemenkeu.go.id",
		"portal.nexus",
		"audit.nexus",
		"cloud.nexus",
	}
	for _, a := range criticalAssets {
		l.DomainStats[a] = &DomainStatsEntry{}
	}

	return l, nil
}

// DomainStatsEntry menampung hitungan performa per domain.
type DomainStatsEntry struct {
	Allowed  int
	Honeypot int
	Blocked  int
}

// EnrichLog melengkapi log telemetri dengan sidik jari digital (fingerprinting) dan deteksi perangkat.
//
// Alasan Arsitektural (Why):
// 1. Menerapkan pengujian cache (`fingerprintCache`) di tingkat awal untuk menghindari kalkulasi SHA-256 yang mahal
//    dan parsing string User-Agent yang berat secara berulang untuk paket-paket dari sumber yang sama.
// 2. Menggunakan algoritma SHA-256 dari kombinasi IP + User-Agent klien untuk menghasilkan pengenal unik
//    (`AttackerID`). Hal ini mempermudah pelacakan peretas meskipun mereka mencoba memalsukan identitas request.
func (l *Logger) EnrichLog(log *TelemetryLog, r *http.Request) {
	if r == nil {
		return
	}

	host := r.Host
	if host == "" {
		host = "all"
	}
	log.TargetDomain = host

	uaStr := r.Header.Get("User-Agent")
	cacheKey := log.SourceIP + uaStr

	// Uji keberadaan data profil klien di memori cache lokal
	l.mu.RLock()
	cache, exists := l.fingerprintCache[cacheKey]
	l.mu.RUnlock()

	if exists {
		log.AttackerID = cache.AttackerID
		log.GeoLocation = cache.GeoLocation
		log.ISP = cache.ISP
		log.DeviceFingerprint = cache.DeviceFingerprint
		return
	}

	// 1. SHA-256 Digital Fingerprinting
	h := sha256.New()
	h.Write([]byte(cacheKey))
	log.AttackerID = fmt.Sprintf("APT-ID-%X", h.Sum(nil)[:4])

	// 2. User-Agent Parsing
	ua := user_agent.New(uaStr)
	osInfo := ua.OS()
	browser, _ := ua.Browser()
	log.DeviceFingerprint = fmt.Sprintf("%s (%s)", osInfo, browser)

	// 3. Klasifikasi Zona Koneksi (Mock Dynamic GeoIP)
	isLocal := strings.HasPrefix(log.SourceIP, "127.") ||
		strings.HasPrefix(log.SourceIP, "::1") ||
		strings.HasPrefix(log.SourceIP, "[::1]") ||
		log.SourceIP == "localhost"

	if isLocal {
		log.GeoLocation = "Localhost, Nexus Gate"
		log.ISP = "Internal Loopback"
	} else {
		log.GeoLocation = "Global, Secure Zone"
		if strings.Contains(uaStr, "curl") || strings.Contains(uaStr, "python") {
			log.ISP = "Automated Bot/Scanner"
		} else {
			log.ISP = "Residential User"
		}
	}

	// Simpan ke cache lokal untuk request berikutnya
	l.mu.Lock()
	l.fingerprintCache[cacheKey] = *log
	l.mu.Unlock()
}

// LogTraffic mencatat data statistik trafik ke berkas log, memori RAM, dan database relasional.
//
// Alasan Arsitektural (Why):
// 1. Pembaruan metrik domain (`DomainStats`) dinormalisasi dengan membuang nomor port klien
//    agar data tercatat bersih berdasarkan nama domain host utama (misal: "localhost:8080" -> "localhost").
// 2. Penyimpanan ke database PostgreSQL dilakukan secara **asinkron** (`go func`). Operasi I/O jaringan ke DB
//    relasional bisa memakan waktu hingga puluhan milidetik. Memindahkannya ke goroutine latar belakang
//    memastikan latency gateway tidak terpengaruh oleh performa query basis data (ISO 25010 Efficiency).
// 3. Nilai biner payload dibersihkan dengan `strings.ToValidUTF8(..., "?")`. Perintah SQL INSERT PostgreSQL
//    akan memicu error fatal (SQLSTATE 22021) jika mendeteksi deretan byte biner non-UTF8 (misalnya muatan eksploitasi buffer overflow).
//    Pembersihan ini menjamin proses persistensi berjalan andal (Fault Tolerance).
func (l *Logger) LogTraffic(log TelemetryLog) uuid.UUID {
	logID := uuid.New()

	// Tulis baris JSON ke berkas log lokal (JSON Lines standard)
	data, _ := json.Marshal(log)
	l.file.WriteString(string(data) + "\n")

	l.mu.Lock()
	defer l.mu.Unlock()

	// 1. Perbarui Penghitung Global
	switch log.Status {
	case "ALLOWED":
		l.TotalAllowed++
	case "HONEYPOT_REDIRECTED", "DIVERTED_TO_HONEYPOT", "INSTANT_DROP_PATCH":
		l.TotalHoneypot++
	case "RATE_LIMITED", "BLOCKED":
		l.TotalBlocked++
	}

	// 2. Perbarui Penghitung Multi-Tenant (Domain Stats)
	if log.TargetDomain != "" {
		dom := strings.ToLower(log.TargetDomain)
		if idx := strings.Index(dom, ":"); idx != -1 {
			dom = dom[:idx]
		}

		if _, ok := l.DomainStats[dom]; !ok {
			l.DomainStats[dom] = &DomainStatsEntry{}
		}

		stats := l.DomainStats[dom]
		switch log.Status {
		case "ALLOWED":
			stats.Allowed++
		case "HONEYPOT_REDIRECTED", "DIVERTED_TO_HONEYPOT", "INSTANT_DROP_PATCH":
			stats.Honeypot++
		case "RATE_LIMITED", "BLOCKED":
			stats.Blocked++
		}
	}

	// 3. Batasi buffer memori (geser data lama) agar RAM tidak membengkak tanpa batas (OOM protection)
	l.recentLogs = append(l.recentLogs, log)
	if len(l.recentLogs) > 50 {
		l.recentLogs = l.recentLogs[len(l.recentLogs)-50:]
	}

	fmt.Printf("[NET-TRAFFIC] %s | %s | %s | %s -> %s | %dms\n",
		log.Timestamp.Format("15:04:05"), log.Status, log.TargetDomain, log.SourceIP, log.Endpoint, log.LatencyMS)

	// 4. Persistensi Audit Database Asinkron (ISO 27001)
	if database.DB != nil {
		go func(l TelemetryLog) {
			severity := 1
			if l.Status == "BLOCKED" || l.Status == "RATE_LIMITED" {
				severity = 3
			}
			if l.Status == "HONEYPOT_REDIRECTED" || l.Status == "DIVERTED_TO_HONEYPOT" {
				severity = 4
			}
			if strings.Contains(strings.ToUpper(l.ThreatDetail), "SQL") || strings.Contains(strings.ToUpper(l.ThreatDetail), "XSS") {
				severity = 5
			}

			// Bersihkan byte non-UTF8 untuk keamanan Postgres INSERT
			cleanPayload := strings.ToValidUTF8(l.PayloadSample, "?")

			dbLog := models.ThreatLog{
				SourceIP:      l.SourceIP,
				Endpoint:      l.Endpoint,
				Method:        l.Method,
				Status:        l.Status,
				ThreatType:    l.ThreatDetail,
				Severity:      severity,
				UserAgent:     l.DeviceFingerprint,
				LatencyMs:     int(l.LatencyMS),
				PayloadSample: cleanPayload,
			}
			dbLog.ID = logID

			if err := database.DB.Create(&dbLog).Error; err != nil {
				fmt.Printf("[DB-ERROR] Failed to save threat log: %v\n", err)
			}
		}(log)
	}

	return logID
}

// GetRecentLogs mengembalikan salinan log telemetri terbaru demi keamanan pembacaan bersama.
func (l *Logger) GetRecentLogs() []TelemetryLog {
	l.mu.RLock()
	defer l.mu.RUnlock()

	cpy := make([]TelemetryLog, len(l.recentLogs))
	copy(cpy, l.recentLogs)
	return cpy
}

// LogAIEvent mencatat aktivitas kognitif AI secara real-time.
func (l *Logger) LogAIEvent(event AIEventLog) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	data, _ := json.Marshal(event)
	l.mu.Lock()
	if l.aiFile != nil {
		l.aiFile.WriteString(string(data) + "\n")
	}

	l.recentAIEvents = append(l.recentAIEvents, event)
	if len(l.recentAIEvents) > 20 {
		l.recentAIEvents = l.recentAIEvents[len(l.recentAIEvents)-20:]
	}

	fmt.Printf("[AI-COGNITION] %s | Layer: %s | Status: %s | %s\n",
		event.Timestamp.Format("15:04:05"), event.Layer, event.Status, event.DetailAction)

	if l.OnAIEvent != nil {
		l.OnAIEvent(event)
	}
	l.mu.Unlock()
}

// GetRecentAIEvents mengembalikan daftar riwayat aktivitas kognitif AI terbaru.
func (l *Logger) GetRecentAIEvents() []AIEventLog {
	l.mu.RLock()
	defer l.mu.RUnlock()

	cpy := make([]AIEventLog, len(l.recentAIEvents))
	copy(cpy, l.recentAIEvents)
	return cpy
}

// Close menutup objek berkas logging secara aman sewaktu aplikasi berhenti.
func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
	}
	if l.aiFile != nil {
		l.aiFile.Close()
	}
}

// GetDomainStats mengambil data performa domain tertentu.
func (l *Logger) GetDomainStats(domain string) (Allowed, Blocked, Honeypot int) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if stats, ok := l.DomainStats[domain]; ok {
		return stats.Allowed, stats.Blocked, stats.Honeypot
	}
	return 0, 0, 0
}

// GetDomains mengembalikan seluruh nama domain terdaftar di dalam pelacakan logger.
func (l *Logger) GetDomains() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	domains := make([]string, 0, len(l.DomainStats))
	for d := range l.DomainStats {
		domains = append(domains, d)
	}
	return domains
}

// ResetAll membersihkan seluruh metrik sistem secara total (Cognitive Purge).
func (l *Logger) ResetAll() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.TotalAllowed = 0
	l.TotalBlocked = 0
	l.TotalHoneypot = 0
	l.TotalPanic = 0

	l.DomainStats = make(map[string]*DomainStatsEntry)

	l.recentLogs = make([]TelemetryLog, 0)
	l.recentAIEvents = make([]AIEventLog, 0)
	l.fingerprintCache = make(map[string]TelemetryLog)

	fmt.Println("[SYSTEM-RESET] Cognitive purge complete. All metrics zeroed.")
}

// DeleteDomain menghapus metrik dan catatan buffer memori domain tertentu dari sistem secara permanen.
func (l *Logger) DeleteDomain(domain string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	dom := strings.ToLower(domain)
	delete(l.DomainStats, dom)

	filteredLogs := make([]TelemetryLog, 0)
	for _, log := range l.recentLogs {
		if strings.ToLower(log.TargetDomain) != dom {
			filteredLogs = append(filteredLogs, log)
		}
	}
	l.recentLogs = filteredLogs

	fmt.Printf("[DOMAIN-DELETE] Domain '%s' and its metrics have been purged from memory.\n", dom)
}

// AddDomain mendaftarkan domain baru secara manual ke dalam pelacakan telemetri dasbor.
func (l *Logger) AddDomain(domain string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	dom := strings.ToLower(domain)
	if _, ok := l.DomainStats[dom]; !ok {
		l.DomainStats[dom] = &DomainStatsEntry{}
		fmt.Printf("[DOMAIN-ADD] Domain '%s' manually registered in the matrix.\n", dom)
	}
}
