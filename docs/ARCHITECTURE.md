# 🏗️ NEXUS CYBER ARCHITECTURE

Dokumen ini memvisualisasikan struktur folder dan spesifikasi teknologi yang digunakan untuk membangun infrastruktur pertahanan data nasional.

## 💻 Tech Stack Proposal

### 1. Backend Defense (nexus-core-gateway)
- **Language**: **Go (Golang)**. Dipilih karena performa tinggi, konkurensi aman, dan ekosistem library keamanan yang matang.
- **Framework**: Standard library `net/http` & `httputil` untuk proxy, Gin/Fiber untuk API internal.
- **AI Integration**: gRPC/REST untuk Qwen (Reflex) dan Llama 3 (Reasoning).
- **PQC Library**: `circl` (Cloudflare) atau bindings ML-KEM NIST.
- **Database**: **Redis** (Real-time MTD Tracking) & **PostgreSQL** (Metadata/Audit Log).

### 2. Frontend Command Center (nexus-admin-dashboard)
- **Framework**: **Next.js 14+ (App Router)**.
- **Styling**: **Tailwind CSS**.
- **Visualisasi**: **Tremor** / **Recharts** untuk monitor grafik real-time.
- **Icons**: **Lucide-React**.
- **State Management**: **Zustand** atau React Context.

### 3. AI Layers (Dual-Brain)
- **Reflex Layer**: Qwen (Optimization for speed).
- **Reasoning Layer**: Llama 3 (Optimization for context/intent).

---

## 📁 Directory Structure

```text
nexus-cyber/
├── .agents/                    # AI Agent Configs & Skills
├── nexus-core-gateway/          # BACKEND (GO)
│   ├── cmd/
│   │   └── gateway/             # Entry point aplikasi
│   ├── internal/                # Privat logic
│   │   ├── ai/                  # Dual-Brain Logic
│   │   ├── mtd/                 # Moving Target Defense
│   │   ├── crypto/              # PQC (NIST ML-KEM)
│   │   ├── proxy/               # Reverse Proxy Core
│   │   └── repair/              # Self-Repair Scripts
│   ├── pkg/                     # Public shared packages
│   ├── configs/                 # YAML/Env Configurations
│   ├── go.mod
│   └── README.md
├── nexus-admin-dashboard/       # FRONTEND (NEXT.JS)
│   ├── app/                     # Next.js App Router
│   ├── components/              # UI Components (Cyber Aesthetic)
│   ├── lib/                     # Utils & Hooks
│   ├── public/                  # Assets
│   ├── tailwind.config.ts
│   └── package.json
├── scripts/                     # Deployment & Maintenance Scripts
│   └── setup.sh                 # Scaffolding Automation
├── docs/                        # Technical Documentation
└── NEXUS_CORE_DIRECTIVES.md     # Governance Rules
```

---
*Antara `nexus-core-gateway` dan `nexus-admin-dashboard` terhubung via API internal yang dilindungi oleh PQC.*
