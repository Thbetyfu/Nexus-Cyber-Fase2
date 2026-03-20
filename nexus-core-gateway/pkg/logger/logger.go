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

	"github.com/mssola/user_agent"
)

// TelemetryLog represents traffic metadata for Nexus Dashboard.
type TelemetryLog struct {
	Timestamp         time.Time `json:"timestamp"`
	SourceIP          string    `json:"source_ip"`
	AttackerID        string    `json:"attacker_id,omitempty"` // Derived from IP + UA
	GeoLocation       string    `json:"geo_location,omitempty"`
	ISP               string    `json:"isp,omitempty"`
	DeviceFingerprint string    `json:"device_fingerprint,omitempty"`
	Endpoint          string    `json:"endpoint"`
	Method            string    `json:"method"`
	Status            string    `json:"status"` // ALLOWED, BLOCKED, FLAG_TO_REASONING
	ThreatDetail      string    `json:"threat_detail,omitempty"`
	LatencyMS         int64     `json:"latency_ms"`
}

type Logger struct {
	file       *os.File
	mu         sync.RWMutex
	recentLogs []TelemetryLog

	// Global Counters
	TotalAllowed  int
	TotalBlocked  int
	TotalHoneypot int

	// Profiling Cache for performance (ISO 25010 Efficiency)
	fingerprintCache map[string]TelemetryLog
}

func NewLogger() (*Logger, error) {
	f, err := os.OpenFile("nexus_traffic.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &Logger{
		file:             f,
		recentLogs:       make([]TelemetryLog, 0, 100),
		fingerprintCache: make(map[string]TelemetryLog),
	}, nil
}

// EnrichLog adds attacker profiling for security intelligence.
// Implements SHA-256 fingerprinting and GeoIP lookup logic.
func (l *Logger) EnrichLog(log *TelemetryLog, r *http.Request) {
	if r == nil {
		return
	}

	uaStr := r.Header.Get("User-Agent")
	// Use IP + UA for fingerprinting (ISO 25010 Reliability)
	cacheKey := log.SourceIP + uaStr

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

	// 1. Digital Fingerprinting (SHA-256)
	h := sha256.New()
	h.Write([]byte(cacheKey))
	log.AttackerID = fmt.Sprintf("APT-ID-%X", h.Sum(nil)[:4])

	// 2. User-Agent Profiling
	ua := user_agent.New(uaStr)
	osInfo := ua.OS()
	browser, _ := ua.Browser()
	log.DeviceFingerprint = fmt.Sprintf("%s (%s)", osInfo, browser)

	// 3. GeoIP Lookup (Simulator via Local/External lookup)
	if strings.HasPrefix(log.SourceIP, "127.") || log.SourceIP == "::1" || log.SourceIP == "[::1]" {
		log.GeoLocation = "Localhost, Nexus Gate"
		log.ISP = "Internal Loopback"
	} else {
		// Mock dynamic GeoIP (can be integrated with ip-api.com)
		log.GeoLocation = "Global, Secure Zone"
		if strings.Contains(uaStr, "curl") || strings.Contains(uaStr, "python") {
			log.ISP = "Automated Bot/Scanner"
		} else {
			log.ISP = "Residential User"
		}
	}

	// Update Cache for performance (Efficiency)
	l.mu.Lock()
	l.fingerprintCache[cacheKey] = *log
	l.mu.Unlock()
}

func (l *Logger) LogTraffic(log TelemetryLog) {
	data, _ := json.Marshal(log)
	fmt.Println(string(data)) // Concurrent CLI output
	l.file.WriteString(string(data) + "\n")

	// Store in memory for Live Telemetry Dashboard
	l.mu.Lock()

	// Update Counters
	switch log.Status {
	case "ALLOWED":
		l.TotalAllowed++
	case "HONEYPOT_REDIRECTED", "DIVERTED_TO_HONEYPOT":
		l.TotalHoneypot++
	case "RATE_LIMITED", "BLOCKED":
		l.TotalBlocked++
	}

	l.recentLogs = append(l.recentLogs, log)
	if len(l.recentLogs) > 100 {
		// Keep last 100 logs
		l.recentLogs = l.recentLogs[len(l.recentLogs)-100:]
	}
	l.mu.Unlock()
}

// GetRecentLogs returns a copy of the recent logs for the API.
func (l *Logger) GetRecentLogs() []TelemetryLog {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Return a copy to prevent race conditions during JSON serialization
	cpy := make([]TelemetryLog, len(l.recentLogs))
	copy(cpy, l.recentLogs)
	return cpy
}

func (l *Logger) Close() {
	l.file.Close()
}
