// Package proxy mengimplementasikan gateway proxy reverse otonom dengan kecerdasan MTD.
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

// NexusProxy adalah inti penggerak dari gerbang pertahanan siber otonom (SOC Gateway).
//
// Alasan Arsitektural (Why):
// NexusProxy bertindak sebagai mediator pusat (Mediator Pattern) yang mengkoordinasikan MTD (Moving Target Defense),
// sistem kecerdasan buatan berlapis Dual-Brain, dan pembersihan visual AVSE.
// Modul ini dirancang agar tahan terhadap beban trafik tinggi (High Throughput) dengan tingkat ketersediaan tinggi
// (High Availability) sesuai parameter ISO 25010.
type NexusProxy struct {
	// Pointer atomik ke reverse proxy aktif untuk memfasilitasi rotasi port Lock-Free.
	proxyPtr  unsafe.Pointer
	Filter    *ai.ReflexFilter
	Reasoning *ai.ReasoningEngine
	Logger    *logger.Logger
	Honeypot  *mtd.HoneypotServer
	Shuffler  *mtd.TopologyShuffler
	Router    *DynamicRouter // Router multi-host dinamis terdistribusi
	
	// Peta antibodi (virtual patches) resident memori untuk pencocokan instan O(1)
	Patches      sync.Map
	PatchesCount int32
	
	// SSE Multi-client Fan-Out untuk visualisasi real-time 3D di dasbor Command Center
	ThreatListeners sync.Map
}

// NewNexusProxy mengonstruksi gerbang pertahanan NexusProxy.
//
// Alasan Teknis (Why):
// Menggunakan penunjuk atomik (Atomic Pointer Store) untuk menginisialisasi target awal.
// Menyinkronkan entri pemetaan domain dasar (`localhost`, `ojk.localhost`) secara asinkron ke Redis,
// serta menyalakan modul sinkronisasi antibodi otomatis.
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
		Router:    NewDynamicRouter(10 * time.Second), // Cache TTL lokal 10 detik
	}
	atomic.StorePointer(&np.proxyPtr, unsafe.Pointer(initialProxy))

	// Inisialisasi rute standard di tabel dinamis.
	np.Router.AddRoute("localhost", target)
	np.Router.AddRoute("ojk.localhost", target)
	np.Router.AddRoute("kemenkeu.localhost", target)
	np.Router.AddRoute("bi.localhost", target)
	
	// Jalankan sinkronisasi background antibodi imun.
	np.StartImmunitySync()

	return np, nil
}

