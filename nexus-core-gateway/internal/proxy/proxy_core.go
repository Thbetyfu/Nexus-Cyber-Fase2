package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/nexus-cyber/nexus-core-gateway/internal/ai"
	"github.com/nexus-cyber/nexus-core-gateway/internal/mtd"
	"github.com/nexus-cyber/nexus-core-gateway/pkg/logger"
)

// NexusProxy implements the core @skill-mtd and @skill-dual-brain gateway.
// Phase 5: MTD-aware proxy with Honeypot routing, Token Bucket, and shuffled backends.
type NexusProxy struct {
	// atomic pointer to current reverse proxy — updated on MTD shuffle
	proxyPtr  unsafe.Pointer
	Filter    *ai.ReflexFilter
	Reasoning *ai.ReasoningEngine
	Logger    *logger.Logger
	Honeypot  *mtd.HoneypotServer
	Shuffler  *mtd.TopologyShuffler
}

func NewNexusProxy(
	target string,
	filter *ai.ReflexFilter,
	reasoning *ai.ReasoningEngine,
	log *logger.Logger,
	shuffler *mtd.TopologyShuffler,
	honeypot *mtd.HoneypotServer,
) (*NexusProxy, error) {
	remote, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	initialProxy := httputil.NewSingleHostReverseProxy(remote)

	np := &NexusProxy{
		Filter:    filter,
		Reasoning: reasoning,
		Logger:    log,
		Honeypot:  honeypot,
		Shuffler:  shuffler,
	}
	atomic.StorePointer(&np.proxyPtr, unsafe.Pointer(initialProxy))
	return np, nil
}

// UpdateTarget atomically swaps the reverse proxy to a new backend.
// This is the Graceful Handoff mechanism — in-flight requests on old proxy
// complete normally; new requests go to the new target.
func (np *NexusProxy) UpdateTarget(newTarget string) error {
	remote, err := url.Parse(newTarget)
	if err != nil {
		return fmt.Errorf("mtd_handoff: invalid target URL: %v", err)
	}
	newProxy := httputil.NewSingleHostReverseProxy(remote)
	atomic.StorePointer(&np.proxyPtr, unsafe.Pointer(newProxy))
	return nil
}

// getProxy returns the currently active reverse proxy (thread-safe via atomic pointer).
func (np *NexusProxy) getProxy() *httputil.ReverseProxy {
	return (*httputil.ReverseProxy)(atomic.LoadPointer(&np.proxyPtr))
}

func (np *NexusProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// 1. Capture payload for analysis
	var body []byte
	if r.Body != nil {
		body, _ = io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	// 2. Dual-Brain Layer 1: Reflex (Sync - Fast)
	query, _ := url.QueryUnescape(r.URL.RawQuery)
	analysisData := query + " " + string(body)
	isReflexThreat, threatType := np.Filter.InspectRequest(analysisData)

	if isReflexThreat {
		latency := time.Since(start).Milliseconds()
		tLog := logger.TelemetryLog{
			Timestamp:    time.Now(),
			SourceIP:     r.RemoteAddr,
			Endpoint:     r.URL.Path,
			Method:       r.Method,
			Status:       "HONEYPOT_REDIRECTED",
			ThreatDetail: threatType,
			LatencyMS:    latency,
		}
		np.Logger.EnrichLog(&tLog, r)
		np.Logger.LogTraffic(tLog)

		// MTD Phase 5: DIGITAL HALLUCINATION
		// Instead of dropping, silently redirect to Honeypot Tarpit.
		np.routeToHoneypot(w, r)
		return
	}

	// 3. Dual-Brain Layer 2: Reasoning (Async - Intent Analysis)
	go func(tData string, source string, path string, req *http.Request) {
		isMalicious, err := np.Reasoning.AnalyzeIntent(tData)
		if err != nil {
			return // Fail-Open on AI timeout
		}

		if isMalicious {
			tLog := logger.TelemetryLog{
				Timestamp:    time.Now(),
				SourceIP:     source,
				Endpoint:     path,
				Method:       "ASYNC",
				Status:       "MALICIOUS_DETECTED_REASONING",
				ThreatDetail: "Llama3_Intent_Analysis_Malicious",
				LatencyMS:    time.Since(start).Milliseconds(),
			}
			np.Logger.EnrichLog(&tLog, req)
			np.Logger.LogTraffic(tLog)
			fmt.Printf("[ALERT] Reasoning Layer: BYPASS from %s -> Flagged for MTD blacklist\n", source)
		}
	}(analysisData, r.RemoteAddr, r.URL.Path, r)

	// 4. Proxy Traffic to Current MTD Target (atomic load — always fresh)
	latency := time.Since(start).Milliseconds()
	tLog := logger.TelemetryLog{
		Timestamp: time.Now(),
		SourceIP:  r.RemoteAddr,
		Endpoint:  r.URL.Path,
		Method:    r.Method,
		Status:    "ALLOWED",
		LatencyMS: latency,
	}
	np.Logger.EnrichLog(&tLog, r)
	np.Logger.LogTraffic(tLog)
	np.getProxy().ServeHTTP(w, r)
}

// routeToHoneypot performs silent NAT to the Honeypot — Digital Hallucination.
// The attacker NEVER knows they've been redirected.
func (np *NexusProxy) routeToHoneypot(w http.ResponseWriter, r *http.Request) {
	honeypotURL := "http://localhost:9090" // Honeypot's internal address
	target, _ := url.Parse(honeypotURL)
	hp := httputil.NewSingleHostReverseProxy(target)
	r.Host = target.Host
	hp.ServeHTTP(w, r)
}
