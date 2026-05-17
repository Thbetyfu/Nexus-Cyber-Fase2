// Package proxy mengimplementasikan gateway proxy reverse otonom dengan kecerdasan MTD.
package proxy

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/nexus-cyber/nexus-core-gateway/internal/mtd"
)

// RouteEntry menyimpan data alamat target backend beserta masa kadaluarsanya di memori lokal.
type RouteEntry struct {
	TargetURL string
	ExpiresAt time.Time
}

// DynamicRouter mengelola pemetaan domain (host) ke backend secara dinamis berbasis Redis + Cache Memori Lokal.
//
// Alasan Arsitektural (Why):
// Modul ini mematuhi standar ISO 25010 (Time Behavior & Resource Utilization).
// Melakukan query ke Redis terdistribusi untuk setiap HTTP request yang masuk akan membebani I/O jaringan
// dan meningkatkan latensi gerbang secara signifikan. Router ini menerapkan arsitektur caching dua tingkat:
// - Tier 1: RAM lokal (In-Memory Cache) berlatency sub-mikrodetik menggunakan RWMutex.
// - Tier 2: Penyimpanan terdistribusi Redis Hash untuk sinkronisasi antar-node gateway.
type DynamicRouter struct {
	cache map[string]RouteEntry
	mu    sync.RWMutex // Lock baca/tulis yang dioptimalkan untuk performa tinggi pada konkurensi tinggi
	ttl   time.Duration // Durasi hidup cache memori lokal sebelum divalidasi ulang ke Redis
}

// NewDynamicRouter membuat instansi router dinamis baru dengan TTL cache lokal tertentu.
func NewDynamicRouter(cacheTTL time.Duration) *DynamicRouter {
	return &DynamicRouter{
		cache: make(map[string]RouteEntry),
		ttl:   cacheTTL,
	}
}

// Lookup menemukan URL target backend berdasarkan nama domain (host).
//
// Alasan Teknis (Why):
// 1. Menguji cache memori lokal dengan Read-Lock (`RLock`) terlebih dahulu.
//    Read-Lock memungkinkan ratusan goroutine membaca cache secara simultan tanpa saling memblokir (high concurrency).
// 2. Jika kadaluarsa atau tidak ditemukan, sistem melakukan fallback ke Redis dengan batas waktu ketat (`500ms timeout`).
//    Batas waktu ini memastikan jika Redis mengalami kemacetan, gateway tidak ikut macet dan tetap responsif.
// 3. Hasil dari Redis disimpan secara malas (lazy population) ke memori lokal untuk mempercepat pencarian berikutnya.
func (dr *DynamicRouter) Lookup(host string) (string, bool) {
	// 1. Periksa Cache RAM Lokal (Tier 1 - Kinerja Ekstrem)
	dr.mu.RLock()
	entry, exists := dr.cache[host]
	dr.mu.RUnlock()

	if exists && time.Now().Before(entry.ExpiresAt) {
		return entry.TargetURL, true
	}

	// 2. Fallback ke Redis terdistribusi (Tier 2 - Sinkronisasi Global)
	if mtd.MtdRedis != nil && mtd.MtdRedis.Enabled {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		target, err := mtd.MtdRedis.Client.HGet(ctx, "nexus:routes", host).Result()
		if err == nil && target != "" {
			// Perbarui cache memori lokal menggunakan Write-Lock (`Lock`)
			dr.mu.Lock()
			dr.cache[host] = RouteEntry{
				TargetURL: target,
				ExpiresAt: time.Now().Add(dr.ttl),
			}
			dr.mu.Unlock()
			return target, true
		}
	}

	return "", false
}

// AddRoute mendaftarkan pemetaan rute baru secara instan di memori lokal dan menyinkronkannya ke Redis.
//
// Alasan Teknis (Why):
// Menggunakan Write-Lock (`Lock`) eksklusif untuk menghindari kondisi balapan data (race condition)
// sewaktu memperbarui map cache memori. Sinkronisasi ke Redis dilakukan menggunakan batasan timeout 2 detik.
func (dr *DynamicRouter) AddRoute(host, target string) error {
	// 1. Perbarui Cache Memori Lokal (Penyediaan instan untuk request lokal)
	dr.mu.Lock()
	dr.cache[host] = RouteEntry{
		TargetURL: target,
		ExpiresAt: time.Now().Add(dr.ttl),
	}
	dr.mu.Unlock()

	// 2. Sinkronisasi Global ke Redis (Persistensi kluster)
	if mtd.MtdRedis != nil && mtd.MtdRedis.Enabled {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := mtd.MtdRedis.Client.HSet(ctx, "nexus:routes", host, target).Err()
		if err != nil {
			return err
		}
	}

	log.Printf("[ROUTER] Mapping established: %s -> %s", host, target)
	return nil
}

// GetAllRoutes menarik seluruh data pemetaan rute yang aktif dari Redis terdistribusi.
func (dr *DynamicRouter) GetAllRoutes() (map[string]string, error) {
	if mtd.MtdRedis != nil && mtd.MtdRedis.Enabled {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		return mtd.MtdRedis.Client.HGetAll(ctx, "nexus:routes").Result()
	}
	return nil, nil
}
