// Package ai mengimplementasikan logika filtrasi cerdas untuk mendeteksi ancaman secara real-time.
package ai

import (
	"regexp"
	"strings"
)

// ReflexFilter mengimplementasikan Filter Refleks Cepat (Phase 1 Heuristics) di bawah arsitektur Hybrid Intelligence.
//
// Alasan Arsitektural (Why):
// Sebelum request dikirim ke model AI (Qwen/Llama) yang memerlukan biaya komputasi tinggi (latency ms/detik),
// ReflexFilter melakukan pemindaian awal dengan Regex yang telah di-kompilasi sebelumnya (pre-compiled).
// Ini menjamin pemblokiran instan (<1ms) untuk signature SQLi/XSS/Traversal klasik, menghemat bandwidth AI,
// serta menjaga performa gateway tetap tinggi (ISO 25010 - Time Behavior & Performance).
type ReflexFilter struct {
	sqliPatterns      []*regexp.Regexp
	xssPatterns       []*regexp.Regexp
	traversalPatterns []*regexp.Regexp
}

// NewReflexFilter menginisialisasi Regex bawaan untuk memindai payload request.
//
// Alasan Arsitektural (Why):
// Pola Regex di-compile di awal (singleton-like initialization) menggunakan regexp.MustCompile.
// Meng-compile regex di setiap request (in-flight request) akan membebani memori dan CPU (CPU spikes),
// sehingga inisialisasi di awal wajib dilakukan untuk stabilitas produksi.
func NewReflexFilter() *ReflexFilter {
	// Pola SQL Injection Umum (Union select, comment bypass, sleep based timing attack, hex encoding)
	sqliRegex := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(UNION|SELECT|INSERT|UPDATE|DELETE|DROP|ALTER|CREATE|TRUNCATE).*`),
		regexp.MustCompile(`(?i)' OR '.*'='`),
		regexp.MustCompile(`(?i)" OR ".*"="`),
		regexp.MustCompile(`(?i)--`),
		regexp.MustCompile(`(?i);`),
		regexp.MustCompile(`(?i)0x[0-9a-fA-F]+`),
		regexp.MustCompile(`(?i)SLEEP\s*\(`),
	}

	// Pola Cross-Site Scripting (XSS) (HTML script injection, element handler hijack, javascript scheme)
	xssRegex := []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script.*?>.*?</script.*?>`),
		regexp.MustCompile(`(?i)on\w+\s*=\s*".*?"`),
		regexp.MustCompile(`(?i)on\w+\s*=\s*'.*?'`),
		regexp.MustCompile(`(?i)javascript:`),
		regexp.MustCompile(`(?i)alert\s*\(`),
		regexp.MustCompile(`(?i)document\.cookie`),
	}

	// Pola Path Traversal & File Access (LFI/RFI)
	traversalRegex := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\.{2,}[/\\]`), // Mendeteksi ../ atau ..\ (Directory traversal)
		regexp.MustCompile(`(?i)%2e%2e%2f`),   // Mendeteksi URL encoded traversal
		regexp.MustCompile(`(?i)/etc/passwd`), // Lokasi sensitif Linux
		regexp.MustCompile(`(?i)/etc/shadow`), 
		regexp.MustCompile(`(?i)C:\\`),        // Drive utama Windows
		regexp.MustCompile(`(?i)win\.ini`),     
	}

	return &ReflexFilter{
		sqliPatterns:      sqliRegex,
		xssPatterns:       xssRegex,
		traversalPatterns: traversalRegex,
	}
}

// InspectRequest memindai string input untuk mencari pola eksploitasi dasar.
// Mengembalikan status ancaman (bool) dan jenis ancaman jika terdeteksi.
func (f *ReflexFilter) InspectRequest(data string) (isThreat bool, threatType string) {
	// Konversi input ke lowercase sekali saja untuk menghemat CPU dibanding mencocokkan case-insensitive berulang kali.
	data = strings.ToLower(data)

	// 1. Pemindaian SQLi
	for _, p := range f.sqliPatterns {
		if p.MatchString(data) {
			return true, "SQL_INJECTION_DETECTED"
		}
	}

	// 2. Pemindaian Path Traversal
	for _, p := range f.traversalPatterns {
		if p.MatchString(data) {
			return true, "PATH_TRAVERSAL_DETECTED"
		}
	}

	// 3. Pemindaian XSS
	for _, p := range f.xssPatterns {
		if p.MatchString(data) {
			return true, "XSS_DETECTED"
		}
	}

	return false, ""
}

// InspectAdvanced memperluas analisis dengan memverifikasi User-Agent untuk mendeteksi bot/scanner siber.
//
// Alasan Arsitektural (Why):
// Banyak peretas menggunakan alat pemindai otomatis (seperti sqlmap atau nikto) untuk mencari celah keamanan.
// Mengidentifikasi header User-Agent pemindai di awal memungkinkan gateway langsung memblokir request
// sebelum program peretas sempat mengirim payload eksploitasi sesungguhnya.
func (f *ReflexFilter) InspectAdvanced(data string, ua string) (isThreat bool, threatType string) {
	// 1. Deteksi Alat Pemindai / Rekognisi Otomatis
	ua = strings.ToLower(ua)
	scanners := []string{"sqlmap", "gobuster", "dirb", "nmap", "nikto", "burp", "zap", "acunetix"}
	for _, s := range scanners {
		if strings.Contains(ua, s) {
			return true, "MALICIOUS_SCANNER_TOOL_DETECTED"
		}
	}

	// 2. Jika bukan scanner, lanjutkan ke pemindaian payload standar.
	return f.InspectRequest(data)
}

// Sanitize membersihkan karakter berbahaya secara lokal (Opsional, untuk pengembangan fitur berikutnya).
func (f *ReflexFilter) Sanitize(data string) string {
	return strings.ReplaceAll(data, "<", "&lt;")
}
