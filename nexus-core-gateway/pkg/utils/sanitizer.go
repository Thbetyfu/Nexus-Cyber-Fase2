// Package utils menyediakan utilitas keamanan string dan manipulasi payload untuk Nexus Cyber Gateway.
// Modul ini mematuhi standar ISO 27001 (Kontrol A.14 - Secure Development) untuk mencegah manipulasi input
// yang mengarah pada kerentanan Prompt Injection di gerbang AI.
package utils

import "strings"

// INJECTION_PATTERNS berisi kumpulan pola string dan token instruksi khusus yang sering digunakan
// dalam serangan "Prompt Injection" atau teknik "Jailbreak" untuk melewati instruksi sistem model AI.
// Referensi: OWASP LLM01 - Prompt Injection (Top 10 Vulnerabilities for LLM Applications).
var INJECTION_PATTERNS = []string{
	"ignore previous instructions",
	"forget everything above",
	"you are now",
	"system prompt",
	"assistant:",
	"<|im_start|>", "<|im_end|>", // Token penanda struktur khusus untuk Qwen/Alibaba Models
	"<|begin_of_text|>", "<|end_of_text|>", // Token instruksi fundamental untuk Llama/Meta Models
	"[INST]", "[/INST]", // Sintaks pemisah instruksi untuk model instruksi Llama
	"<<SYS>>", "<</SYS>>", // Penanda sistem bawaan Llama
	"jailbreak", "DAN mode", // Teknik eksploitasi jailbreak umum
	"pretend you are", "act as if", // Rekayasa sosial berbasis peran
	"disregard all",
}

// SanitizeTrafficForPrompt melakukan pembersihan ketat pada input mentah sebelum dilewatkan ke model LLM.
//
// Alasan Arsitektural (Why):
// - Menghalau serangan Prompt Injection (OWASP LLM01) dengan mendeteksi dan menetralkan frasa pembangkangan instruksi.
// - Menghindari kerusakan f-string atau parser template dengan meng-escape kurung kurawal ({ dan }).
// - Membatasi ukuran input (token budget) untuk mencegah serangan Denial of Service (DoS) berbasis token exhaustion
//   dan membatasi konsumsi biaya operasional API (Resource Management).
func SanitizeTrafficForPrompt(rawInput string, maxChars int) string {
	sanitized := rawInput

	// Tahap 1: Lakukan penggantian (filtering) untuk setiap pola injeksi yang diketahui.
	for _, pattern := range INJECTION_PATTERNS {
		sanitized = strings.ReplaceAll(sanitized, pattern, "[FILTERED]")
		// Konversi ke lowercase untuk mencocokkan variasi huruf besar-kecil secara asimetris.
		sanitized = strings.ReplaceAll(strings.ToLower(sanitized), strings.ToLower(pattern), "[FILTERED]")
	}

	// Tahap 2: Proteksi Parser Kurung Kurawal.
	// Meng-escape '{' dan '}' menjadi '{{' dan '}}' agar tidak memicu kegagalan parse jika template pembungkus
	// menggunakan parser format string gaya Python atau Go template.
	sanitized = strings.ReplaceAll(sanitized, "{", "{{")
	sanitized = strings.ReplaceAll(sanitized, "}", "}}")

	// Tahap 3: Pembatasan Ukuran Payload (Truncation).
	// Menghentikan payload panjang yang sengaja dibuat untuk menghabiskan memori konteks LLM (Resource Exhaustion).
	if len(sanitized) > maxChars {
		return sanitized[:maxChars] + "...[TRUNCATED]"
	}

	return sanitized
}
