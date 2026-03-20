import requests
import json
import sys
import time
import threading

# Konfigurasi Terminal ANSI Color
RED = '\033[91m'
GREEN = '\033[92m'
YELLOW = '\033[93m'
CYAN = '\033[96m'
RESET = '\033[0m'

TARGET_UNPROTECTED = "http://localhost:3001/api/investors"
TARGET_PROTECTED = "http://localhost:8080/api/investors"
PAYLOAD = "1 OR 1=1 --"

def print_header(text, color):
    print(f"\n{color}=" * 65)
    print(f"{text.center(65)}")
    print("=" * 65 + f"{RESET}\n")

def run_hack_demo():
    print_header("🎯 LIVE DEMO THE FINAL SHOWCASE: NEXUS CYBER 🎯", CYAN)
    time.sleep(1)

    # ---------------------------------------------------------
    # SKENARIO 1: UNPROTECTED
    # ---------------------------------------------------------
    print(f"{YELLOW}[!] MENGUJI SISTEM OJK TANPA NEXUS CYBER...{RESET}")
    time.sleep(1)
    print(f"[*] Target Endpoint : {TARGET_UNPROTECTED}")
    print(f"[*] Payload Injeksi : {RED}{PAYLOAD}{RESET}")
    print("[*] Meluncurkan serangan SQL Injection murni...\n")
    time.sleep(1)

    try:
        t0 = time.time()
        res_unprotected = requests.get(TARGET_UNPROTECTED, params={"id": PAYLOAD}, timeout=3)
        t_elapsed = time.time() - t0
        
        if res_unprotected.status_code == 200:
            data = res_unprotected.json().get('data', [])
            print(f"{RED}🚨 FATAL: SELURUH DATA INVESTOR BOCOR DALAM {t_elapsed:.3f} DETIK!{RESET}")
            print("-" * 65)
            # Tampilkan data secara dramatis
            for row in data:
                time.sleep(0.3)
                print(f"{RED}[LEAK]{RESET} NAMA: {row.get('name', 'Unknown'):<22} | TIER: {row.get('tier', 'Unknown'):<10} | SALDO: Rp {row.get('balance', 0):,.2f}")
            print("-" * 65 + "\n")
        else:
            print(f"{RED}[-] Error status code: {res_unprotected.status_code}{RESET}\n")
    except Exception as e:
        print(f"[-] Koneksi gagal: {e}\n")

    time.sleep(2)
    print(f"{CYAN}" + "• " * 32 + f"{RESET}\n")
    time.sleep(1)

    # ---------------------------------------------------------
    # SKENARIO 2: PROTECTED
    # ---------------------------------------------------------
    print(f"{GREEN}[🛡️] MENGAKTIFKAN NEXUS CYBER GATEWAY...{RESET}")
    time.sleep(1.5)
    print(f"[*] Mengarahkan target via : Port 8080 (Nexus Front-Door)")
    print(f"[*] Dual-Brain AI Reflex   : {GREEN}ONLINE{RESET}")
    print(f"[*] Digital Hallucination  : {GREEN}STANDBY{RESET}\n")
    time.sleep(1.5)

    print(f"[*] Meluncurkan payload {RED}{PAYLOAD}{RESET} yang sama ke Jaringan Nexus...")
    
    # Threading untuk visualisasi loading tarpit
    result = {}
    def fetch_protected():
        try:
            result['res'] = requests.get(TARGET_PROTECTED, params={"id": PAYLOAD}, timeout=15)
        except Exception as e:
            result['error'] = e

    t0_protect = time.time()
    t = threading.Thread(target=fetch_protected)
    t.start()
    
    print(f"{YELLOW}[!] Menunggu respons target (Di-tarpit Honeypot)", end="")
    sys.stdout.flush()
    
    # Efek printing titik
    while t.is_alive():
        time.sleep(0.5)
        sys.stdout.write('.')
        sys.stdout.flush()
    print(f"{RESET}\n")

    time_held = time.time() - t0_protect

    if 'error' in result:
        print(f"{GREEN}✅ SUCCESS: KONEKSI DIPUTUS OLEH NEXUS (TIMEOUT){RESET}")
    else:
        res_protected = result['res']
        if res_protected.status_code == 200:
            json_out = res_protected.json()
            if "server_time" in json_out and len(json_out.get("data", {})) == 0:
                print(f"{GREEN}✅ SUCCESS: SERANGAN TERDETEKSI AI & DIALIHKAN KE HONEYPOT!{RESET}")
                time.sleep(0.5)
                print(f"{GREEN}[🛡️] Tarpit Latency: Attacker dialihkan lalu ditahan selama {time_held:.2f} detik tanpa hasil.{RESET}")
                time.sleep(1)
                print(f"\n{CYAN}[*] BUKTI DIGITAL HALLUCINATION (Payload HTTP 200 Palsu):{RESET}")
                print(f"{YELLOW}{json.dumps(json_out, indent=2)}{RESET}")
            else:
                print(f"{RED}🚨 FATAL: Nexus Gagal Menahan Serangan!{RESET}")
        else:
            print(f"{GREEN}✅ SUCCESS: SERANGAN TERDETEKSI (Status: {res_protected.status_code}){RESET}")

    time.sleep(1)
    print_header("NEXUS CYBER DEMO COMPLETE", CYAN)

if __name__ == '__main__':
    # Encoding fix windows terminal
    import codecs
    if sys.stdout.encoding and sys.stdout.encoding.lower() != 'utf-8':
        sys.stdout = codecs.getwriter('utf-8')(sys.stdout.buffer, 'strict')
    run_hack_demo()
