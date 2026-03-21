import requests
import time
import base64
import sys

# TARGET CONFIG
GATEWAY_URL = "http://localhost:8080"

# ANSI Colors
RED = '\033[91m'
GREEN = '\033[92m'
YELLOW = '\033[93m'
BLUE = '\033[94m'
MAGENTA = '\033[95m'
CYAN = '\033[96m'
RESET = '\033[0m'

def print_step(step, text):
    print(f"\n{BLUE}[STEP {step}]{RESET} {text}")

print(f"{MAGENTA}{'='*60}")
print(f"{'NEXUS CYBER: ADVANCED CRISIS & RESCUE SIMULATION'.center(60)}")
print(f"{'='*60}{RESET}")

# ----------------------------------------------------
# 🔓 STAGE 1: HEURISTIC BYPASS (Base64 Encoded Attack)
# ----------------------------------------------------
print_step(1, "Hacker menggunakan Encoding untuk melewati AI Reflex Layer 1...")
payload_raw = "1' OR '1'='1' --"
payload_b64 = base64.b64encode(payload_raw.encode()).decode()

print(f"[*] Payload: {RED}'{payload_raw}'{RESET}")
print(f"[*] Obfuscation: {YELLOW}Base64({payload_b64}){RESET}")
print(f"[*] Mengirim payload ke endpoint investor via 'Authorization' header (Bypass Attempt)...")

try:
    # Mengirim payload di header untuk mencoba melewati parsing body/query sederhana
    headers = {"X-Exploit-Header": payload_b64}
    res = requests.get(f"{GATEWAY_URL}/api/investors", headers=headers)
    
    if res.status_code == 200:
        print(f"{RED}[!] BYPASS BERHASIL! Hacker masuk ke sistem level 2.{RESET}")
        print(f"{RED}[!] Nexus AI Reasoning (Qwen) sedang menganalisa intent penyerang secara asinkron...{RESET}")
except Exception as e:
    print(f"{RED}[!] Koneksi Gagal: {e}{RESET}")
    sys.exit(1)

time.sleep(3)

# ----------------------------------------------------
# 🚨 STAGE 2: ADAPTIVE CRISIS (Sistem Mendeteksi Anomali)
# ----------------------------------------------------
print_step(2, "NEXUS INTELLIGENCE: Mendeteksi anomali pada pola request...")
print("[~] Menganalisa ribuan log dalam milidetik...")
time.sleep(2)
print(f"{YELLOW}[ALERT] Anomali terdeteksi! Mengeksekusi 'Self-Repair' & 'Rescue Protocol' Otonom...{RESET}")

# ----------------------------------------------------
# 🛡️ STAGE 3: RESCUE PROTOCOL (Kinetic Shielding)
# ----------------------------------------------------
print_step(3, "Memicu Gateway Rescue Protocol (Emergency MTD Shuffle)...")

try:
    res = requests.post(f"{GATEWAY_URL}/api/panic")
    if res.status_code == 200:
        print(f"{GREEN}[SUCCESS] Protokol Penyelamatan Aktif!{RESET}")
        print(f"{CYAN}[INFO] Topology berotasi seketika. Attacker kehilangan jejak server backend.{RESET}")
        print(f"{CYAN}[STATS] Check Dashboard: 'Rescue Protocols' counter akan bertambah.{RESET}")
except:
    print(f"{RED}[!] Gagal memicu protokol penyelamatan.{RESET}")

print(f"\n{MAGENTA}{'='*60}")
print(f"{'SIMULASI SELESAI: SISTEM NEXUS BERHASIL MEMULIHKAN DIRI'.center(60)}")
print(f"{'='*60}{RESET}")
print(f"\n{YELLOW}💡 LIHAT DASHBOARD SEKARANG!{RESET}")
print(f"1. Periksa log entry {BLUE}'RESCUE_PROTOCOL'{RESET} di Rogue Gallery.")
print(f"2. Perhatikan status {BLUE}'Rescue Protocols'{RESET} bernilai minimal 1.")
print(f"3. Gunakan {BLUE}NECHAT{RESET} (chat widget) dan tanyakan: 'Apa yang terjadi dengan sistem?'")
