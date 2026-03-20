import requests
import time
import concurrent.futures

BASE_URL = "http://localhost:8080"

def test_attack(name, payload, method="GET"):
    try:
        r = requests.request(method, BASE_URL + "/get", params=payload if method == "GET" else None, data=payload if method == "POST" else None)
        status = "BLOCKED (403)" if r.status_code == 403 else "BYPASSED (200)"
        print(f"[{name:.<25}] Payload: {str(payload)[:50]}... | Result: {status}")
        return r.status_code == 403
    except Exception as e:
        print(f"❌ Connection error for {name}: {e}")
        return False

def run_battle_test():
    print("🔥 NEXUS CYBER - ADVANCED BATTLE TEST INITIATED")
    print("-" * 60)
    
    results = []
    
    # 1. Obfuscated XSS (Hex/Encoding)
    results.append(test_attack("Obfuscated XSS (Hex)", {"id": "%3cscript%3ealert(1)%3c/script%3e"}))
    
    # 2. Case Variation
    results.append(test_attack("Case Variation XSS", {"data": "<sCrIpT>alert(1)</ScRiPt>"}))
    
    # 3. Logical SQL Bypass (Comments)
    results.append(test_attack("SQL Comment Bypass", {"q": "SEL/**/ECT * FR/**/OM users"}))
    
    # 4. Unicode Variation
    results.append(test_attack("Unicode XSS", {"u": "\u003cscript\u003ealert(1)\u003c/script\u003e"}))

    # 5. POST Request Bypass (Stress/Payload size)
    results.append(test_attack("POST Payload Check", {"body": "OR 1=1 --"}, method="POST"))

    print("-" * 60)
    print("⚡ STRESS TEST: Sending 100 requests in parallel...")
    start_time = time.time()
    with concurrent.futures.ThreadPoolExecutor(max_workers=50) as executor:
        futures = [executor.submit(test_attack, f"Stress_{i}", {"t": i}) for i in range(100)]
        concurrent.futures.wait(futures)
    
    duration = time.time() - start_time
    print("-" * 60)
    print(f"📊 SUMMARY: {results.count(True)} Blocked | {results.count(False)} Bypassed")
    print(f"⏱️  Stress Test Duration: {duration:.2f} seconds.")
    print("-" * 40)

if __name__ == "__main__":
    run_battle_test()
