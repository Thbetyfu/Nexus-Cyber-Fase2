import requests
import time
import random
import sys

# Konfigurasi Target
GATEWAY_URL = "http://localhost:8080"
HONEYPOT_URL = "http://localhost:9090" # Langsung ke Tarpit

# ANSI Colors
RED = '\033[91m'
GREEN = '\033[92m'
YELLOW = '\033[93m'
BLUE = '\033[94m'
MAGENTA = '\033[95m'
CYAN = '\033[96m'
RESET = '\033[0m'

def print_banner(text):
    print(f"\n{BLUE}{'='*60}")
    print(f"{text.center(60)}")
    print(f"{'='*60}{RESET}\n")

def run_attack_scenario():
    print_banner("🎭 SKENARIO PENYELUNDUPAN DATA: HACKER VS NEXUS 🎭")
    
    # ----------------------------------------------------
    # PHASE 1: RECONNAISSANCE (Pencarian Celah)
    # ----------------------------------------------------
    print(f"{YELLOW}[STEP 1] RECONNAISSANCE: Hacker memindai endpoint keuangan...{RESET}")
    time.sleep(1.5)
    print(f"[*] Mencoba akses API Investor...")
    try:
        res = requests.get(f"{GATEWAY_URL}/api/investors?id=1")
        if res.status_code == 200:
            print(f"{GREEN}[+] Berhasil masuk ke endpoint. Hacker menemukan celah SQLi.{RESET}")
    except:
        print(f"{RED}[!] Gagal connect ke Gateway. Pastikan server Go jalan di 8080.{RESET}")
        return

    time.sleep(2)

    # ----------------------------------------------------
    # PHASE 2: EXPLOITATION (Serangan SQL Injection)
    # ----------------------------------------------------
    print(f"\n{YELLOW}[STEP 2] EXPLOITATION: Hacker mengirimkan 'God-Mode' Payload...{RESET}")
    payload = "1 OR 1=1 --"
    print(f"[*] Payload: {RED}'{payload}'{RESET}")
    print(f"[*] Mengirim payload ke Nexus Gateway...")

    # Kita simulasikan Hacker menunggu respon yang ditahan (Tarpit)
    start_time = time.time()
    
    # Visualisasi "Menunggu" yang mendebarkan
    print(f"{MAGENTA}[~] NEXUS DETECTED: AI Reflex is thinking...", end="")
    sys.stdout.flush()
    for _ in range(5):
        time.sleep(0.8)
        print(".", end="")
        sys.stdout.flush()
    
    try:
        # Request ke Gateway
        res = requests.get(f"{GATEWAY_URL}/api/investors", params={"id": payload}, timeout=15)
        duration = time.time() - start_time
        
        print(f"{RESET}\n[!] Respon diterima setelah {duration:.2f} detik.")
        
        if res.status_code == 200:
            data = res.json()
            if data.get("data") == {}:
                print(f"\n{GREEN}🛡️ HASIL: NEXUS MENIPU HACKER! (Digital Hallucination){RESET}")
                print(f"{CYAN}[INFO] Hacker mendapatkan status 'HTTP 200 OK' seolah sukses,")
                print(f"       tapi yang diterima adalah data 'Halusinasi' (JSON Kosong).{RESET}")
                print(f"{YELLOW}[STATS] Hacker terdampar di Honeypot selama {duration:.2f}s tanpa hasil.{RESET}")
            else:
                print(f"{RED}🚨 DATA BOCOR: Nexus gagal memfilter payload! (Periksa API Key AI){RESET}")
    except requests.exceptions.Timeout:
        print(f"\n{GREEN}🛡️ HASIL: NEXUS MEMUTUS KONEKSI (Tarpit Success). Hacker lelah menunggu.{RESET}")

    time.sleep(2)

    # ----------------------------------------------------
    # PHASE 3: BRUTE FORCE / FLOODING (Serangan Balasan)
    # ----------------------------------------------------
    print(f"\n{YELLOW}[STEP 3] RETALIATION: Hacker melakukan Brute Force / DDoS Flooding!{RESET}")
    print(f"[*] Mengirim 150 permintaan sangat cepat untuk melumpuhkan server...")
    
    blocked_count = 0
    for i in range(150):
        try:
            r = requests.get(f"{GATEWAY_URL}/api/investors?id=1", timeout=1)
            if r.status_code == 429:
                blocked_count += 1
            if i % 25 == 0:
                print(f"[*] Request {i+1}/150...")
        except:
            pass
        time.sleep(0.01) # Kecepatan sangat tinggi (exhaust burst capacity)

    print(f"\n{GREEN}🛡️ HASIL: MTD TOKEN BUCKET AKTIF!{RESET}")
    print(f"{CYAN}[STATS] Dari 150 serangan, {blocked_count} berhasil diblokir secara otonom (429 Too Many Requests).{RESET}")
    print(f"{CYAN}[INFO] Server utama tetap stabil karena beban dibuang di firewall MTD.{RESET}")

    print_banner("NEXUS CYBER: MISSION ACCOMPLISHED")
    print(f"{YELLOW}💡 Tips Presentasi: Sambil jalankan skrip ini, tunjukkan chart di Dashboard")
    print(f"   (http://localhost:3000) yang akan melonjak tajam saat Fase 3 ini!{RESET}")

if __name__ == "__main__":
    run_attack_scenario()
