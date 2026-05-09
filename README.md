# 🛡️ Nexus Cyber SOC v13.2
**Autonomous Tactical Defense Grid & Geospatial Threat Intelligence Command Center**

![Nexus Cyber Dashboard V13](./docs/img/dashboard_v13.png)

Nexus Cyber adalah sistem pertahanan siber otonom (SOC) yang menggabungkan AI Lokal (**Nexus-Brain**) dengan teknologi **Moving Target Defense (MTD)** untuk melindungi infrastruktur kritis dari serangan modern.

---

## 📂 Dokumentasi Proyek
Silakan baca dokumen di bawah ini untuk memahami sistem secara mendalam:

*   [🏗️ **Architecture & Flow**](./docs/ARCHITECTURE.md) - Detail teknis MTD & AI Layers.
*   [🛡️ **Capabilities**](./docs/CAPABILITIES.md) - Daftar serangan yang bisa dicegah.
*   [⚠️ **Limitations**](./docs/LIMITATIONS.md) - Batasan perlindungan sistem.
*   [🕹️ **CLI Guide**](./docs/CLI_GUIDE.md) - Panduan perintah Command Center.
*   [🛠️ **Git Workflow**](./docs/GIT_WORKFLOW.md) - Panduan Push & Pull (Submodule).

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
*Nexus Cyber SOC v13.2: Menjaga Kedaulatan Digital Indonesia dengan Imunitas Otonom & Intelijen Taktis.*
