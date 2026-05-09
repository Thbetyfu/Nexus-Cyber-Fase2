# 🛠️ Nexus Cyber Git Workflow Guide

Dokumen ini berisi panduan cara mengelola kode (Push & Pull) pada proyek Nexus Cyber yang menggunakan sistem **Git Submodule** untuk folder `Portfolio-website-main`.

---

## 1. Persiapan Awal (First Time Setup)
Jika Anda baru pertama kali meng-clone repositori ini, pastikan Anda juga menarik data submodule-nya:

```bash
# Clone dengan submodule sekaligus
git clone --recursive https://github.com/YourUsername/Nexus-Cyber-Otonous.git

# ATAU jika sudah terlanjur clone biasa:
git submodule update --init --recursive
```

---

## 2. Cara Menarik Update (Pulling)
Karena ada repositori di dalam repositori, perintah `git pull` biasa tidak akan meng-update isi folder portofolio. Gunakan perintah ini:

```bash
# Menarik update di Main Repo dan Submodule sekaligus
git pull origin main --recurse-submodules
```

---

## 3. Cara Mengubah Portofolio (Pushing in Submodule)
Jika Anda melakukan perubahan di dalam folder `Portfolio-website-main`, Anda harus melakukan push di **dua tempat**:

### Langkah A: Di dalam folder Portofolio
```bash
cd Portfolio-website-main
git add .
git commit -m "Update tampilan portofolio"
git push origin main
```

### Langkah B: Di folder Utama (Nexus Root)
Setelah portofolio di-push, folder utama Nexus akan mencatat bahwa ada "versi baru" dari portofolio. Anda harus meng-update referensinya:
```bash
cd ..
git add Portfolio-website-main
git commit -m "chore: update portfolio reference to latest commit"
git push origin main
```

---

## 4. Ringkasan Perintah Penting

| Kebutuhan | Perintah |
| :--- | :--- |
| **Update Semua** | `git submodule update --remote --merge` |
| **Cek Status** | `git submodule status` |
| **Clone Baru** | `git clone --recursive [URL]` |

---
> **PENTING**: Jangan pernah menghapus folder `.git` di dalam `Portfolio-website-main` karena itu akan memutuskan koneksi submodule.
