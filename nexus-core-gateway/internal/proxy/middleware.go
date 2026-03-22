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

// BrowserIntegrityCheck implements the CGNAT Bypass JS Challenge.
// It leverages a simple math-based Proof-of-Work to differentiate human
// browsers from scriptless bots, issuing a 24h HMAC session cookie.
func BrowserIntegrityCheck(next http.Handler) http.Handler {
	secret := os.Getenv("NEXUS_SESSION_SECRET")
	if secret == "" {
		secret = "default-matrix-key"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Skip challenge for API/Telemetry and Honeypot
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/api/verify-session" {
			next.ServeHTTP(w, r)
			return
		}

		// 2. Validate nexus_session cookie
		cookie, err := r.Cookie("nexus_session")
		if err == nil {
			if isValidSession(cookie.Value, secret) {
				next.ServeHTTP(w, r)
				return
			}
		}

		// 3. Return Challenge Page
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, generateChallengeHTML(r.URL.Path))
	})
}

// verifySessionHandler handles the challenge submission
func (np *NexusProxy) VerifySessionHandler(w http.ResponseWriter, r *http.Request) {
	secret := os.Getenv("NEXUS_SESSION_SECRET")
	if secret == "" {
		secret = "default-matrix-key"
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Simple check: client sends the answer to (1234 * 5678) = 7006652
	answer := r.FormValue("answer")
	targetPath := r.FormValue("target")
	if targetPath == "" {
		targetPath = "/"
	}

	if answer == "7006652" {
		// Issue session
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

func generateToken(expiry int64, secret string) string {
	payload := fmt.Sprintf("%d", expiry)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	sig := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return payload + "." + sig
}

func isValidSession(token, secret string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}

	payload, sig := parts[0], parts[1]

	// Verify signature
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	expectedSig := base64.URLEncoding.EncodeToString(h.Sum(nil))

	if sig != expectedSig {
		return false
	}

	// Verify expiry (Phase 7 accuracy)
	var expiry int64
	fmt.Sscanf(payload, "%d", &expiry)
	return time.Now().Unix() < expiry
}

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
        // Simple Proof-of-Work Challenge
        // Math: 1234 * 5678
        setTimeout(() => {
            const res = 1234 * 5678;
            document.getElementById('answer').value = res;
            document.getElementById('challenge').submit();
        }, 800);
    </script>
</body>
</html>`
}
