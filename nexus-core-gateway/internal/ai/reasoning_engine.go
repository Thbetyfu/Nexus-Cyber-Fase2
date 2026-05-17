// Package ai menyediakan antarmuka terpadu ke model kecerdasan buatan untuk pertahanan siber aktif.
package ai

import (
	"fmt"
	"os"
)

// ReasoningEngine bertindak sebagai Facade Pattern (Pola Fasad) yang kompatibel ke belakang (backward-compatible)
// untuk mengakses modul Analisis Forensik Mendalam (Reasoning Layer).
//
// Alasan Arsitektural (Why):
// Mengisolasi logika komunikasi OpenRouter/vLLM dari proxy_core.go. Jika backend model AI diubah di masa depan
// (misalnya bermigrasi dari OpenRouter ke model on-premise lokal), kode inti di proxy_core.go tidak perlu
// mengalami modifikasi sama sekali (ISO 25010 - Maintainability & Modularity).
type ReasoningEngine struct {
	client  *LlamaClient
	Enabled bool
}

// NewReasoningEngine mengkonstruksi ReasoningEngine yang didukung oleh LlamaClient (OpenRouter).
// Parameter `url` dan `model` tetap dipertahankan untuk kompatibilitas fungsi dengan pemanggilan warisan (legacy),
// namun secara cerdas melakukan penyesuaian model modern di balik layar.
func NewReasoningEngine(url, model string) *ReasoningEngine {
	// Alasan Teknis (Why):
	// Jika sistem lawas masih meneruskan model default "llama3" atau string kosong, secara otomatis
	// sistem meningkatkan level intelijen ke model "qwen/qwen3-235b-a22b" yang memiliki penalaran forensik
	// jauh lebih tinggi dalam mendeteksi taktik serangan persisten (APT).
	if model == "llama3" || model == "" {
		model = "qwen/qwen3-235b-a22b"
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		// Degradasi Anggun (Fail-Open Policy):
		// Sistem tidak boleh crash jika kunci API kosong, melainkan melanjutkan operasi dengan status
		// degradasi agar kelancaran trafik pengguna (availability) tetap terjaga.
		fmt.Println("[WARN] OPENROUTER_API_KEY not set — Reasoning Layer degraded (Fail-Open).")
	}

	return &ReasoningEngine{
		client:  NewLlamaClient(model),
		Enabled: true,
	}
}

// AnalyzeIntent adalah antarmuka utama yang dipanggil oleh proxy_core.go.
// Menghasilkan analisis forensik terstruktur secara asinkron dari model AI tingkat tinggi.
func (re *ReasoningEngine) AnalyzeIntent(payload string) (*LlamaForensicResult, error) {
	if !re.Enabled {
		return nil, fmt.Errorf("reasoning engine disabled")
	}
	return re.client.AnalyzeIntent(payload)
}
