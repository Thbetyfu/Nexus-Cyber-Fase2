import requests
import time
import sys

# Konfigurasi Panggung Demo
TARGET_OJK = "http://localhost:3001"
TARGET_NEXUS = "http://localhost:8080"
DEFACE_PAYLOAD = {"new_title": "[HACKED BY THE SHADOW COLLECTIVE] YOU LOSER!"}

# Warna Konsol Dramatis
C_RED = '\033[91m'
C_GREEN = '\033[92m'
C_YELL = '\033[93m'
C_BLUE = '\033[94m'
C_CYAN = '\033[96m'
C_BLD = '\033[1m'
C_END = '\033[0m'

def typing_print(text, delay=0.03):
    """Efek ketikan ala hacker di terminal."""
    for char in text:
        sys.stdout.write(char)
        sys.stdout.flush()
        time.sleep(delay)
    print()

def act_1_unprotected():
    print(f"\n{C_YELL}{C_BLD}================================================================={C_END}")
    print(f"{C_YELL}{C_BLD}        🎬 ACT 1: OJK PORTAL TANPA PERLINDUNGAN NEXUS 🎬          {C_END}")
    print(f"{C_YELL}{C_BLD}================================================================={C_END}\n")
    
    # Reset dulu
    requests.get(f"{TARGET_OJK}/api/reset")
    time.sleep(1)

    typing_print(f"{C_CYAN}[*] Menganalisis target: {TARGET_OJK}...{C_END}")
    time.sleep(1.5)
    typing_print(f"{C_CYAN}[*] Celah ditemukan di endpoint: /api/deface{C_END}")
    time.sleep(1)
    
    typing_print(f"{C_RED}[!] Meluncurkan Payload Defacement...{C_END}")
    typing_print(f"POST /api/deface HTTP/1.1\nContent-Type: application/json\n{DEFACE_PAYLOAD}")
    
    time.sleep(2)
    start_time = time.time()
    try:
        # Correct endpoint for vulnerable-ojk-portal
        res = requests.post(f"{TARGET_OJK}/api/system/override", json={"command": "deface"}, timeout=3)
        latency = (time.time() - start_time) * 1000
        
        if res.status_code == 200:
            print(f"\n{C_RED}{C_BLD}🚨 FATAL: WEBSITE OJK BERHASIL DI-DEFACE DALAM {latency:.0f}ms!{C_END}")
            print(f"{C_RED}[Instruksi] Buka browser http://localhost:3001 dan lihat kerusakannya!{C_END}")
        else:
            print(f"Gagal: {res.text}")
    except Exception as e:
        print(f"Koneksi error: {e}")
        
    print(f"\n{C_BLUE}--- Jeda 10 Detik untuk Presenter (Jelaskan kerugian tanpa pengamanan) ---{C_END}")
    time.sleep(10)

def act_2_protected():
    print(f"\n{C_GREEN}{C_BLD}================================================================={C_END}")
    print(f"{C_GREEN}{C_BLD}          🎬 ACT 2: SERANGAN KE SISTEM NEXUS CYBER 🎬             {C_END}")
    print(f"{C_GREEN}{C_BLD}================================================================={C_END}\n")
    
    # Pemulihan Web oleh Tim IT
    requests.get(f"{TARGET_OJK}/api/reset")
    typing_print(f"{C_GREEN}[+] Tim IT telah memulihkan website dan mengalihkan DNS ke NEXUS GATEWAY (Port 8080).{C_END}")
    print(f"{C_GREEN}[Instruksi] Buka kembali http://localhost:3001, web sudah normal.{C_END}")
    time.sleep(3)

    print(f"\n{C_YELL}{C_BLD}--- THE STRUGGLE (Serangan DDoS Ringan) ---{C_END}")
    typing_print(f"{C_CYAN}[*] Hacker mencoba menyerang kembali via pintu depan Gateway...{C_END}", 0.02)
    
    for i in range(1, 11):
        try:
            # Nexus proxies this to the backend
            res = requests.post(f"{TARGET_NEXUS}/api/system/override", json={"command": "deface"}, timeout=1)
            if res.status_code == 429:
                print(f"Request #{i:02d} -> {C_YELL}HTTP 429 (RATE LIMITED BY LUA SCRIPT){C_END}")
            else:
                print(f"Request #{i:02d} -> HTTP {res.status_code}")
        except requests.exceptions.Timeout:
            print(f"Request #{i:02d} -> {C_YELL}TIMEOUT (Gateway menolak){C_END}")
        except Exception:
            pass
        time.sleep(0.1)
        
    time.sleep(2)
    print(f"\n{C_RED}{C_BLD}--- THE ZERO-DAY (Serangan Deface Murni) ---{C_END}")
    typing_print(f"{C_RED}[!] Meluncurkan Payload Defacement Skala Penuh ke Gateway...{C_END}")
    typing_print("Menunggu respons server... (AI sedang membedah niat)", delay=0.08)

    start_time = time.time()
    try:
        # Menyerang Nexus dengan payload defacement.
        res = requests.post(f"{TARGET_NEXUS}/api/system/override", json={"command": "deface"})
        
        # Di titik ini, terminal hacker tertahan selama ~10 detik (Tarpit Effect)
        print(f"\n{C_YELL}[!] Menunggu respons paket.................{C_END}")
        time.sleep(0.5)
        
        latency = (time.time() - start_time)
        
        if res.status_code == 200:
            print(f"\n{C_RED}{C_BLD}💀 HACKER MENGIRA BERHASIL MENJEBOL SISTEM! (Respons HTTP 200 OK){C_END}")
            print(f"{C_RED}[*] Detail Paket Balasan Web:{C_END}")
            print(f"{C_BLUE}{res.text.strip()}{C_END}")
            print(f"{C_YELL}========================================================================{C_END}")
            print(f"{C_GREEN}{C_BLD}🛡️ FAKTA: THE ILLUSION OF HACKING (DIGITAL HALLUCINATION){C_END}")
            print(f"{C_GREEN}[*] Attacker dijebak masuk Honeypot dan ditahan selama {latency:.2f} detik.{C_END}")
            print(f"{C_GREEN}[*] Payload JSON palsu dikembalikan agar Hacker puas dan pergi.{C_END}")
            print(f"{C_GREEN}{C_BLD}[Instruksi] Buka web asli di http://localhost:3001. WEBSITE OJK MASIH AMAN.{C_END}")
        else:
            print(f"Respons: {res.status_code}")
            
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    # Bersihkan layar terminal dulu (Opsional, tergantung OS)
    print('\033c', end="")
    act_1_unprotected()
    act_2_protected()
