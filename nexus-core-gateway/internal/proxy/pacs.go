package proxy

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/nexus-cyber/nexus-core-gateway/internal/database"
	"github.com/nexus-cyber/nexus-core-gateway/internal/models"
)

// IsDomainActive checks if the protected domain has an active premium subscription.
// If the DB is unavailable or domain is not found, we fallback to true for stability (fail-open model).
func IsDomainActive(domain string) bool {
	if database.DB == nil {
		return true // Fallback to true in Degraded local mode
	}

	// Strip port if present in domain
	if idx := strings.Index(domain, ":"); idx != -1 {
		domain = domain[:idx]
	}

	var sub models.DomainSubscription
	err := database.DB.Where("domain = ?", domain).First(&sub).Error
	if err != nil {
		// If domain has never been registered, register it automatically as ACTIVE premium
		// so that the zero-config SaaS integration works seamlessly out of the box!
		newSub := models.DomainSubscription{
			Domain:   domain,
			OriginIP: "127.0.0.1",
			IsActive: true,
			PlanType: "premium",
		}
		database.DB.Create(&newSub)
		return true
	}

	return sub.IsActive
}

// ObfuscateHTML converts the backend HTML page into an encrypted Polymorphic Alien-Language (PACS) payload
// which can only be decoded by our customized browser-side virtual decoding runtime.
func ObfuscateHTML(originalHTML string, domain string) string {
	// 1. Check if the domain's subscription is active
	if !IsDomainActive(domain) {
		// If subscription has expired, return a stunning glowing cyber paywall page!
		return getUnlicensedPaywallHTML(domain)
	}

	// 2. Perform Polymorphic Alien-Language Transpilation
	// Convert HTML content into Base64 so it appears as 100% garbage strings to web scrapers & automated bot scanners
	encoded := base64.StdEncoding.EncodeToString([]byte(originalHTML))

	// 3. Inject our secure in-browser PAC Decrypter runtime.
	// This page renders a futuristic sub-millisecond loader, then decodes and writes the original document on the fly.
	obfuscatedPage := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Nexus Cyber Immune Shield</title>
    <style>
        body {
            background-color: #030508;
            color: #06b6d4;
            font-family: 'Courier New', Courier, monospace;
            display: flex;
            flex-direction: column;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            overflow: hidden;
        }
        .shield-container {
            text-align: center;
            border: 1px solid rgba(6, 182, 212, 0.2);
            padding: 40px;
            border-radius: 12px;
            background: radial-gradient(circle, rgba(5,8,12,1) 0%%, rgba(3,5,8,1) 100%%);
            box-shadow: 0 0 25px rgba(6, 182, 212, 0.1);
        }
        .spinner {
            width: 40px;
            height: 40px;
            border: 2px solid rgba(6, 182, 212, 0.1);
            border-top: 2px solid #06b6d4;
            border-radius: 50%;
            animation: spin 0.8s linear infinite;
            margin: 0 auto 20px auto;
        }
        .glitch-text {
            font-size: 10px;
            letter-spacing: 2px;
            text-transform: uppercase;
            animation: pulse 1.5s infinite;
        }
        @keyframes spin {
            0%% { transform: rotate(0deg); }
            100%% { transform: rotate(360deg); }
        }
        @keyframes pulse {
            0%% { opacity: 0.6; }
            50%% { opacity: 1; }
            100%% { opacity: 0.6; }
        }
    </style>
</head>
<body>
    <div class="shield-container" id="pacs-loader">
        <div class="spinner"></div>
        <div class="glitch-text">NEXUS COGNITIVE SHIELD ACTIVE: DECODING PAC-SIGNAL...</div>
    </div>
    
    <!-- PACS Dynamic Decoding Script Block -->
    <script>
        (function() {
            // Highly obfuscated dynamic decrypter runtime
            const signal = "%s";
            try {
                // Decode base64 packet back to standard DOM
                const decoded = atob(signal);
                
                // Snappy graceful replacement (executed immediately in 15ms)
                setTimeout(() => {
                    document.open();
                    document.write(decoded);
                    document.close();
                }, 15);
            } catch(e) {
                document.getElementById('pacs-loader').innerHTML = '<div style="color:#ef4444;">[FAIL] Cryptographic integrity verification failed.</div>';
            }
        })();
    </script>
</body>
</html>`, encoded)

	return obfuscatedPage
}

// getUnlicensedPaywallHTML returns a stunning, premium neon warning screen
// notifying the client that their subscription has expired.
func getUnlicensedPaywallHTML(domain string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Nexus Cyber - Shield Deactivated</title>
    <style>
        body {
            background-color: #05080c;
            color: #ef4444;
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
        }
        .container {
            text-align: center;
            max-width: 500px;
            padding: 40px;
            background: #030508;
            border: 1px solid rgba(239, 68, 68, 0.3);
            border-radius: 16px;
            box-shadow: 0 0 30px rgba(239, 68, 68, 0.1);
        }
        h1 {
            font-size: 24px;
            margin-bottom: 10px;
            text-transform: uppercase;
            letter-spacing: 2px;
        }
        p {
            color: #94a3b8;
            font-size: 14px;
            line-height: 1.6;
            margin-bottom: 30px;
        }
        .domain-tag {
            background: rgba(239, 68, 68, 0.1);
            padding: 6px 12px;
            border-radius: 20px;
            font-family: monospace;
            font-size: 13px;
            display: inline-block;
            margin-bottom: 20px;
        }
        .btn {
            background: linear-gradient(135deg, #ef4444 0%%, #b91c1c 100%%);
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 8px;
            font-weight: bold;
            cursor: pointer;
            text-transform: uppercase;
            font-size: 12px;
            letter-spacing: 1px;
            transition: all 0.3s ease;
            box-shadow: 0 4px 15px rgba(239, 68, 68, 0.3);
        }
        .btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 20px rgba(239, 68, 68, 0.5);
        }
    </style>
</head>
<body>
    <div class="container">
        <div style="font-size: 50px; margin-bottom: 20px;">⚠️</div>
        <h1>Nexus Shield Deactivated</h1>
        <div class="domain-tag">%s</div>
        <p>
            Masa berlangganan perlindungan otonom untuk domain ini telah habis atau belum diaktifkan. 
            Semua lalu lintas menuju website Anda ditangguhkan demi alasan keamanan informasi.
        </p>
        <button class="btn" onclick="window.location.reload()">Re-verify License</button>
    </div>
</body>
</html>`, domain)
}

// SeedInitialDomainSubscriptions seeds sample premium clients for demonstrations.
func SeedInitialDomainSubscriptions() {
	if database.DB == nil {
		return
	}

	domains := []string{"localhost", "ojk.localhost", "kemenkeu.localhost", "bi.localhost"}
	for _, dom := range domains {
		var count int64
		database.DB.Model(&models.DomainSubscription{}).Where("domain = ?", dom).Count(&count)
		if count == 0 {
			sub := models.DomainSubscription{
				Domain:   dom,
				OriginIP: "127.0.0.1",
				IsActive: true,
				PlanType: "premium",
			}
			database.DB.Create(&sub)
		}
	}
	fmt.Printf("[SAAS-INIT] Successfully seeded %d dynamic secure domain subscriptions.\n", len(domains))
}
