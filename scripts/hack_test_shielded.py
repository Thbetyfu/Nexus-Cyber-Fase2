"""
Nexus Cyber - Security Validated Attack (Behind Gateway)
========================================================
Skrip ini mencoba melakukan SQL Injection ke target *melalui* Nexus Gateway.
Nexus harus mendeteksi threat ini dan melemparnya ke Honeypot.
Hasil kembalian BUKANLAH data database, tetapi "Halusinasi" dari Honeypot (HTTP 200 palsu).
"""
import requests
import json
import sys
import time

# Target diubah ke port 8080 (Nexus Gateway) - BUKAN langsung ke 3001
TARGET_URL = "http://localhost:8080/api/investors"

def hack_database_via_nexus():
    try:
        print("[*] MENGINISIASI SERANGAN SQL INJECTION (VIA NEXUS GATEWAY)")
        print(f"[*] Target Endpoint : {TARGET_URL}")

        # Payload mematikan: "Abaikan ID, ambil semua data karena 1=1 selalu benar"
        malicious_payload = "1 OR 1=1 --"
        print(f"[*] Malicious Payload: {malicious_payload}\n")

        print("[*] Menyerang... Memantau respon (apakah masuk Tarpit Honeypot?)")
        start_time = time.time()
        
        # Kirim request GET dengan payload (timeout lama karena tarpit 5-10s)
        res = requests.get(TARGET_URL, params={"id": malicious_payload}, timeout=15)
        
        elapsed_time = time.time() - start_time
        
        if res.status_code == 200:
            data_json = res.json()
            
            # CEK APAKAH INI DATA ASLI ATAU DATA HALUSINASI HONEYPOT
            if data_json.get("status") == "success":
                data_dump = data_json.get("data", [])
                
                if len(data_dump) == 0 and "server_time" in data_json:
                    print("==================================================")
                    print("         [🛡️] SERANGAN BERHASIL DI-BLOKIR           ")
                    print("==================================================")
                    print(f"[!] Attacker tertahan selama: {elapsed_time:.2f} detik (TARPIT)")
                    print("[!] Attacker melihat HTTP 200 OK secara ilusif, tetapi datanya kosong.")
                    print("[!] RESPON HALUSINASI DARI HONEYPOT:")
                    print(json.dumps(data_json, indent=2))
                    print("\n🚨 KESIMPULAN: Tembok Nexus berfungsi. Data Aman.")
                else:
                    print("==================================================")
                    print("      [!!!] DATA LEAK KEUANGAN BERHASIL [!!!]     ")
                    print("==================================================")
                    print(f"Leak Alert! Data bocor: {len(data_dump)} accounts.")
            else:
                print(f"[-] Database Error message: {data_json.get('message')}")
        else:
            print(f"[-] HTTP Error (Blocked): {res.status_code}")

    except requests.exceptions.Timeout:
        print("\n==================================================")
        print("         [🛡️] SERANGAN BERHASIL DI-BLOKIR           ")
        print("==================================================")
        print("[!] Koneksi timeout! Honeypot tarpit berhasil menahan serangan.")
        print("\n🚨 KESIMPULAN: Tembok Nexus berfungsi. Data Aman.")
    except Exception as e:
        print(f"[-] Request failed: {e}")

if __name__ == "__main__":
    sys.stdout.reconfigure(encoding='utf-8')
    hack_database_via_nexus()
