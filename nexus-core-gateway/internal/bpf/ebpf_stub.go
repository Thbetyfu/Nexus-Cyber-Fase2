// Package bpf mengintegrasikan manajer tingkat rendah (kernel-level) menggunakan eBPF untuk mitigasi serangan siber.
package bpf

import (
	"log"
)

// BpfManager mengimplementasikan jembatan kendali eBPF (Extended Berkeley Packet Filter) di tingkat kernel.
//
// Alasan Arsitektural (Why):
// Untuk menahan serangan Denial of Service (DoS) volumetrik berskala besar, pemblokiran di tingkat aplikasi
// (seperti HTTP 403 Forbidden) sangat tidak efisien karena tetap mengonsumsi CPU socket dan RAM server.
// Dengan eBPF, kita memprogram kartu jaringan (NIC) secara langsung menggunakan XDP (eXpress Data Path).
// Paket dari IP penyerang akan dibuang seketika di tingkat driver jaringan (XDP_DROP) sebelum kernel Linux
// sempat membuat alokasi buffer soket (sk_buff), meminimalkan penggunaan CPU gateway hingga mendekati 0% (Zero CPU Overhead).
type BpfManager struct {
	mapName string
}

// NewBpfManager mengonstruksi manajer eBPF stub.
//
// Alasan Teknis (Why):
// Menggunakan desain Stub Pattern agar gateway tetap dapat dikompilasi dan dijalankan secara mulus di dalam
// lingkungan kontainer terisolasi (seperti Docker standard) yang tidak memiliki hak istimewa kernel (CAP_SYS_ADMIN),
// sekaligus siap untuk diaktifkan penuh pada mesin bare-metal produksi.
func NewBpfManager() *BpfManager {
	return &BpfManager{
		mapName: "nexus_malicious_ips",
	}
}

// BlockIP mendaftarkan IP penyerang ke dalam tabel pemblokiran eBPF.
//
// Alasan Teknis (Why):
// Dalam implementasi Linux bare-metal, fungsi ini memanggil modul pustaka 'cilium/ebpf' untuk menyisipkan IP
// ke dalam tabel hash biner kernel (BPF_MAP_TYPE_HASH atau LPM_TRIE). eBPF Map ini dibaca secara real-time oleh
// modul XDP di kartu jaringan untuk memilah paket secara asinkron.
func (b *BpfManager) BlockIP(ip string) error {
	// Implementasi STUB untuk kompatibilitas multi-platform (macOS/Windows/Docker).
	log.Printf("[eBPF-KERNEL] (STUB) IP %s injected into eBPF map '%s'. Action: XDP_DROP", ip, b.mapName)
	return nil
}

// UnblockIP menghapus IP dari tabel pemblokiran eBPF untuk memulihkan hak akses trafik normal.
func (b *BpfManager) UnblockIP(ip string) error {
	log.Printf("[eBPF-KERNEL] (STUB) IP %s removed from eBPF map '%s'. Action: XDP_PASS", ip, b.mapName)
	return nil
}
