import os
import sys

# Ensure UTF-8 output on Windows to avoid UnicodeEncodeError
if sys.platform == "win32":
    import io
    sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

import requests
import time
import random

DOMAINS = ["ojk.localhost", "kemenkeu.localhost", "bi.localhost", "localhost:8080"]
TARGET_URL = "http://localhost:8080"

def generate_traffic():
    print("[NEXUS] Multi-Tenant Traffic Generator Active.")
    print("Simulating traffic for multiple domains...")
    
    while True:
        domain = random.choice(DOMAINS)
        endpoint = random.choice(["/", "/api/investors", "/api/v1/health", "/login"])
        
        headers = {"Host": domain}
        
        # Randomly choose if it's a "Hacker" or "User"
        is_hacker = random.random() < 0.3
        
        try:
            if is_hacker:
                # Malicious payload
                malicious_query = random.choice(["' OR 1=1 --", "<script>alert(1)</script>", "../../etc/passwd"])
                requests.get(f"{TARGET_URL}{endpoint}?query={malicious_query}", headers=headers, timeout=5)
                print(f"[-] ATTACK: Targeted {domain} -> {endpoint} (Flagged)")
            else:
                requests.get(f"{TARGET_URL}{endpoint}", headers=headers, timeout=5)
                print(f"[+] NORMAL: Targeted {domain} -> {endpoint} (Allowed)")
        except Exception as e:
            print(f"[!] Error: {e}")
            
        time.sleep(random.uniform(0.5, 2.0))

if __name__ == "__main__":
    generate_traffic()
