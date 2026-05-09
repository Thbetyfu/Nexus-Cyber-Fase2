# 🕹️ Command Center CLI Guide

Berikut adalah daftar perintah yang bisa digunakan oleh Admin SOC melalui antarmuka Command Center:

| Perintah | Deskripsi Teknis |
| :--- | :--- |
| `/help` | Menampilkan manifest perintah bantuan. |
| `/status` | Melakukan audit kesehatan telemetri & Redis Probe. |
| `/ban [IP]` | Injeksi antibody manual ke dalam Distributed Set. |
| `@nexus [MSG]` | Melakukan query kognitif ke AI Reasoning Engine. |

## Audit & Pengujian
Untuk melakukan pengujian sistem secara menyeluruh, gunakan aplikasi auditor terpadu kami:
```bash
cd nexus-brain-lab
python3 nexus_tester.py
```
