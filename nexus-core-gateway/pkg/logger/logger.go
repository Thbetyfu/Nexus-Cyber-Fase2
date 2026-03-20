package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// TelemetryLog represents traffic metadata for Nexus Dashboard.
type TelemetryLog struct {
	Timestamp    time.Time `json:"timestamp"`
	SourceIP     string    `json:"source_ip"`
	Endpoint     string    `json:"endpoint"`
	Method       string    `json:"method"`
	Status       string    `json:"status"` // ALLOWED, BLOCKED, FLAG_TO_REASONING
	ThreatDetail string    `json:"threat_detail,omitempty"`
	LatencyMS    int64     `json:"latency_ms"`
}

type Logger struct {
	file       *os.File
	mu         sync.RWMutex
	recentLogs []TelemetryLog

	// Global Counters
	TotalAllowed  int
	TotalBlocked  int
	TotalHoneypot int
}

func NewLogger() (*Logger, error) {
	f, err := os.OpenFile("nexus_traffic.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &Logger{
		file:       f,
		recentLogs: make([]TelemetryLog, 0, 100),
	}, nil
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
