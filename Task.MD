1. dibuat menjadi sesuai uji coba dimana  saya ingin di buat menjadi sesutu website uji ketahanan, jadi dimana website ini akan di deploy lalu Nexus cyber harus bisa mempertahankan website portofolio ini dari serangan cyber, sehingga kekuatan dari nexus cyber ini benar terbukti tetapi belum tau apa saja rencana yang harus di lakukan, dan saya ingin di website portofolionya di tambahkan untuk input gambar, di bagian paling bawah, "tambahkan foto foto kalian" coba pikirkan rencananya, bukti sudah berhasilnya adalah nanti akanada pasword disitu, lalu ketika masuk menggunakan pasword itu maka bisa mendapatkan link dimana link beruba hadiah beruba uang , linknya akan berupa shope kaget (tapi itu nanti akan di pikirkan), jadi yang butuh dikejrakna adlaha:
   1. set pasword
   2. set arsitektur dan user flownya
   3. tambahkan fitur tambahkan foto di bagian paling bawah dan input pasword
   4. 
2. apakah nexus cyber ini sudah benar benar bisa di deploy?
3. tambahkan lebih banyak fitur pada bagain terminal agar tidak menjadi terminal yang kakku
4. uji tetsing (SELESAI - 17/17 MTD Checks Passed, [Laporan Audit](file:///home/taqy/.gemini/antigravity/brain/ca754110-72d4-48a0-b93a-4b25725cef2c/qa_phase_10_stress_test_report.md) Terbit)
5. performa CPU dan RAM butuh berapa?
6. apakah bisa di tambahkan fitur pelacakan bagi orang yang hack ??
7.

---

## 🚀 SPRINT ROADMAP: MULTI-TENANT DEDICATED SAAS & GLOBAL LICENSE LOCKOUT (NEXUS CYBER SAAS)

Rencana kerja strategis untuk mengubah seluruh infrastruktur Nexus Cyber menjadi sistem SOC (Security Operations Center) SaaS berbasis langganan yang dapat disewa secara mandiri oleh klien:

### 🏢 1. Modul Verifikasi Lisensi Global (Backend Layer)
* [ ] **Task 1: Penyiapan Modul `internal/licensing`**
  * Buat pustaka otentikasi lisensi terpusat di sisi Go Gateway.
  * Tambahkan penanganan parameter variabel lingkungan `NEXUS_LICENSE_KEY`.
* [ ] **Task 2: Penyiapan Handshake Rutin**
  * Jalankan goroutine asinkron yang melakukan verifikasi masa sewa secara berkala (misal: setiap 1 jam sekali) ke server lisensi pusat (`https://license.nexus-cyber.com/verify`).
* [ ] **Task 3: Implementasi Fail-Secure Lockout**
  * Jika status lisensi `EXPIRED` atau `REVOKED`, gateway akan menangguhkan seluruh rute proxying atau mengembalikan kode pemblokiran total ke browser pengunjung.

### 🎭 2. Layar Pengunci Dasbor (Frontend Layer)
* [ ] **Task 4: Dynamic License Guard (React Router Context)**
  * Deteksi status validitas lisensi dari respon backend REST/SSE API di dasbor Command Center.
* [ ] **Task 5: Premium Dark Lockout Overlay**
  * Buat komponen modal overlay berukuran penuh (*full-screen unbypassable overlay*) berwarna gelap premium dengan teks peringatan: `"SISTEM DITANGGUHKAN: Masa sewa langganan Nexus Cyber Anda telah berakhir."`
  * Lindungi layar pengunci agar tidak dapat dilewati menggunakan manipulasi DOM atau inspeksi elemen.

### 📦 3. Orkestrasi Pendaftaran Otomatis (Infrastructure Layer)
* [ ] **Task 6: Dynamic Provisioner Script (`scripts/provisioner.sh`)**
  * Buat skrip asinkron yang meluncurkan kontainer Docker-Compose baru secara terisolasi (Gateway + Dashboard + PostgreSQL) untuk setiap penyewa (*tenant*) baru.
* [ ] **Task 7: Integrasi Midtrans/Stripe Webhook**
  * Sambungkan gerbang pembayaran otomatis ke skrip penyediaan kontainer untuk aktivasi instan setelah klien berhasil melakukan pembayaran bulanan.
