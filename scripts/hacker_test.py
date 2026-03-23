import requests
import time
import sys

# TARGET: Nexus Gateway
GATEWAY_URL = "http://localhost:8080"
TARGET_DOMAIN = "ojk.localhost"

def print_banner():
    print("--- 🕵️ NEXUS CYBER: RED-TEAM ATTACK SIMULATOR ---")
    print(f"Targeting: {TARGET_DOMAIN} via {GATEWAY_URL}")
    print("Status: Initializing Exploit Payload Stream...\n")

def simulate_attack(name, path, params=None):
    print(f"[ATTACK] Executing {name}...")
    headers = {"Host": TARGET_DOMAIN, "User-Agent": "Mozilla/5.0 (Nexus-Hacker-Tool)"}
    try:
        res = requests.get(f"{GATEWAY_URL}{path}", params=params, headers=headers, timeout=5)
        print(f"  Result: Status {res.status_code}")
        if res.status_code == 403:
            if "Browser Integrity Check" in res.text or "VERIFYING TERMINAL INTEGRITY" in res.text:
                print("  [SHIELD] Blocked by Layer 0: Browser Integrity Check (JS Challenge)")
            else:
                print("  [SHIELD] Blocked by Layer 1: AI Reflex Filter (Deception redirect)")
        elif res.status_code == 404:
            print("  [SHIELD] Blocked by Layer 1: AI Reflex Filter (Deception 404)")
        elif res.status_code == 200:
            print("  [WARNING] Attack bypassed initial filters. Checking AI Reasoning...")
        return res
    except Exception as e:
        print(f"  [ERROR] Connection reset or dropped (Potential Kernel Kill): {e}")

if __name__ == "__main__":
    print_banner()
    
    # 1. SQL Injection (Should trigger Reflex Block + Kernel Drop)
    simulate_attack("SQL Injection", "/api/v1/auth", {"q": "' OR 1=1 --"})
    
    time.sleep(2)
    
    # 2. Path Traversal (Should trigger Reflex Block + Kernel Drop)
    simulate_attack("Path Traversal", "/api/logs", {"file": "../../etc/passwd"})
    
    time.sleep(2)
    
    # 3. Simple Bot/Crawler (Should trigger JS Challenge 403)
    simulate_attack("Scraper Bot", "/index.html")
    
    print("\n--- ATTACK SESSION COMPLETE ---")
    print("[INFO] Check Gateway logs for 'CRITICAL_DEFENSE' or 'OS-FIREWALL' entries.")
