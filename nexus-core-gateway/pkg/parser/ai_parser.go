// Package parser menyediakan modul parsing data tingkat lanjut untuk memproses output dari Large Language Models (LLM).
// Modul ini mematuhi standar ISO 25010 (Sub-karakteristik: Fault Tolerance) untuk menjamin sistem tetap handal
// meskipun respons dari LLM tidak deterministik atau mengandung teks tambahan non-JSON.
package parser

import (
	"encoding/json"
	"fmt"
	"strings"
)

// AIParseError mendefinisikan tipe error kustom ketika seluruh strategi parsing JSON mengalami kegagalan.
// Menyimpan respon mentah (raw response) untuk mempermudah analisis forensik dan debugging di SOC Command Center.
type AIParseError struct {
	RawResponse string
}

// Error mengembalikan representasi string dari AIParseError dengan membatasi panjang cuplikan respon mentah.
// Hal ini mencegah token banjir (log flooding) di console log dan database audit trail.
func (e *AIParseError) Error() string {
	preview := e.RawResponse
	if len(preview) > 200 {
		preview = preview[:200]
	}
	return fmt.Sprintf("ai_parse_error: cannot parse AI response: %s", preview)
}

// ParseAIJSON mengimplementasikan Strategi Parsing JSON 3-Tahap yang sangat toleran terhadap ketidakpastian output LLM.
//
// Alasan Arsitektural (Why):
// Model AI sering kali menyertakan narasi pembuka/penutup, tanda kutip markdown, atau karakter sampah
// di luar format JSON murni. Modul ini menjamin parser tidak langsung crash, melainkan mencoba mengekstrak JSON
// secara progresif guna mempertahankan fungsionalitas sistem (Self-Repairing Logic).
//
// Tahapan:
// - Stage 1 (Direct Parse): Percobaan unmarshal langsung. Paling efisien jika AI mematuhi format dengan sempurna.
// - Stage 2 (Codeblock Extraction): Mencari blok ```json ... ``` yang sering disertakan model instruksi.
// - Stage 3 (Bracket Search): Mencari kurung kurawal pertama '{' hingga penutup '}' terakhir. Paling tangguh untuk mengekstrak JSON dari tengah-tengah percakapan/narasi teks bebas.
func ParseAIJSON(raw string, target interface{}) error {
	raw = strings.TrimSpace(raw)

	// Stage 1: Percobaan unmarshal langsung untuk efisiensi maksimum (Happy Path).
	if err := json.Unmarshal([]byte(raw), target); err == nil {
		return nil
	}

	// Stage 2: Ekstraksi blok kode JSON markdown.
	// Mengatasi kecenderungan LLM yang memformat JSON di dalam blok kode visual markdown.
	if idx := strings.Index(raw, "```json"); idx != -1 {
		end := strings.Index(raw[idx+7:], "```")
		if end != -1 {
			block := strings.TrimSpace(raw[idx+7 : idx+7+end])
			if err := json.Unmarshal([]byte(block), target); err == nil {
				return nil
			}
		}
	}

	// Stage 3: Pencarian Kurung Kurawal (Bracket Search) - Fallback Terkuat.
	// Mengisolasi data JSON dari kebisingan teks bebas di sekitarnya dengan melacak penanda struktur objek { ... }.
	start := strings.Index(raw, "{")
	last := strings.LastIndex(raw, "}")
	if start != -1 && last != -1 && last > start {
		if err := json.Unmarshal([]byte(raw[start:last+1]), target); err == nil {
			return nil
		}
	}

	// Jika ketiga tahap gagal, kembalikan AIParseError untuk penanganan degradasi anggun (graceful degradation).
	return &AIParseError{RawResponse: raw}
}
