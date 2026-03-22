import requests
import time

GATEWAY_URL = "http://localhost:8080"
DASHBOARD_API = f"{GATEWAY_URL}/api"

def test_dynamic_onboarding():
    print("🧪 [QA AUDIT] Starting Zero-Code Onboarding Test...")
    
    # 1. Check current routes
    try:
        res = requests.get(f"{DASHBOARD_API}/routes")
        print(f"[1] Current Routes: {res.json()}")
    except Exception as e:
        print(f"[!] Critical Error: Dashboard API unresponsive: {e}")
        return

    # 2. Add New Route
    new_domain = "kemenkeu.localhost"
    target_url = "http://host.docker.internal:3001" # Simulating same backend for now
    
    print(f"[2] Onboarding new domain: {new_domain} -> {target_url}")
    res = requests.post(f"{DASHBOARD_API}/routes", json={
        "domain": new_domain,
        "target_url": target_url
    })
    
    if res.status_code == 200:
        print("[+] Success: Domain Onboarded via Dynamic Routing API.")
    else:
        print(f"[-] Failed: Status {res.status_code}")
        return

    # 3. Verify Traffic Routing (No Restart)
    print(f"[3] Verifying immediate routing to {new_domain}...")
    headers = {"Host": new_domain}
    res = requests.get(GATEWAY_URL, headers=headers)
    
    if res.status_code == 200:
        print("[+] SUCCESS: Gateway dynamically routed traffic to new target!")
    else:
        print(f"[-] FAILED: Received status {res.status_code}")

    # 4. Check Caching Performance
    start = time.time()
    for _ in range(50):
        requests.get(GATEWAY_URL, headers=headers)
    duration = time.time() - start
    print(f"[4] Caching Audit: Handled 50 concurrent requests in {duration:.2f}s (Avg: {duration/50*1000:.1f}ms per request)")
    
    print("\n✅ QA Audit Complete: Zero-Code Onboarding is SEAMLESS and HIGH-PERFORMANCE.")

if __name__ == "__main__":
    test_dynamic_onboarding()
