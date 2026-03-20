# 🛡️ NEXUS CORE DIRECTIVES
**Project: Nexus Cyber - Autonomous Database Security Gateway**
**Status: Phase 1 - Initiation of Skills & Knowledge Base**

Dokumen ini merangkum aturan main utama dan arsitektur inti dari sistem Infrastructure/Database Security Gateway yang dibangun sebagai pertahanan vital nasional.

---

## 🧠 @skill-dual-brain: Ensemble AI Architecture (Revision 2 — Cloud Inference)
Setiap perancangan logika analitik trafik WAJIB menggunakan arsitektur **Ensemble AI tertutup**.
1. **Reflex Layer — Qwen3 32B (Groq API)**: Filtrasi real-time trafik massal dengan latensi < 50ms menggunakan Groq LPU. API Key: `GROQ_API_KEY` (env var, JANGAN hardcoded). Model slug: `qwen3-32b`.
2. **Reasoning Layer — Qwen3 235B-A22B (OpenRouter API)**: Analisis forensik kontekstual mendalam (*Contextual Intent Analysis*) untuk trafik yang lolos dari Layer 1. Dijalankan secara **asinkron** (goroutine) dengan timeout 30 detik. API Key: `OPENROUTER_API_KEY` (env var). Model slug: `qwen/qwen3-235b-a22b`.

**Migrasi**: Sistem resmi meninggalkan Ollama (model lokal). Cloud inference dipilih untuk mengejar latensi ultra-rendah dan kapasitas penalaran tingkat tinggi.
*Peringatan: Dilarang menggunakan model AI tunggal. Dilarang meng-hardcode API Key.*

## 🛠️ @skill-self-repair: Autonomous Recursive Repair
Mitigasi dan perbaikan bug/celah keamanan harus bersifat otonom tanpa intervensi manual admin.
1. **Recursive Self-Repair**: Menulis skrip perbaikan mandiri.
2. **Virtual Patching**: Patching otonom terhadap ancaman baru.
3. **Instant Rollback**: Jika baseline sistem berubah, rollback dilakukan dalam hitungan milidetik (*Zero-Downtime*).

## 🌐 @skill-mtd: Moving Target Defense
Implementasi pada jaringan, port, dan IP internal untuk mengecoh penyerang.
1. **Dynamic Configuration Randomization**: Mengacak target secara stokastik.
2. **Digital Hallucination**: Menciptakan Sandbox palsu (Deception Technology) untuk menjebak peretas.

## 🔒 @skill-pqc: Post-Quantum Cryptography
DILARANG menggunakan enkripsi legasi (RSA, AES standar, ECC murni) untuk data sensitif.
1. **Lattice-based Cryptography**: Menggunakan standar NIST (ML-KEM/Kyber).
2. **Harvest Now, Decrypt Later Mitigation**: Fokus pada keamanan jangka panjang terhadap ancaman komputer kuantum.

## 🎨 @skill-ui-ux-design: Command Center UI/UX
Desain antarmuka admin yang elegan, berwibawa, dan efisien untuk pusat komando nasional.
1. **Next.js + Tailwind CSS**: Stack utama untuk performa dan desain responsif.
2. **Cyber Aesthetic / Dark Mode**: Tema profesional yang meminimalkan distraksi visual.
3. **Real-time Data Visualization**: Fokus pada grafik anomali trafik, status MTD, Dan log enkripsi PQC.
4. **Mission Critical Clarity**: Admin harus memahami status keamanan dalam "satu lirikan mata".

## 🕵️ @skill-qa-iso-auditor: Zero-Tolerance Quality Audit
Gatekeeper kualitas kode dan arsitektur berdasarkan standar ISO/IEC 25010 & 27001.
1. **Security Audit**: Audit kebocoran data, kredensial, dan celah injeksi di setiap modul.
2. **Performance KPI**: Mengawasi latensi tinggi terutama pada layer AI dan PQC.
3. **Reliability & Failover**: Memastikan sistem tetap berjalan meskipun Layer AI atau jaringan gagal/terganggu.
4. **UU PDP Compliance**: Memastikan setiap data pribadi warga negara yang diproses mematuhi regulasi Pelindungan Data Pribadi Indonesia.

---
*"Zero-Tolerance for Vulnerabilities. Maximum Performance for Sovereignty."*

---
*"Demi kedaulatan data nasional, Nexus Cyber beroperasi dengan presisi dan otonomi penuh."*
