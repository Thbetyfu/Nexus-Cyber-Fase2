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
	TargetDomain      string    `json:"target_domain"` // Domain tracking for Multi-Tenancy
	LatencyMS         int64     `json:"latency_ms"`
}

// AIEventLog records cognitive activities and self-repair actions
type AIEventLog struct {
	Timestamp    time.Time `json:"timestamp"`
	Layer        string    `json:"layer"`         // e.g., "Reflex", "Reasoning", "Self-Repair"
	Status       string    `json:"status"`        // e.g., "Analyzing", "Mitigating", "Repairing"
	DetailAction string    `json:"detail_action"` // e.g., "Blocked pattern 'UNION SELECT' from IP X"
}

type Logger struct {
	file           *os.File
	aiFile         *os.File
	mu             sync.RWMutex
	recentLogs     []TelemetryLog
	recentAIEvents []AIEventLog
	OnAIEvent      func(AIEventLog)

	// Global Counters
	TotalAllowed  int
	TotalBlocked  int
	TotalHoneypot int
	TotalPanic    int

	// Domain Stats (Cumulative per Workspace)
	DomainStats map[string]*DomainStatsEntry

	// Profiling Cache for performance (ISO 25010 Efficiency)
	fingerprintCache map[string]TelemetryLog
}

func NewLogger() (*Logger, error) {
	f, err := os.OpenFile("nexus_traffic.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	aiF, err := os.OpenFile("nexus_ai_events.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		file:             f,
		aiFile:           aiF,
		recentLogs:       make([]TelemetryLog, 0),
		recentAIEvents:   make([]AIEventLog, 0),
		fingerprintCache: make(map[string]TelemetryLog),
		DomainStats:      make(map[string]*DomainStatsEntry),
	}, nil
}

type DomainStatsEntry struct {
	Allowed  int
	Honeypot int
	Blocked  int
}

// EnrichLog adds attacker profiling for security intelligence.
// Implements SHA-256 fingerprinting and GeoIP lookup logic.
func (l *Logger) EnrichLog(log *TelemetryLog, r *http.Request) {
	if r == nil {
		return
	}

	// 0. Extract Domain (Multi-Tenant Hub)
	host := r.Host
	if host == "" {
		host = "all"
	}
	log.TargetDomain = host

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
	isLocal := strings.HasPrefix(log.SourceIP, "127.") ||
		strings.HasPrefix(log.SourceIP, "::1") ||
		strings.HasPrefix(log.SourceIP, "[::1]") ||
		log.SourceIP == "localhost"

	if isLocal {
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
	// Persistence (JSON Line standard)
	data, _ := json.Marshal(log)
	l.file.WriteString(string(data) + "\n")

	l.mu.Lock()
	defer l.mu.Unlock()

	// 1. Update Global Counters (Dashboard Main)
	switch log.Status {
	case "ALLOWED":
		l.TotalAllowed++
	case "HONEYPOT_REDIRECTED", "DIVERTED_TO_HONEYPOT":
		l.TotalHoneypot++
	case "RATE_LIMITED", "BLOCKED":
		l.TotalBlocked++
	}

	// 2. Update Multi-Tenant Counters (Workspace Selection)
	if log.TargetDomain != "" {
		dom := strings.ToLower(log.TargetDomain)
		// Domain Normalization: Strip potential port suffixes added by clients
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
		case "HONEYPOT_REDIRECTED", "DIVERTED_TO_HONEYPOT":
			stats.Honeypot++
		case "RATE_LIMITED", "BLOCKED":
			stats.Blocked++
		}
	}

	// 3. Maintain Memory Buffer (Live Terminal)
	l.recentLogs = append(l.recentLogs, log)
	if len(l.recentLogs) > 50 {
		l.recentLogs = l.recentLogs[len(l.recentLogs)-50:]
	}
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

// LogAIEvent records cognitive AI events
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

	// Trigger real-time broadcast if callback is set
	if l.OnAIEvent != nil {
		l.OnAIEvent(event)
	}
	l.mu.Unlock()
}

// GetRecentAIEvents returns the 20 most recent AI activity logs
func (l *Logger) GetRecentAIEvents() []AIEventLog {
	l.mu.RLock()
	defer l.mu.RUnlock()

	cpy := make([]AIEventLog, len(l.recentAIEvents))
	copy(cpy, l.recentAIEvents)
	return cpy
}

func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
	}
	if l.aiFile != nil {
		l.aiFile.Close()
	}
}

func (l *Logger) GetDomainStats(domain string) (Allowed, Blocked, Honeypot int) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if stats, ok := l.DomainStats[domain]; ok {
		return stats.Allowed, stats.Blocked, stats.Honeypot
	}
	return 0, 0, 0
}

func (l *Logger) GetDomains() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	domains := make([]string, 0, len(l.DomainStats))
	for d := range l.DomainStats {
		domains = append(domains, d)
	}
	return domains
}
