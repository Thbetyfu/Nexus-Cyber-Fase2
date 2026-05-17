// Package models mendefinisikan skema tabel relational database (ORM) untuk Nexus Cyber SOC.
// Model ini dirancang khusus untuk mematuhi regulasi ISO 27001 (Kontrol Keamanan Informasi)
// dan UU PDP No. 27/2022 guna memastikan integritas data telemetri.
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base bertindak sebagai fondasi untuk seluruh model database dengan menangani pengidentifikasi unik dan stempel waktu.
//
// Alasan Arsitektural (Why):
// Menggunakan UUID v4 secara default (`gen_random_uuid()`) alih-alih ID Integer berurutan.
// Hal ini mencegah serangan ID Enumeration (peretas memetakan total log kita dengan menebak angka berurutan)
// serta menjamin tidak ada konflik ID saat melakukan merger database multi-node secara terdistribusi.
type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"` // Menggunakan Soft Delete untuk mematuhi retensi data investigasi forensik
}

// ThreatLog memetakan struktur tabel `threat_logs` untuk menyimpan rekam jejak ancaman (Forensic Audit).
//
// Alasan Teknis (Why):
// Kolom SourceIP, Status, dan ThreatType di-indeks secara eksplisit (`index`) untuk mempercepat query
// hingga sub-milidetik ketika dashboard NCC memproses visualisasi peta ancaman 3D real-time berarus tinggi.
type ThreatLog struct {
	Base
	SourceIP      string `gorm:"type:varchar(45);index"` // Mendukung IPv4 dan IPv6 (maksimal 45 karakter)
	Endpoint      string `gorm:"type:varchar(255)"`
	Method        string `gorm:"type:varchar(10)"`
	Status        string `gorm:"type:varchar(50);index"`
	ThreatType    string `gorm:"type:varchar(100);index"`
	Severity      int    `gorm:"type:int"`
	PayloadSample string `gorm:"type:text"`
	UserAgent     string `gorm:"type:text"`
	LatencyMs     int    `gorm:"type:int"` // Latensi pemrosesan internal gateway
}

// MTDAuditTrail memetakan tabel `mtd_audit_trail` untuk merekam rotasi konfigurasi pertahanan dinamis (MTD).
// Memastikan setiap perubahan port backend terdokumentasi lengkap untuk audit regulator (BSSN/OJK).
type MTDAuditTrail struct {
	Base
	OldPort       int    `gorm:"type:int"`
	NewPort       int    `gorm:"type:int"`
	TriggerReason string `gorm:"type:varchar(100)"` // SCHEDULED_ROTATION atau EMERGENCY_MANUAL_SHUFFLE
	Status        string `gorm:"type:varchar(50)"`
}

// IntelBlacklist memetakan tabel `intel_blacklist` untuk memblokir IP penyerang secara persisten.
type IntelBlacklist struct {
	Base
	IPAddress string     `gorm:"type:varchar(45);uniqueIndex"`
	Reason    string     `gorm:"type:varchar(255)"`
	ExpiresAt *time.Time `gorm:"type:timestamp"` // Nullable: Jika NULL, maka pemblokiran bersifat permanen (Permanent Ban)
	IsActive  bool       `gorm:"type:boolean;default:true"`
}

// AIInsight menyimpan analisis kecerdasan buatan mendalam yang di-eskalasi dari Reflex Layer.
//
// Alasan Arsitektural (Why):
// Memiliki relasi One-to-One dengan ThreatLogID melalui constraint kunci asing (foreign key) CASCADE.
// Jika log dihapus, data analisis AI yang berelasi akan disesuaikan secara otomatis untuk integritas relasional.
type AIInsight struct {
	Base
	ThreatLogID       uuid.UUID `gorm:"type:uuid;index"`
	ThreatLog         ThreatLog `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	AIModel           string    `gorm:"type:varchar(100)"` // Qwen/Qwen3-235B atau Llama3
	AnalysisText      string    `gorm:"type:text"`         // Ulasan intensi peretas dan analisa APT
	RecommendedAction string    `gorm:"type:varchar(255)"` // Mitigasi spesifik (misal: "BLOCK_IP")
}
