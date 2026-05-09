# ⚠️ Nexus Cyber Limitations

Meskipun Nexus Cyber adalah sistem pertahanan otonom yang sangat tangguh, ada beberapa area yang secara teknis berada di luar cakupan perlindungan Gateway:

## 1. Social Engineering (Faktor Manusia)
Nexus bisa memblokir skrip jahat, tapi ia tidak bisa mencegah jika seorang karyawan memberikan password-nya secara sukarela karena tertipu telepon phishing atau pesan WhatsApp palsu.

## 2. Physical Security (Keamanan Fisik)
Nexus tidak bisa melindungi jika seseorang datang langsung ke ruang server Anda dan mencabut kabel listrik atau mencuri harddisk secara fisik.

## 3. Insider Threat (Ancaman Orang Dalam)
Jika seorang Admin yang sudah memiliki akses penuh memutuskan untuk menghapus database secara sengaja dari dalam terminal server (bukan lewat gateway), Nexus tidak akan mendeteksinya.

## 4. Hardware-Level Vulnerabilities
Serangan seperti Spectre atau Meltdown yang menyerang kelemahan pada chipset CPU adalah ranah firmware dan hardware vendor.

## 5. Email-Specific Attacks
Nexus menjaga trafik web (HTTP/TCP), tapi ia bukan sistem Email Security. Ia tidak memindai lampiran email yang berisi virus (malware) yang dikirim langsung ke kotak masuk pengguna.
