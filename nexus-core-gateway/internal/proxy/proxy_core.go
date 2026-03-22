package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
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
	Router    *DynamicRouter // Dynamic multi-host routing engine
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
		Router:    NewDynamicRouter(10 * time.Second), // 10s local cache TTL
	}
	atomic.StorePointer(&np.proxyPtr, unsafe.Pointer(initialProxy))

	// Pre-load current target to Redis for dynamic routing consistency
	np.Router.AddRoute("localhost:8080", target)
	np.Router.AddRoute("ojk.localhost", target)
	np.Router.AddRoute("kemenkeu.localhost", target)
	np.Router.AddRoute("bi.localhost", target)

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

	// SECURITY BYPASS: Jangan memfilter internal API (Dashboard & NECHAT)
	// Kita tidak ingin admin diblokir saat mendiskusikan "SQL Injection" dengan AI.
	if len(r.URL.Path) >= 5 && r.URL.Path[:5] == "/api/" {
		np.getProxy().ServeHTTP(w, r)
		return
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
			TargetDomain: r.Host,
			ThreatDetail: threatType,
			LatencyMS:    latency,
		}
		np.Logger.EnrichLog(&tLog, r)
		np.Logger.LogTraffic(tLog)

		// LOG AI COGNITIVE EVENT
		np.Logger.LogAIEvent(logger.AIEventLog{
			Layer:        "Reflex",
			Status:       "Mitigating",
			DetailAction: fmt.Sprintf("Dynamic WAF Rule Generated: Blocked pattern '%s' from IP %s", threatType, r.RemoteAddr),
		})

		// CRITICAL DEFENSE ASYNC: Block at OS Level for identified threat patterns
		// (Only for severe threats like SQLi or Path Traversal)
		if strings.Contains(threatType, "SQL_INJECTION") || strings.Contains(threatType, "DIRECTORY_TRAVERSAL") {
			go mtd.BlockIPAtOSLevel(r.RemoteAddr)
		}

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
				TargetDomain: req.Host,
				LatencyMS:    time.Since(start).Milliseconds(),
			}
			np.Logger.EnrichLog(&tLog, req)
			np.Logger.LogTraffic(tLog)

			// LOG AI COGNITIVE EVENT
			np.Logger.LogAIEvent(logger.AIEventLog{
				Layer:        "Reasoning",
				Status:       "Analyzing",
				DetailAction: fmt.Sprintf("Deep payload inspection flagged malicious intent from %s", source),
			})
			fmt.Printf("[ALERT] Reasoning Layer: BYPASS from %s -> Flagged for MTD blacklist\n", source)
		}
	}(analysisData, r.RemoteAddr, r.URL.Path, r)

	// 4. Proxy Traffic to Current MTD Target (atomic load — always fresh)
	latency := time.Since(start).Milliseconds()
	tLog := logger.TelemetryLog{
		Timestamp:    time.Now(),
		SourceIP:     r.RemoteAddr,
		Endpoint:     r.URL.Path,
		Method:       r.Method,
		Status:       "ALLOWED",
		TargetDomain: r.Host,
		LatencyMS:    latency,
	}
	np.Logger.EnrichLog(&tLog, r)
	np.Logger.LogTraffic(tLog)

	targetURL, found := np.Router.Lookup(r.Host)
	if !found {
		// Domain not protected, return 404
		fmt.Printf("[NEXUS] ROUTING_ERROR: Host '%s' not found in dynamic router table.\n", r.Host)
		http.Error(w, fmt.Sprintf("NEXUS [404]: Domain '%s' is not protected by this matrix.", r.Host), http.StatusNotFound)
		return
	}

	// Dynamic proxying to the discovered target
	remote, err := url.Parse(targetURL)
	if err != nil {
		http.Error(w, "INTERNAL_GATEWAY_ERROR: Invalid target configuration.", http.StatusInternalServerError)
		return
	}

	// Create a new proxy for the specific target
	// IMPROVEMENT: In the future, we can pool these or cache them in RouteEntry
	dynProxy := httputil.NewSingleHostReverseProxy(remote)
	dynProxy.ServeHTTP(w, r)
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
