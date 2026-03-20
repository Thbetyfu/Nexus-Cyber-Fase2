import requests
import time

def test_gateway():
    url = "http://localhost:8080"
    print("🚀 Starting Nexus Gateway Functional Test...")
    
    # 1. Test Normal Request
    print("\n[TEST 1] Normal Traffic")
    try:
        r = requests.get(url + "/get")
        print(f"Status: {r.status_code}")
        if r.status_code == 200:
            print("✅ Normal traffic passed.")
    except Exception as e:
        print(f"❌ Failed to reach gateway: {e}")
        return

    # 2. Test SQL Injection Detection
    print("\n[TEST 2] SQL Injection Attack Simulation")
    payload = {"id": "1' OR '1'='1' --"}
    r = requests.get(url + "/get", params=payload)
    print(f"URL: {r.url}")
    print(f"Status: {r.status_code}")
    if r.status_code == 403:
        print("🛡️  SUCCESS: SQLi Attempt Blocked by Reflex Layer.")
    else:
        print("❌ FAILED: SQLi Attempt was NOT blocked.")

    # 3. Test XSS Detection
    print("\n[TEST 3] XSS Attack Simulation")
    payload = {"data": "<script>alert('hacked')</script>"}
    r = requests.get(url + "/get", params=payload)
    print(f"Status: {r.status_code}")
    if r.status_code == 403:
        print("🛡️  SUCCESS: XSS Attempt Blocked by Reflex Layer.")
    else:
        print("❌ FAILED: XSS Attempt was NOT blocked.")

if __name__ == "__main__":
    print("💡 Make sure your gateway is running (go run cmd/gateway/main.go)")
    time.sleep(2)
    test_gateway()
