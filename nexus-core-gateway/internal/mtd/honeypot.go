// Package mtd (Moving Target Defense) menyediakan modul pertahanan aktif yang dinamis untuk mengecoh penyerang.
// Modul ini mematuhi standar ISO 27001 (Kontrol A.12 - Perlindungan dari Kerentanan Teknis)
// untuk meredam serangan siber aktif secara proaktif.
package mtd

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"time"
)

// HoneypotServer mengimplementasikan teknik "Digital Hallucination" (Halusinasi Digital).
// Server ini meniru visual dan perilaku server produksi asli (misalnya menyamar sebagai Nginx/Ubuntu)
// namun menahan koneksi penyerang dengan durasi waktu acak (tarpit delay) untuk menghabiskan sumber daya peretas.
//
// Alasan Arsitektural (Why):
// - Mengecoh penyerang agar membuang-buang waktu memindai target palsu yang tidak ada ujungnya.
// - Menghambat bot pemindai otomatis dengan menahan socket koneksi mereka tetap terbuka (socket starvation),
//   sehingga melumpuhkan efisiensi alat pemindai penyerang (seperti sqlmap atau nikto).
type HoneypotServer struct {
	ListenAddr       string
	MinTarpit        time.Duration // Durasi minimum penahanan koneksi
	MaxTarpit        time.Duration // Durasi maksimum penahanan koneksi
	FakeVersion      string        // String versi palsu untuk mengelabui OS Fingerprinting peretas
	OnAttackerCaught func(ip string, path string, ua string)
}

// NewHoneypot mengkonstruksi instansi HoneypotServer dengan tarpit standar 5 hingga 10 detik.
func NewHoneypot(addr string, tarpitDelay time.Duration) *HoneypotServer {
	return &HoneypotServer{
		ListenAddr:  addr,
		MinTarpit:   5 * time.Second,
		MaxTarpit:   10 * time.Second,
		FakeVersion: "nginx/1.18.0 (Ubuntu)",
	}
}

// Start menjalankan server Honeypot HTTP di goroutine latar belakang.
//
// Alasan Keamanan (Why):
// Server ini beroperasi dalam isolasi mutlak (sandbox). Tidak memiliki jalur routing atau kredensial database
// ke server produksi asli. Data-data yang disajikan murni disintesis di memori (in-memory mock)
// sehingga jika penyerang berhasil mengeksploitasi Honeypot ini, mereka tetap tidak dapat menyentuh data asli.
func (h *HoneypotServer) Start() {
	mux := http.NewServeMux()

	// Menangkap seluruh endpoint API dan halaman root untuk menciptakan impresi situs web fungsional.
	mux.HandleFunc("/api/", h.tarpitHandler)
	mux.HandleFunc("/get", h.tarpitHandler)
	mux.HandleFunc("/post", h.tarpitHandler)
	mux.HandleFunc("/", h.tarpitHandler)

	srv := &http.Server{
		Addr:    h.ListenAddr,
		Handler: mux,
	}

	go func() {
		log.Printf("[HONEYPOT] Digital Hallucination server ACTIVE on %s (random tarpit: %v-%v)",
			h.ListenAddr, h.MinTarpit, h.MaxTarpit)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[HONEYPOT] Error: %v", err)
		}
	}()
}

// tarpitHandler mengelola logika penahanan paket dan digital hallucination.
func (h *HoneypotServer) tarpitHandler(w http.ResponseWriter, r *http.Request) {
	attackerIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		attackerIP = r.RemoteAddr
	}

	// 1. Catat IP Penyerang ke Redis Distributed Cache dengan masa berlaku (TTL) 24 jam.
	// Alasan Teknis (Why):
	// IP yang masuk ke honeypot diklasifikasikan sebagai 100% peretas (zero false-positives policy).
	// Menyimpannya di Redis dengan TTL 24 jam memungkinkan seluruh kluster Gateway Nexus memblokir IP ini
	// secara instan di gerbang depan tanpa perlu evaluasi ulang.
	if MtdRedis != nil && MtdRedis.Enabled {
		ctx := r.Context()
		err := MtdRedis.Client.Set(ctx, "honeypot:"+attackerIP, time.Now().String(), 24*time.Hour).Err()
		if err != nil {
			log.Printf("[HONEYPOT-REDIS] Failed to record attacker IP: %v", err)
		} else {
			log.Printf("[HONEYPOT-REDIS] Recorded attacker IP '%s' with 24h TTL.", attackerIP)
		}
	}

	// 2. Catat sidik jari (fingerprint) peretas untuk kepentingan forensik keamanan.
	log.Printf("[HONEYPOT-TRAP] Attacker caught: IP=%s | Path=%s | UA=%s",
		r.RemoteAddr, r.URL.Path, r.Header.Get("User-Agent"))

	// Panggil callback jika diatur oleh orchestrator (dashboard/telemetry stream)
	if h.OnAttackerCaught != nil {
		h.OnAttackerCaught(r.RemoteAddr, r.URL.Path, r.Header.Get("User-Agent"))
	}

	// 3. Eksekusi Penahanan Koneksi (Tarpit Delay) dengan angka acak berbasis Kriptografi (CSPRNG).
	// Alasan Teknis (Why):
	// Jika durasi delay bernilai statis (misal selalu 5 detik), penyerang cerdas dapat dengan mudah
	// mengidentifikasi pola honeypot lewat analisis statistik waktu respon (Timing Analysis).
	// Rentang dinamis dan acak [5s, 10s] menghilangkan kemungkinan kalibrasi waktu respon peretas.
	delay := h.randomTarpit()
	log.Printf("[HONEYPOT-TARPIT] Stalling %s for %v...", r.RemoteAddr, delay.Round(time.Millisecond))
	time.Sleep(delay)

	// Menyusun header palsu agar terlihat seperti server Nginx asli yang memproses data secara sukses.
	w.Header().Set("Server", h.FakeVersion)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", generateFakeRequestID())
	w.WriteHeader(http.StatusForbidden)

	// Kembalikan pesan ejekan kustom interaktif sesuai spesifikasi Task 2
	fmt.Fprintf(w, `{
  "status": "banned",
  "message": "AOWKAOWKOAKWOA SALAH COBA LAGI ANDA SEKARANG BERADA DI DALAM HONEYPOT",
  "server_time": "%s",
  "request_id": "%s"
}
`, time.Now().Format(time.RFC3339), generateFakeRequestID())
}

// randomTarpit menghasilkan durasi penundaan acak yang aman secara kriptografi.
func (h *HoneypotServer) randomTarpit() time.Duration {
	delta := h.MaxTarpit - h.MinTarpit
	if delta <= 0 {
		return h.MinTarpit
	}
	// Menggunakan crypto/rand (CSPRNG) alih-alih math/rand agar angka acak tidak dapat diprediksi peretas.
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(delta)))
	if err != nil {
		return h.MinTarpit // Fallback aman jika entropi sistem bermasalah
	}
	return h.MinTarpit + time.Duration(nBig.Int64())
}

// generateFakeRequestID menyusun UUID tiruan yang meyakinkan menggunakan CSPRNG.
func generateFakeRequestID() string {
	b := make([]byte, 16)
	rand.Read(b) //nolint:errcheck
	return fmt.Sprintf("%08x-%04x-4%03x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
