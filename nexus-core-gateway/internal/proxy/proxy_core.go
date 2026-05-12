package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/nexus-cyber/nexus-core-gateway/internal/ai"
	"github.com/nexus-cyber/nexus-core-gateway/internal/avse"
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
	// [NEW: VIRTUAL PATCHING] Memory-Resident Antibodies for instant O(1) blocking
	Patches      sync.Map
	PatchesCount int32
	// [NEW: THREAT MAP] Internal hub for live visualization in Local Mode
	ThreatListeners sync.Map // Map[string]chan string for multi-client SSE fan-out
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
	np.Router.AddRoute("localhost", target)
	np.Router.AddRoute("ojk.localhost", target)
	np.Router.AddRoute("kemenkeu.localhost", target)
	np.Router.AddRoute("bi.localhost", target)
	// [NEW: VIRTUAL PATCHING] Activate background antibody synchronization
	np.StartImmunitySync()

	return np, nil
}

// [NEW: VIRTUAL PATCHING] StartImmunitySync initiates the background autonomous sync
// from Redis. This ensures antibodies created on one gateway node are shared instantly.
func (np *NexusProxy) StartImmunitySync() {
	if mtd.MtdRedis == nil || !mtd.MtdRedis.Enabled {
		return
	}

	ticker := time.NewTicker(10 * time.Second)
	go func() {
		ctx := context.Background()
		for range ticker.C {
			// Pull "nexus:virtual_patches" from Redis Set
			patches, err := mtd.MtdRedis.Client.SMembers(ctx, "nexus:virtual_patches").Result()
			if err == nil {
				// Atomic update of the local antibody map
				var count int32
				for _, p := range patches {
					np.Patches.Store(p, true)
					count++
				}
				atomic.StoreInt32(&np.PatchesCount, count)
			}
		}
	}()
}

// [NEW: VIRTUAL PATCHING] AddAntibody caches a malicious pattern for instant blocking
func (np *NexusProxy) AddAntibody(payload string) {
	// 1. Local Cache Insert (Sub-millisecond protection)
	np.Patches.Store(payload, true)
	atomic.AddInt32(&np.PatchesCount, 1)

	// 2. Global Persistence Sync (Share with other Nexus Gateways)
	if mtd.MtdRedis != nil && mtd.MtdRedis.Enabled {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			mtd.MtdRedis.Client.SAdd(ctx, "nexus:virtual_patches", payload)
		}()
	}
}