// StartImmunitySync menginisialisasi goroutine latar belakang untuk menarik berkas antibodi (virtual patches) dari Redis.
// Menjamin setiap node gateway Nexus saling berbagi ilmu pertahanan secara instan (Global Immunity Sync).
func (np *NexusProxy) StartImmunitySync() {
	if mtd.MtdRedis == nil || !mtd.MtdRedis.Enabled {
		return
	}

	ticker := time.NewTicker(10 * time.Second)
	go func() {
		ctx := context.Background()
		for range ticker.C {
			patches, err := mtd.MtdRedis.Client.SMembers(ctx, "nexus:virtual_patches").Result()
			if err == nil {
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

// AddAntibody mendaftarkan pola serangan siber mencurigakan ke database memori lokal dan Redis terdistribusi.
func (np *NexusProxy) AddAntibody(payload string) {
	// 1. Sisipkan ke cache lokal (O(1) protection) untuk proteksi instan sub-milidetik.
	np.Patches.Store(payload, true)
	atomic.AddInt32(&np.PatchesCount, 1)

	// 2. Publikasikan secara asinkron ke server kluster Redis.
	if mtd.MtdRedis != nil && mtd.MtdRedis.Enabled {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			mtd.MtdRedis.Client.SAdd(ctx, "nexus:virtual_patches", payload)
		}()
	}
}

// ResetAntibodies membersihkan seluruh memori antibodi lokal dan Redis (System Wiping / Reset).
func (np *NexusProxy) ResetAntibodies() {
	// 1. Bersihkan RAM lokal
	np.Patches.Range(func(key, value interface{}) bool {
		np.Patches.Delete(key)
		return true
	})
	atomic.StoreInt32(&np.PatchesCount, 0)

	// 2. Bersihkan Redis
	if mtd.MtdRedis != nil && mtd.MtdRedis.Enabled {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		mtd.MtdRedis.Client.Del(ctx, "nexus:virtual_patches")
	}
	fmt.Println("[SYSTEM-RESET] Virtual Patch Antibody database purged.")
}

// UpdateTarget menukar target reverse proxy secara dinamis tanpa mengganggu request yang sedang berjalan.
//
// Alasan Arsitektural (Why):
// Mekanisme pertukaran target Lock-Free (Graceful Handoff).
// Menggunakan operasi penunjuk atomik (Atomic Store) untuk mengarahkan request HTTP baru ke port backend MTD yang baru
// secara instan, sementara request lama yang sedang diproses di proxy lama tetap dibiarkan selesai secara normal.
// Ini menjamin pemotongan (switching) port tidak mengakibatkan kegagalan paket trafik (Zero Packet Loss).
func (np *NexusProxy) UpdateTarget(newTarget string) error {
	remote, err := url.Parse(newTarget)
	if err != nil {
		return fmt.Errorf("mtd_handoff: invalid target URL: %v", err)
	}
	newProxy := httputil.NewSingleHostReverseProxy(remote)
	atomic.StorePointer(&np.proxyPtr, unsafe.Pointer(newProxy))
	return nil
}

// getProxy membaca reverse proxy yang sedang aktif secara thread-safe menggunakan Atomic Load.
func (np *NexusProxy) getProxy() *httputil.ReverseProxy {
	return (*httputil.ReverseProxy)(atomic.LoadPointer(&np.proxyPtr))
}

// ServeHTTP adalah pintu masuk utama seluruh paket data HTTP yang melintasi gerbang proxy.
func (np *NexusProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. Capture payload dan bersihkan malware visual via AVSE.
	var body []byte
	if r.Body != nil {
		body, _ = io.ReadAll(r.Body)
		
		// [AVSE - INTELLIGENT MULTIMEDIA FILTERING]
		// Alasan Keamanan (Why):
		// Peretas sering menyisipkan malware biner di dalam berkas gambar JPEG/PNG (steganografi).
		// Gateway mendeteksi tipe konten gambar dan mengirimkannya ke `avse.SanitizeImage` untuk kompresi ulang
		// biner steril, menghancurkan payload eksploitasi visual tanpa merusak estetika gambar asli.
		contentType := r.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "image/jpeg") || strings.HasPrefix(contentType, "image/png") {
			cleanResult, err := avse.SanitizeImage(body)
			if err != nil {
				// [ANTI-IMAGE BOMB PROTECTION]
				// Blokir seketika jika struktur gambar rusak atau resolusi melampaui batas Megapiksel wajar.
				np.Logger.LogAIEvent(logger.AIEventLog{
					Layer:        "AVSE (Visual Shield)",
					Status:       "BLOCKED",
					DetailAction: fmt.Sprintf("Blocked suspicious image: %v", err),
				})
				http.Error(w, "Nexus [403]: Image blocked for security reasons (Resolution too high or corrupt).", http.StatusForbidden)
				return
			}
			
			// Ganti body request dengan byte gambar steril yang telah dibersihkan.
			body = cleanResult.Data
			
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

	// [BYPASS INTERNAL CHAT APIS]
	// Alasan Teknis (Why):
	// Diskusi admin dengan modul asisten AI mengenai celah SQLi/XSS tidak boleh disaring atau disanitasi oleh gateway,
	// karena akan menyebabkan kegagalan respon chat admin (False Positive Lockout).
	if len(r.URL.Path) >= 5 && r.URL.Path[:5] == "/api/" {
		np.getProxy().ServeHTTP(w, r)
		return
	}

	// Normalisasi Domain (Potong port pemanggil untuk pencocokan tabel router)
	host := r.Host
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	// Cari pemetaan domain ke backend steril otonom.
	targetURL, found := np.Router.Lookup(host)
	if !found {
		fmt.Printf("[NEXUS] ROUTING_ERROR: Host '%s' (normalized: '%s') not found in dynamic router table.\n", r.Host, host)
		http.Error(w, fmt.Sprintf("NEXUS [404]: Domain '%s' is not protected by this matrix.", host), http.StatusNotFound)
		return
	}

	// Eksekusi Proxying dinamis ke tujuan backend steril yang tepat dengan Enkripsi Transpilasi Polimorfik PACS.
	remote, err := url.Parse(targetURL)
	if err != nil {
		http.Error(w, "INTERNAL_GATEWAY_ERROR: Invalid target configuration.", http.StatusInternalServerError)
		return
	}

	dynProxy := httputil.NewSingleHostReverseProxy(remote)
	dynProxy.ModifyResponse = func(resp *http.Response) error {
		// Deteksi tipe konten HTML untuk pengacakan sandi visual (PACS)
		contentType := resp.Header.Get("Content-Type")
		if strings.Contains(contentType, "text/html") {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			resp.Body.Close()

			// Jalankan transpilasi biner alien
			obfuscated := ObfuscateHTML(string(bodyBytes), host)

			resp.Body = io.NopCloser(strings.NewReader(obfuscated))
			resp.ContentLength = int64(len(obfuscated))
			resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(obfuscated)))
		}
		return nil
	}

	dynProxy.ServeHTTP(w, r)
}

// ThreatData merepresentasikan skema koordinat geografis serangan siber untuk peta ancaman 3D (3D Threat Map).
type ThreatData struct {
	ID         string  `json:"id"`
	AttackerIP string  `json:"attacker_ip"`
	SourceLat  float64 `json:"source_lat"`
	SourceLng  float64 `json:"source_lng"`
	TargetLat  float64 `json:"target_lat"`
	TargetLng  float64 `json:"target_lng"`
	Type       string  `json:"type"`
}

// PublishThreat mempublikasikan visualisasi serangan siber otonom secara global dan real-time.
//
// Alasan Arsitektural (Why):
// - Menggunakan pelacakan IP Geografis real-time via API eksternal (ip-api.com) jika IP bersifat publik.
// - Menyediakan simulasi koordinat acak real-time berkoordinat negara peretas (China, Rusia, USA, Singapura, Jerman)
//   jika sistem berjalan secara offline (Local Host Mode), menjaga kemeriahan visualisasi Command Center (NCC).
// - Mengirim data serangan ke kluster Redis Pub/Sub untuk sinkronisasi multi-dashboard,
//   serta menyiarkan acara (broadcast) secara aman ke Map Listener SSE internal tanpa penundaan (sub-milidetik latency).
func (np *NexusProxy) PublishThreat(ip string, threatType string) {
	var lat, lng float64
	sourceName := "SIMULATED_VEC"

	cleanIP := ip
	if strings.Contains(ip, ":") {
		cleanIP = strings.Split(ip, ":")[0]
	}

	isLocal := cleanIP == "127.0.0.1" || cleanIP == "::1" || cleanIP == "localhost"

	if !isLocal {
		// Panggilan pelacakan lokasi IP publik
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

	// Fallback koordinat kota besar penyerang untuk war-room simulation lokal.
	if lat == 0 {
		sources := [][]float64{
			{55.75, 37.61}, {39.90, 116.40}, {38.90, -77.03}, {51.50, -0.12},
			{35.67, 139.65}, {48.85, 2.35}, {52.52, 13.40}, {-23.55, -46.63},
			{-33.86, 151.20}, {1.35, 103.81},
		}
		src := sources[time.Now().UnixNano()%int64(len(sources))]
		lat, lng = src[0], src[1]
	}

	// Lokalisasi titik target Nexus (default: Jakarta)
	targetLat, targetLng := -6.20, 106.81
	domain := strings.ToLower(threatType)
	if strings.Contains(domain, "portal") {
		targetLat, targetLng = 1.35, 103.81 // Target Singapura
	} else if strings.Contains(domain, "audit") {
		targetLat, targetLng = -33.86, 151.20 // Target Sydney
	} else if strings.Contains(domain, "cloud") {
		targetLat, targetLng = 50.11, 8.68 // Target Frankfurt
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

	// Redis Pub/Sub Broadcast
	if mtd.MtdRedis != nil && mtd.MtdRedis.Enabled {
		ctx := context.Background()
		mtd.MtdRedis.Client.Publish(ctx, "nexus:threat_stream", payload)
	}

	// Penyiaran internal asinkron (SSE listener fan-out)
	np.ThreatListeners.Range(func(key, value interface{}) bool {
		ch := value.(chan string)
		select {
		case ch <- msg:
		default:
		}
		return true
	})
}

// routeToHoneypot melakukan NAT silap mata (silent diversion) ke server Honeypot.
func (np *NexusProxy) routeToHoneypot(w http.ResponseWriter, r *http.Request) {
	honeypotURL := "http://localhost:9090"
	target, _ := url.Parse(honeypotURL)
	hp := httputil.NewSingleHostReverseProxy(target)
	r.Host = target.Host
	hp.ServeHTTP(w, r)
}
