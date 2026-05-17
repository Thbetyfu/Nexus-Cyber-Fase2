// Package database mengelola koneksi persistensi PostgreSQL untuk menyimpan telemetri keamanan dan audit trail.
// Mematuhi standar ISO 27001 (Kontrol A.12.4 - Logging dan Pemantauan) untuk memastikan log audit siber
// disimpan secara permanen, terstruktur, dan tidak dapat dimanipulasi dengan mudah.
package database

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-cyber/nexus-core-gateway/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB adalah referensi singleton global untuk koneksi database GORM PostgreSQL.
var DB *gorm.DB

// InitPostgres menginisialisasi pool koneksi database PostgreSQL dan menjalankan migrasi skema otomatis.
//
// Alasan Arsitektural (Why):
// - Jika environment `POSTGRES_DSN` kosong, sistem mengalami degradasi anggun (degraded mode) tanpa crash,
//   memungkinkan gateway beroperasi dalam mode in-memory/cache (ISO 25010 - Fault Tolerance).
// - Menggunakan Auto-Migrate untuk memastikan tabel audit trail penting seperti ThreatLog dan MTDAuditTrail
//   selalu sinkron dengan struktur data terbaru saat gateway pertama kali dijalankan.
func InitPostgres() {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		log.Println("[DB-WARNING] POSTGRES_DSN is not set. Database persistence is disabled (Degraded Local Mode).")
		return
	}

	// Membuka koneksi pool dengan logger Warn untuk menghemat I/O disk dari pencatatan query SELECT yang berlebihan.
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("[DB-ERROR] Failed to connect to PostgreSQL: %v", err)
	}

	log.Println("[DB-INIT] Successfully connected to PostgreSQL.")

	// Auto-Migrate Tabel Forensik dan Log Keamanan untuk pemenuhan standar kepatuhan BSSN & OJK (ISO 27001).
	log.Println("[DB-INIT] Running Auto-Migrations for ISO 27001 Schema...")
	err = db.AutoMigrate(
		&models.ThreatLog{},
		&models.MTDAuditTrail{},
		&models.IntelBlacklist{},
		&models.AIInsight{},
		&models.DomainSubscription{},
	)
	if err != nil {
		log.Fatalf("[DB-ERROR] Failed to run migrations: %v", err)
	}

	log.Println("[DB-INIT] Auto-Migrations completed successfully.")
	DB = db
}

// IsIPBlacklisted memeriksa apakah IP penyerang terdaftar dalam daftar hitam (blacklist) yang masih aktif.
//
// Alasan Teknis (Why):
// Penyerang sering memalsukan port sumber (source port) untuk melewati pemeriksaan keamanan.
// Fungsi ini melakukan normalisasi IP (stripping port) dengan membuang tanda titik dua ":" dan nomor port di belakangnya
// sebelum melakukan query. Ini menjamin pemblokiran IP bersifat mutlak tanpa peduli port mana yang digunakan peretas.
func IsIPBlacklisted(ip string) bool {
	if DB == nil {
		return false
	}

	var blacklist models.IntelBlacklist
	now := time.Now()
	
	// Normalisasi IP: Potong port jika ada (misal "192.168.1.10:49281" -> "192.168.1.10")
	if idx := strings.Index(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	// Query dioptimalkan dengan memverifikasi masa berlaku blacklist (expires_at) secara real-time.
	result := DB.Where("ip_address = ? AND is_active = true AND (expires_at IS NULL OR expires_at > ?)", ip, now).First(&blacklist)
	return result.Error == nil
}

// SaveAIInsight menyimpan hasil analisis forensik kustom dari Llama/Qwen ke database.
//
// Alasan Arsitektural (Why):
// Hasil pemikiran AI (AI Insight) disimpan dalam tabel terpisah yang berelasi One-to-One dengan ThreatLog.
// Pemisahan ini mempermudah audit investigasi insiden siber secara spesifik tanpa memperlambat pembacaan log trafik utama.
func SaveAIInsight(logID uuid.UUID, modelName, analysis, recommendation string) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	insight := models.AIInsight{
		ThreatLogID:       logID,
		AIModel:           modelName,
		AnalysisText:      analysis,
		RecommendedAction: recommendation,
	}

	return DB.Create(&insight).Error
}
