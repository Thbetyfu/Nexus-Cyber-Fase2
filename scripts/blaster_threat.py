import requests
import random
import time
import threading

TARGET = "http://localhost:8080"
PAYLOADS = [
    "/?q=1' OR '1'='1",
    "/?cmd=cat /etc/passwd",
    "/?id=<script>alert('pwned')</script>",
    "/wp-admin.php",
    "/config/db.php?env=prod",
    "/?debug=true&exec=rm -rf /",
    "/api/investors?id=1; DROP TABLE users",
    "/api/telemetry?query=OR 1=1 --",
    "/admin-panel/login?user=admin'--",
    "/etc/shadow"
]

def attack():
    while True:
        payload = random.choice(PAYLOADS)
        try:
            # Mengirimkan serangan ke Gateway Nexus
            requests.get(f"{TARGET}{payload}", timeout=5)
            # print(f"[BLAST] Target: {payload}")
        except:
            pass
        # BALANCED MODE: Moderate frequency bursts
        time.sleep(random.uniform(0.3, 1.2))

if __name__ == "__main__":
    print("=" * 60)
    print(" NEXUS CYBER - BALANCED PROTECTION MODE ".center(60))
    print("=" * 60)
    print("[INFO] Initializing Global Monitoring...")
    print("[INFO] Target: http://localhost:8080")
    print("[INFO] Intensity: 5 Concurrent Attack Vectors")
    print("-" * 60)

    threads = []
    for i in range(5):
        t = threading.Thread(target=attack)
        t.daemon = True
        t.start()
        threads.append(t)
    
    print("[SUCCESS] MONITORING STREAM INITIATED. WATCH THE SKY!")
    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        print("\n[SYS] Simulation Terminated.")
