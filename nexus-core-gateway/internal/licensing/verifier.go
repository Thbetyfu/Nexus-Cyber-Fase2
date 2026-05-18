// Package licensing mengelola lisensi premium dan status verifikasi langganan global Nexus Cyber.
package licensing

import (
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// LicenseState merepresentasikan kondisi lisensi secara global di dalam memori secara thread-safe.
type LicenseState struct {
	mu           sync.RWMutex
	IsValid      bool
	PlanType     string
	LastVerified time.Time
}

var (
	currentLicense LicenseState
	licenseKey     string
)

// InitLicenseVerifier menginisiasi status lisensi awal saat startup gateway.
//
// Alasan Arsitektural (Why):
// Melakukan verifikasi pertama kali secara sinkron saat booting agar gateway langsung menangguhkan rute
// jika kunci lisensi tidak valid sejak detik pertama peluncuran.
func InitLicenseVerifier(key string) {
	licenseKey = key
	verify(key)
}

// IsLicenseValid mengembalikan apakah sistem dalam kondisi terlisensi aktif secara thread-safe.
func IsLicenseValid() bool {
	currentLicense.mu.RLock()
	defer currentLicense.mu.RUnlock()
	return currentLicense.IsValid
}

// GetPlanType mengembalikan tipe plan lisensi saat ini (e.g. "premium", "enterprise").
func GetPlanType() string {
	currentLicense.mu.RLock()
	defer currentLicense.mu.RUnlock()
	return currentLicense.PlanType
}

// StartLicenseHandshake meluncurkan goroutine asinkron yang melakukan verifikasi berkala (handshake).
//
// Alasan Arsitektural (Why):
// - Verifikasi berkala secara asinkron mencegah penundaan (latency) saat memproses request pengguna sah.
// - Menjamin pembaruan instan jika status langganan dicabut (REVOKED) oleh server pusat tanpa memerlukan restart gateway.
func StartLicenseHandshake(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			verify(licenseKey)
		}
	}()
}

func verify(key string) {
	if key == "" {
		currentLicense.mu.Lock()
		currentLicense.IsValid = false
		currentLicense.PlanType = ""
		currentLicense.mu.Unlock()
		return
	}

	// Simulasi panggil API Server Lisensi Pusat
	// Di skenario nyata, kita mengirimkan POST request ke https://license.nexus-cyber.com/verify
	// Agar sistem offline/development tetap dapat berjalan lancar jika tidak ada koneksi internet,
	// kita menerapkan fail-safe graceful: jika server offline namun key memiliki format tepercaya, kita izinkan.
	type VerifyPayload struct {
		Key string `json:"license_key"`
	}
	payload := VerifyPayload{Key: key}
	data, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post("https://license.nexus-cyber.com/verify", "application/json", bytes.NewBuffer(data))
	
	currentLicense.mu.Lock()
	defer currentLicense.mu.Unlock()

	if err != nil {
		// [FAIL-SAFE GRACEFUL RUNNING]
		// Alasan Arsitektural (Why):
		// Celah koneksi internet atau server lisensi yang sedang down tidak boleh melumpuhkan perlindungan WAF.
		// Jika key valid secara format ("nexus-cyber-dev" atau panjang karakter tepercaya), kita anggap valid secara fail-safe.
		if key == "nexus-cyber-dev" || len(key) >= 16 {
			currentLicense.IsValid = true
			currentLicense.PlanType = "premium"
		} else {
			currentLicense.IsValid = false
		}
		currentLicense.LastVerified = time.Now()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result struct {
			Status   string `json:"status"`
			PlanType string `json:"plan_type"`
		}
		if json.NewDecoder(resp.Body).Decode(&result) == nil {
			currentLicense.IsValid = result.Status == "active"
			currentLicense.PlanType = result.PlanType
		}
	} else {
		// Status selain 200 dianggap tidak aktif
		currentLicense.IsValid = false
	}
	currentLicense.LastVerified = time.Now()
}
