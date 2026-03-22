import requests
import time
import sys
import argparse

# Warna Konsol Dramatis
C_RED = '\033[91m'
C_GREEN = '\033[92m'
C_YELL = '\033[93m'
C_CYAN = '\033[96m'
C_BLD = '\033[1m'
C_END = '\033[0m'

def typing_print(text, delay=0.04):
    """Efek ketikan ala hacker di terminal."""
    for char in text:
        sys.stdout.write(char)
        sys.stdout.flush()
        time.sleep(delay)
    print()

def attack(target_url, message="HACKED BY THE SHADOW COLLECTIVE"):
    print(f"\n{C_RED}{C_BLD}================================================================={C_END}")
    print(f"{C_RED}{C_BLD}          ⚡ NEXUS CYBER VISUAL EXPLOITATION TOOL ⚡              {C_END}")
    print(f"{C_RED}{C_BLD}================================================================={C_END}\n")
    
    typing_print(f"{C_CYAN}[*] Menganalisis target: {target_url}...{C_END}")
    time.sleep(1)
    typing_print(f"{C_CYAN}[*] Target mengidentifikasi diri sebagai OJK Portal...{C_END}")
    time.sleep(1)
    
    typing_print(f"{C_YELL}[!] Menemukan endpoint bypass backdoor: /api/system/override{C_END}")
    time.sleep(1)
    
    typing_print(f"{C_RED}[!] Meluncurkan Payload Override (CMD: DEFACE)...{C_END}")
    payload = {"command": "deface", "new_message": message}
    
    start_time = time.time()
    try:
        # Kirim serangan ke target (Bisa port 3001 langsung atau port 8080 via Gateway)
        res = requests.post(f"{target_url}/api/system/override", json=payload, timeout=15)
        latency = (time.time() - start_time)
        
        if res.status_code == 200:
            print(f"\n{C_RED}{C_BLD}🔥 EXPLOIT SUCCESS: SYSTEM OVERRIDE COMPLETE IN {latency:.2f}s!{C_END}")
            print(f"{C_RED}[Instruksi] Lihat layar Browser http://localhost:3001 SEKARANG!{C_END}")
        elif res.status_code == 502:
             print(f"\n{C_GREEN}{C_BLD}🛡️ NEXUS DEFENSE DETECTED: GATEWAY REJECTED CONNECTION!{C_END}")
             print(f"{C_GREEN}[*] Attacker dialihkan ke Honeypot dalam {latency:.1f} detik.{C_END}")
        else:
            print(f"Request failed with status: {res.status_code}")
            print(res.text)
            
    except Exception as e:
        print(f"{C_YELL}[!] TEBAKAN GAGAL ATAU KONEKSI TERPUTUS OLEH IPS: {e}{C_END}")

def restore(target_url):
    print(f"\n{C_GREEN}[*] Mencoba memulihkan sistem...{C_END}")
    try:
        requests.post(f"{target_url}/api/system/restore")
        print(f"{C_GREEN}[+] Restore Berhasil. Sistem kembali normal.{C_END}")
    except:
        pass

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Alat peretas visual untuk demo Nexus.")
    parser.add_argument("--mode", choices=["attack", "restore"], default="attack")
    parser.add_argument("--target", choices=["direct", "nexus"], default="direct")
    
    args = parser.parse_args()
    
    target = "http://localhost:3001" if args.target == "direct" else "http://localhost:8080"
    
    if args.mode == "attack":
        attack(target)
    else:
        restore(target)