// UpdateTarget atomically swaps the reverse proxy to a new backend.
// This is the Graceful Handoff mechanism — in-flight requests on old proxy
// complete normally; new requests go to the new target.
// ResetAntibodies clears all learned malicious patterns (Virtual Patches)
func (np *NexusProxy) ResetAntibodies() {
	// 1. Local Cache Purge
	np.Patches.Range(func(key, value interface{}) bool {
		np.Patches.Delete(key)
		return true
	})
	atomic.StoreInt32(&np.PatchesCount, 0)

	// 2. Global Persistence Wipe (Redis)
	if mtd.MtdRedis != nil && mtd.MtdRedis.Enabled {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		mtd.MtdRedis.Client.Del(ctx, "nexus:virtual_patches")
	}
	fmt.Println("[SYSTEM-RESET] Virtual Patch Antibody database purged.")
}

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
	// 1. Capture payload for analysis
	var body []byte
	if r.Body != nil {
		body, _ = io.ReadAll(r.Body)
		
		// [AVSE - PHASE 1 INTEGRATION]
		// Cek apakah request ini berisi gambar
		contentType := r.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "image/jpeg") || strings.HasPrefix(contentType, "image/png") {
			cleanResult, err := avse.SanitizeImage(body)
			if err != nil {
				// [PHASE 4: ANTI-BOMB PROTECTION]
				// Jika gambar terlalu besar atau rusak, blokir demi keamanan server
				np.Logger.LogAIEvent(logger.AIEventLog{
					Layer:        "AVSE (Visual Shield)",
					Status:       "BLOCKED",
					DetailAction: fmt.Sprintf("Blocked suspicious image: %v", err),
				})
				http.Error(w, "Nexus [403]: Image blocked for security reasons (Resolution too high or corrupt).", http.StatusForbidden)
				return
			}
			
			// Gunakan data yang sudah dicuci
			body = cleanResult.Data
			
			// Log ke Dashboard (Fase 3: Risk Reporting)
			status := "SANITIZED"
			if cleanResult.RiskScore > 70 {
				status = "THREAT_CLEANED"
			}
			
			np.Logger.LogAIEvent(logger.AIEventLog{
				Layer:        "AVSE (Visual Shield)",
				Status:       status,
				DetailAction: fmt.Sprintf("Visual Clean [%d%% Risk]: %d B -> %d B (%s)", cleanResult.RiskScore, cleanResult.OriginalSize, cleanResult.CleanedSize, cleanResult.Format),
			})
		}

		r.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	// SECURITY BYPASS: Jangan memfilter internal API (Dashboard & NECHAT)
	// Kita tidak ingin admin diblokir saat mendiskusikan "SQL Injection" dengan AI.
	if len(r.URL.Path) >= 5 && r.URL.Path[:5] == "/api/" {
		np.getProxy().ServeHTTP(w, r)
		return
	}

	// [NEW] Host normalization: Strip port if present (fixes localhost:8080 match)
	host := r.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	targetURL, found := np.Router.Lookup(host)
	if !found {
		// Domain not protected, return 404
		fmt.Printf("[NEXUS] ROUTING_ERROR: Host '%s' (normalized: '%s') not found in dynamic router table.\n", r.Host, host)
		http.Error(w, fmt.Sprintf("NEXUS [404]: Domain '%s' is not protected by this matrix.", host), http.StatusNotFound)
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

// [NEW: THREAT MAP] ThreatData represents a visual attack event for the 3D map
type ThreatData struct {
	ID         string  `json:"id"`
	AttackerIP string  `json:"attacker_ip"`
	SourceLat  float64 `json:"source_lat"`
	SourceLng  float64 `json:"source_lng"`
	TargetLat  float64 `json:"target_lat"`
	TargetLng  float64 `json:"target_lng"`
	Type       string  `json:"type"`
}

// PublishThreat broadcasts a threat event with real IP tracking or simulation fallback
func (np *NexusProxy) PublishThreat(ip string, threatType string) {
	var lat, lng float64
	sourceName := "SIMULATED_VEC"

	// 🛰️ REAL-IP SATELLITE TRACE
	cleanIP := ip
	if strings.Contains(ip, ":") {
		cleanIP = strings.Split(ip, ":")[0] // Strip port
	}

	isLocal := cleanIP == "127.0.0.1" || cleanIP == "::1" || cleanIP == "localhost"

	if !isLocal {
		// Attempt to resolve real Location via Geo-API
		resp, err := http.Get("http://ip-api.com/json/" + cleanIP)
		if err == nil {
			defer resp.Body.Close()
			var geo struct {
				Lat     float64 `json:"lat"`
				Lon     float64 `json:"lon"`
				Country string  `json:"country"`
			}
			if json.NewDecoder(resp.Body).Decode(&geo) == nil && geo.Lat != 0 {
				lat = geo.Lat
				lng = geo.Lon
				sourceName = geo.Country
			}
		}
	}

	// 🛡️ FALLBACK TO WAR-ROOM SIMULATION (If Local or API Failed)
	if lat == 0 {
		sources := [][]float64{
			{55.75, 37.61}, {39.90, 116.40}, {38.90, -77.03}, {51.50, -0.12},
			{35.67, 139.65}, {48.85, 2.35}, {52.52, 13.40}, {-23.55, -46.63},
			{-33.86, 151.20}, {1.35, 103.81},
		}
		src := sources[time.Now().UnixNano()%int64(len(sources))]
		lat, lng = src[0], src[1]
	}

	// 🔵 DYNAMIC TARGET LOCALIZATION (Nexus Sentinel Hub)
	targetLat, targetLng := -6.20, 106.81 // Default: Jakarta
	domain := strings.ToLower(threatType) // This is a bit dirty, let's use a mapping logic later
	if strings.Contains(domain, "portal") {
		targetLat, targetLng = 1.35, 103.81 // Singapore
	} else if strings.Contains(domain, "audit") {
		targetLat, targetLng = -33.86, 151.20 // Sydney
	} else if strings.Contains(domain, "cloud") {
		targetLat, targetLng = 50.11, 8.68 // Frankfurt
	}

	threat := ThreatData{
		ID:         fmt.Sprintf("TRT-%d-%s", time.Now().UnixNano(), cleanIP),
		AttackerIP: cleanIP,
		SourceLat:  lat,
		SourceLng:  lng,
		TargetLat:  targetLat,
		TargetLng:  targetLng,
		Type:       threatType + "_" + sourceName,
	}

	payload, _ := json.Marshal(threat)
	msg := string(payload)

	// Redis Broadcast
	if mtd.MtdRedis != nil && mtd.MtdRedis.Enabled {
		ctx := context.Background()
		mtd.MtdRedis.Client.Publish(ctx, "nexus:threat_stream", payload)
	}

	// Internal Broadcast to Dashboards
	np.ThreatListeners.Range(func(key, value interface{}) bool {
		ch := value.(chan string)
		select {
		case ch <- msg:
		default:
		}
		return true
	})
}

// routeToHoneypot performs silent NAT to the Honeypot — Digital Hallucination.
func (np *NexusProxy) routeToHoneypot(w http.ResponseWriter, r *http.Request) {
	honeypotURL := "http://localhost:9090" // Honeypot's internal address
	target, _ := url.Parse(honeypotURL)
	hp := httputil.NewSingleHostReverseProxy(target)
	r.Host = target.Host
	hp.ServeHTTP(w, r)
}
