# 🛡️ Nexus-Cyber tahap 2

**Autonomous Tactical Defense Grid & Geospatial Threat Intelligence Command Center**

### 1. Pemuatan Modul Keamanan (Boot Sequence)
![System Boot Sequence](./docs/img/Opening-Nexus-Cyber.jpeg)

### 2. Panel Kendali Utama (SOC Command Center Dashboard)
![SOC Command Center Dashboard](./docs/img/Dashboard-Nexus-Cyber.jpeg)

### 3. Layar Pengunci Lisensi (Subscription Lockout Overlay)
![System License Lockout](./docs/img/System-Lock-Nexus-Cyber.jpeg)

Nexus Cyber adalah sistem pertahanan siber otonom (SOC) yang menggabungkan AI Lokal (**Nexus-Brain**) dengan teknologi **Moving Target Defense (MTD)** untuk melindungi infrastruktur kritis dari serangan modern.

---

## 🛠️ Fitur Utama (Features)

Nexus-Cyber tahap 2 dilengkapi dengan berbagai teknologi keamanan mutakhir yang terbagi dalam beberapa lapisan pertahanan:

### 1. Pertahanan AI Berbasis Dua Lapis (Dual-Brain AI Shield)

* **Reflex Layer**: Deteksi cepat menggunakan model Qwen3 32B via cloud API (Groq) dengan latensi ultra-rendah (<50ms) untuk pemblokiran serangan secara instan (SQL Injection, XSS, SSRF).
* **Reasoning Layer**: Analisis forensik mendalam untuk mengidentifikasi intensi penyerang secara asinkron menggunakan model Qwen3 235B-A22B via OpenRouter API.

### 2. Pertahanan Dinamis (Moving Target Defense - MTD)

* **Topology Port Shuffling**: Rotasi port komunikasi internal secara berkala berbasis CSPRNG untuk mengecoh pemetaan jaringan (*network scanning*) oleh peretas.
* **Emergency Manual Shuffle**: Fitur rotasi port instan secara manual dari terminal jika terdeteksi kondisi darurat.

### 3. Teknologi Deception & Stalling (Honeypot Sandbox)

* **Isolated Honeypot**: Server umpan terisolasi pada port `:9090` untuk menjebak pemindai otomatis hacker.
* **Tarpit Delay**: Menahan koneksi penyerang selama 8 detik secara sengaja untuk menguras *resource* penyerang, yang kemudian disiarkan langsung ke dasbor telemetri.

### 4. Sanitasi Berkas Visual (AVSE - Anti-Vulnerability SQL/XSS Engine)

* **Magic Byte Verification**: Verifikasi signature biner asli untuk mencegah bypass ekstensi ganda (seperti berkas `shell.php.png`).
* **Visual Steganography Stripping**: Dekode dan re-encode biner piksel untuk melumpuhkan kode exploit yang sengaja disisipkan di ekor berkas gambar.
* **EXIF/GPS Purging**: Pembersihan otomatis seluruh metadata lokasi kamera demi privasi pengunggah berkas.

### 5. Proteksi Client-Side Anti-Inspect Hardening

* **Context Menu Blocking**: Mencegah klik kanan untuk membatalkan akses menu "Inspect Element".
* **Keyboard Shortcut Hooks**: Memblokir pintasan devtools (`F12`, `Ctrl+Shift+I/J/C`, `Ctrl+U`).
* **Debugger Infinite Loop Tarpit**: Membekukan peramban hacker menggunakan loop debugger terus-menerus jika dipaksa masuk dari menu browser.
* **Continuous Console Purging**: Pembersihan logs konsol per milidetik untuk mencegah pemetaan API.

### 6. Multi-Tenant SaaS Licensing & Lockout

* **Remote License Verification**: Validasi status lisensi client secara berkala menggunakan kunci `NEXUS_LICENSE_KEY`.
* **Global Lockout Overlay**: Layar pengunci gelap premium berukuran penuh yang tidak dapat dilewati secara DOM jika lisensi kedaluwarsa atau dicabut.

### 7. Terminal Komando Interaktif (SOC Command CLI)

* Mendukung eksekusi perintah administrator seperti: `/help`, `/status`, `/stats`, `/shuffle`, `/ban [IP]`, `/unban [IP]`, `/sub [domain]`, `/unsub [domain]`, `/honeystats`, `/patches`, dan `@nexus [query]` untuk konsultasi AI.

### 8. Database Forensik & Kepatuhan ISO 27001

* Penyimpanan log anomali dan jejak audit secara terstruktur dalam database **PostgreSQL** (`threat_logs`, `mtd_audit_trail`, `intel_blacklist`, `ai_insights`) serta **Redis** untuk *in-memory caching* dan *rate limiting*.

### 9. Fitur Pengujian Simulasi & Ketahanan Riil (Testing & Simulation Mode)

* **Autoban IP & Persistent Blacklist**: Mekanisme pemblokiran IP penyerang secara otomatis setelah 5 kali gagal menebak password vault hadiah. Fitur ini dapat diaktifkan kembali secara fungsional dalam kode untuk menguji skenario pemblokiran riil.
* **Geospatial Tracking & GeoIP Integration**: Melacak asal negara penyerang, nama ISP, koordinat geografis, serta sidik jari perangkat penyerang (*device fingerprinting*) untuk dipetakan secara real-time pada Defense Matrix Dashboard.

---

## 📂 Dokumentasi Proyek

Silakan baca dokumen di bawah ini untuk memahami sistem secara mendalam:

* [🏗️ **Architecture & Flow**](./docs/ARCHITECTURE.md) - Detail teknis MTD & AI Layers.
* [🛡️ **Capabilities**](./docs/CAPABILITIES.md) - Daftar serangan yang bisa dicegah.
* [⚠️ **Limitations**](./docs/LIMITATIONS.md) - Batasan perlindungan sistem.
* [🕹️ **CLI Guide**](./docs/CLI_GUIDE.md) - Panduan perintah Command Center.
* [🛠️ **Git Workflow**](./docs/GIT_WORKFLOW.md) - Panduan Push & Pull (Submodule).

---

## 🚀 Cara Menjalankan (Quick Start)

### 1. Menyalakan Sistem

Gunakan script kendali terpadu di folder `scripts/`:

```bash
./scripts/nexus-ignite.sh
```

### 2. Mematikan Sistem

```bash
./scripts/nexus-kill.sh
```

### 3. Melakukan Audit Keamanan

Jalankan alat uji terpadu untuk memverifikasi seluruh komponen:

```bash
cd nexus-brain-lab
python3 nexus_tester.py
```

---
*Nexus-Cyber tahap 2: Menjaga Kedaulatan Digital Indonesia dengan Imunitas Otonom & Intelijen Taktis.*
