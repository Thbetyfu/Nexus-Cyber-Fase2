# ⚔️ NEXUS CYBER: WAR GAME MANUAL
File ini berisi skenario serangan untuk menguji ketangguhan Nexus AI Gateway.

## 🚀 Cara Penggunaan
1. Pastikan Nexus sudah menyala (`./nexus-ignite.sh`).
2. Buka Dashboard SOC di [http://localhost:3000](http://localhost:3000).
3. Copy-paste URL serangan di bawah ini ke browser Anda.
4. Perhatikan Alarm dan Watermark di Dashboard!

---

## 1. 💉 SQL Injection (Pencurian Data)
Skenario: Hacker mencoba membocorkan seluruh database investor OJK.

**A. Dasar (OR 1=1):**
```text
http://localhost:8080/?search=' OR '1'='1
```

**B. Berbahaya (UNION SELECT):**
```text
http://localhost:8080/?search=UNION SELECT * FROM users--
```

**C. Merusak (DROP TABLE):**
```text
http://localhost:8080/?search='; DROP TABLE investors;--
```

---

## 2. 🛡️ Path Traversal (Akses File Rahasia)
Skenario: Hacker mencoba mengakses file sistem sensitif di server.

**A. Akses Password Linux:**
```text
http://localhost:8080/?file=../../../../etc/passwd
```

**B. Akses Config Windows:**
```text
http://localhost:8080/?file=C:\Windows\System32\drivers\etc\hosts
```

---

## 3. 🌀 XSS - Cross Site Scripting (Injeksi Script)
Skenario: Hacker mencoba mencuri cookie admin dengan menyuntikkan Javascript.

**A. Alert Popup:**
```text
http://localhost:8080/?user=<script>alert('HackedByNexus')</script>
```

**B. Onclick Event:**
```text
http://localhost:8080/?name=<b onmouseover="alert('XSS')">Hover Me</b>
```

---

## 4. ⚡ Brute Force & Rate Limiting
Skenario: Hacker mencoba membombardir server dengan ribuan request per detik.

**Cara Tes:**
Tekan tombol **Refresh (F5)** di browser Anda berkali-kali secepat mungkin (minimal 5-10 kali per detik).
*   **Hasil**: Nexus akan mendeteksi lonjakan trafik dan memblokir sementara IP Anda.
*   **Lihat Dashboard**: Status akan berubah menjadi `RATE_LIMITED`.

---

## 5. 🤖 Virtual Patching (Imunitas Otomatis)
Skenario: Setelah satu serangan terdeteksi, Nexus akan menciptakan "Antibodi".

**Cara Tes:**
1. Lakukan serangan `UNION SELECT` sekali.
2. Tunggu 2 detik.
3. Lakukan serangan yang **SAMA PERSIS** lagi.
4. **Hasil**: Lihat Dashboard, statusnya bukan lagi `DIVERTED` tapi **`IMMUNE`**. Artinya Nexus sudah tidak perlu menganalisis lagi karena sudah punya antibodinya.

---

> **CATATAN**: Seluruh serangan di atas akan dialihkan ke **Honeypot**. Anda akan melihat respons JSON "Success" dengan data kosong. Cek Watermark di Dashboard untuk konfirmasi!
