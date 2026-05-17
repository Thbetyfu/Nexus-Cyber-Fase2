// Package mtd (Moving Target Defense) menyediakan modul kontrol pertahanan aktif dan manipulasi firewall OS.
package mtd

import (
	"log"
	"os/exec"
	"runtime"
	"strings"
)

// BlockIPAtOSLevel melakukan pemblokiran IP penyerang secara permanen di tingkat sistem operasi (Kernel Space).
//
// Alasan Arsitektural (Why):
// Menggunakan pemblokiran Netfilter (iptables) untuk menjamin paket dibuang seketika oleh kernel Linux.
// Ini adalah pertahanan tingkat lanjut yang meminimalisir overhead pemrosesan data (CPU) di User Space.
func BlockIPAtOSLevel(ip string) {
	/*
	   NEXUS_FIX_LOG: [KERNEL_LOCKOUT_GUARD]
	   - Kenapa ini ada? Dashboard administrasi dan Gateway berjalan di atas jaringan bridge Docker (172.x.x.x).
	   - Jika kita tidak mengecualikan subnet ini, kesalahan deteksi (false alarm) dapat memblokir jaringan internal,
	     menyebabkan Dashboard terputus dari Gateway secara permanen (Self-Inflicted Denial of Service).
	   - Dilarang memblokir loopback (127.0.0.1, ::1) untuk menjamin admin tetap memiliki akses debugging lokal.
	*/
	if strings.HasPrefix(ip, "172.") || strings.Contains(ip, "127.0.0.1") ||
		strings.Contains(ip, "::1") || strings.Contains(ip, "localhost") {
		log.Printf("[OS-FIREWALL] Authorized Node Bypass: %s (Protected from self-lockout)", ip)
		return
	}

	// Normalisasi IP: Bersihkan sisa port yang melekat pada remote address (misal "8.8.8.8:4928" -> "8.8.8.8").
	// Alasan Teknis: Perintah iptables akan memicu error sintaks jika mendeteksi simbol titik dua ":" dan nomor port.
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	// Penanganan kompatibilitas lintas platform (Windows/macOS) untuk keperluan testing lokal tanpa crash.
	if runtime.GOOS == "windows" {
		log.Printf("[OS-FIREWALL] skipping iptables on Windows for IP: %s (Simulation Only)", ip)
		return
	}

	log.Printf("[OS-FIREWALL] KERNEL_DROP_INIT: Attempting to block %s via Netfilter...", ip)

	// Memeriksa apakah aturan pemblokiran IP ini sudah eksis di tabel iptables INPUT (Check command "-C").
	// Alasan Teknis (Why): Mencegah duplikasi entri di tabel kernel yang dapat memperlambat proses pencarian rute paket.
	checkCmd := exec.Command("iptables", "-C", "INPUT", "-s", ip, "-j", "DROP")
	if err := checkCmd.Run(); err == nil {
		log.Printf("[OS-FIREWALL] %s is already dropped by kernel. Skipping duplicate rules.", ip)
		return
	}

	// Eksekusi penambahan aturan pemblokiran mutlak di rantai INPUT (Append command "-A").
	cmd := exec.Command("iptables", "-A", "INPUT", "-s", ip, "-j", "DROP")
	err := cmd.Run()
	if err != nil {
		// Log kegagalan (biasanya karena kontainer Docker tidak dijalankan dengan kapabilitas --privileged atau --cap-add=NET_ADMIN).
		log.Printf("[OS-FIREWALL] KERNEL_DROP_ERROR: Failed to execute iptables for %s: %v. Ensure NET_ADMIN privilege is granted.", ip, err)
		return
	}

	log.Printf("[OS-FIREWALL] CRITICAL_DEFENSE: IP '%s' has been officially DROP-ed by Netfilter (Kernel Level).", ip)
}

// UnblockIPAtOSLevel menghapus IP penyerang dari tabel pemblokiran iptables setelah durasi hukuman berakhir (Remediasi).
func UnblockIPAtOSLevel(ip string) {
	if runtime.GOOS == "windows" {
		return
	}

	cmd := exec.Command("iptables", "-D", "INPUT", "-s", ip, "-j", "DROP")
	if err := cmd.Run(); err != nil {
		log.Printf("[OS-FIREWALL] Failed to release IP %s: %v", ip, err)
	}
}
