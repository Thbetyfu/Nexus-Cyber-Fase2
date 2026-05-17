// Package proxy mengimplementasikan gateway proxy reverse otonom dengan kecerdasan MTD.
package proxy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// BrowserIntegrityCheck mengimplementasikan "CGNAT Bypass JS Challenge" (Pemeriksaan Integritas Peramban).
//
// Alasan Arsitektural (Why):
// Sistem perlindungan terhadap bot pemindai scriptless (seperti curl, python-requests, atau zgrab).
// Bot otomatis biasanya tidak mengeksekusi JavaScript. Dengan mengembalikan halaman tantangan matematika ringan
// (Proof-of-Work) yang diselesaikan secara otomatis oleh JavaScript peramban dalam 800ms, kita dapat
// memverifikasi keaslian browser manusia (user-agent integrity) secara transparan tanpa mengganggu kenyamanan pengguna.
func BrowserIntegrityCheck(next http.Handler) http.Handler {
	secret := os.Getenv("NEXUS_SESSION_SECRET")
	if secret == "" {
		secret = "default-matrix-key"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// [LAYER_0_PREFLIGHT_GUARD]
		// Alasan Teknis (Why):
		// Mengizinkan metode OPTIONS untuk bypass pemeriksaan integritas guna mendukung kelancaran komunikasi CORS
		// pada antarmuka admin SOC Dashboard.
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		
		// [LAYER_0_WHITELIST_GUARD]
		// Alasan Teknis (Why):
		// Rute API internal (`/api/`) dibebaskan dari tantangan JS karena dipanggil secara programatik
		// oleh NCC Command Center Desktop/Web. Jika difilter, visualisasi dasbor akan gagal sinkron (blank data).
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/api/verify-session" {
			next.ServeHTTP(w, r)
			return
		}

		// Bypass tantangan untuk jalur sinkronisasi telemetri spesifik.
		if strings.HasPrefix(r.URL.Path, "/api/telemetry") ||
			strings.HasPrefix(r.URL.Path, "/api/ai-events") ||
			strings.HasPrefix(r.URL.Path, "/api/logs") {
			next.ServeHTTP(w, r)
			return
		}

		// 2. Validasi cookie sesi nexus_session
		cookie, err := r.Cookie("nexus_session")
		if err == nil {
			if isValidSession(cookie.Value, secret) {
				next.ServeHTTP(w, r)
				return
			}
		}

		// 3. Kembalikan halaman tantangan (Challenge Page) jika sesi tidak valid/belum terdaftar.
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, generateChallengeHTML(r.URL.Path))
	})
}

// VerifySessionHandler menangani pengiriman jawaban tantangan Proof-of-Work browser.
//
// Alasan Arsitektural (Why):
// - Jika jawaban matematika klien benar (1234 * 5678 = 7006652), sistem menerbitkan cookie otorisasi 24 jam.
// - Cookie diset dengan atribut HttpOnly (mencegah pencurian token via XSS) dan SameSite=Lax (mitigasi CSRF).
func (np *NexusProxy) VerifySessionHandler(w http.ResponseWriter, r *http.Request) {
	secret := os.Getenv("NEXUS_SESSION_SECRET")
	if secret == "" {
		secret = "default-matrix-key"
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	answer := r.FormValue("answer")
	targetPath := r.FormValue("target")
	if targetPath == "" {
		targetPath = "/"
	}

	if answer == "7006652" {
		expiry := time.Now().Add(24 * time.Hour).Unix()
		token := generateToken(expiry, secret)

		http.SetCookie(w, &http.Cookie{
			Name:     "nexus_session",
			Value:    token,
			MaxAge:   86400,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, targetPath, http.StatusFound)
		return
	}

	http.Error(w, "Matrix Verification Failed. Bot Detected.", http.StatusForbidden)
}

// generateToken menyusun token bertanda tangan kriptografi menggunakan HMAC-SHA256.
func generateToken(expiry int64, secret string) string {
	payload := fmt.Sprintf("%d", expiry)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	sig := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return payload + "." + sig
}

// isValidSession memverifikasi integritas tanda tangan token dan masa kedaluarsanya.
func isValidSession(token, secret string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}

	payload, sig := parts[0], parts[1]

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	expectedSig := base64.URLEncoding.EncodeToString(h.Sum(nil))

	if sig != expectedSig {
		return false
	}

	var expiry int64
	fmt.Sscanf(payload, "%d", &expiry)
	return time.Now().Unix() < expiry
}

// generateChallengeHTML menyintesis visual tantangan peramban bertema Matrix minimalis premium.
func generateChallengeHTML(target string) string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>Nexus Cyber | Matrix Verification</title>
    <style>
        body { background: #06080b; color: #10b981; font-family: 'Courier New', monospace; text-align: center; padding-top: 20%; }
        .box { border: 1px solid #10b981; padding: 20px; display: inline-block; border-radius: 8px; background: rgba(16,185,129,0.05); }
        h1 { font-size: 1.2rem; }
        .spinner { border: 4px solid #06080b; border-top: 4px solid #10b981; border-radius: 50%; width: 30px; height: 30px; animation: spin 2s linear infinite; margin: 20px auto; }
        @keyframes spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }
    </style>
</head>
<body>
    <div class="box">
        <h1>VERIFYING TERMINAL INTEGRITY...</h1>
        <p>Bypassing CGNAT via Matrix Sync.</p>
        <div class="spinner"></div>
        <p>Please wait while your browser establishes a secure Nexus Session.</p>
        <form id="challenge" action="/api/verify-session" method="POST" style="display:none;">
            <input type="hidden" name="answer" id="answer">
            <input type="hidden" name="target" value="` + target + `">
        </form>
    </div>
    <script>
        // Eksperimen Proof-of-Work Matematis Sederhana (1234 * 5678)
        setTimeout(() => {
            const res = 1234 * 5678;
            document.getElementById('answer').value = res;
            document.getElementById('challenge').submit();
        }, 800);
    </script>
</body>
</html>`
}
