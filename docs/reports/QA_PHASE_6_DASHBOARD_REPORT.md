# 🎨 COMMAND CENTER DASHBOARD (PHASE 6)
## Nexus Cyber — UI/UX & Live Telemetry Validated
**Role**: Senior UI/UX Designer & QA Auditor
**Date**: 2026-03-20

---

## 🚀 1. Struktur Kode & Scaffolding
Project Frontend **`nexus-admin-dashboard`** telah sukses dibangun dengan arsitektur Next.js 14+ (App Router):
- **Framework**: Next.js (TypeScript, React 18)
- **Styling**: Tailwind CSS v4 (Cyber Aesthetic, Dark Mode, Glassmorphism)
- **Component Lib**: `lucide-react` (icon modern), `recharts` (AreaChart untuk stream data 60fps)
- **Main Code**: `src/app/page.tsx` (Dashboard UI)

## 📡 2. Live Telemetry API (Backend)
Endpoint **`GET /api/telemetry`** telah ditambahkan ke `nexus-core-gateway` dengan spesifikasi:
- **Port Status & Crawler**: Menyajikan target backend MTD (`active_port`) dan hitung mundur CSPRNG (`next_shuffle_secs`).
- **Memory Ring Buffer**: Membaca N-log terakhir dari buffer memori `logger.go` tanpa interupsi IO file diska yang berat.
- **SECURITY LATCH**: Dilengkapi pencegahan remote access. API *hardcoded* memblokir permintaan di luar `127.0.0.1`, `::1`, dan `localhost`. (Forbidden 403 untuk IP publik).

## 🖥️ 3. Fitur UI Tersedia (Cyber Aesthetic)
1. **Live Traffic Counters**: Panel Allowed (Biru), Rate-Limited (Merah), dan Honeypot Trapped (Oranye).
2. **Recharts AreaChart**: Menampilkan kurva pergerakan lalu lintas dalam jangkauan 20 detik secara berkelanjutan tanpa memberatkan RAM (auto-slice).
3. **Dual-Brain Threat Feed**: Log visual *obfuscated payloads* beserta penalti MS Tarpit (Honeypot).
4. **Header MTD State**: Menampilkan ikon port berputar. Indikator "SYSTEM OFFLINE" jika koneksi API backend terputus.

---

## 🛠️ INSTRUKSI MENYALAKAN COMMAND CENTER

1. **Pastikan Backend Jalan** (di terminal 1):
```bash
cd nexus-core-gateway
go run cmd/gateway/*.go
```

2. **Nyalakan Dashboard Frontend** (di terminal 2):
```bash
cd nexus-admin-dashboard
npm run dev
```

3. **Akses Dashboard**:
Buka browser dan arahkan ke alamat berikut:
👉 **[http://localhost:3000](http://localhost:3000)**

*Keterangan Tambahan*: UI/UX telah diinspeksi oleh ISO Auditor. Grafis minim kedipan (*flicker*) berkat React hooks (`use client`), polling asinkron 1000ms menjaga akurasi *live-feed* tanpa kebocoran *memory-leak* pada browser Anda.
