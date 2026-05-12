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

var DB *gorm.DB

// InitPostgres initializes the PostgreSQL connection and runs auto-migrations
func InitPostgres() {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		log.Println("[DB-WARNING] POSTGRES_DSN is not set. Database persistence is disabled.")
		return
	}

	// Connect to PostgreSQL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn), // Only log warnings and errors to keep console clean
	})
	if err != nil {
		log.Fatalf("[DB-ERROR] Failed to connect to PostgreSQL: %v", err)
	}

	log.Println("[DB-INIT] Successfully connected to PostgreSQL.")

	// Auto-Migrate the schema
	log.Println("[DB-INIT] Running Auto-Migrations for ISO 27001 Schema...")
	err = db.AutoMigrate(
		&models.ThreatLog{},
		&models.MTDAuditTrail{},
		&models.IntelBlacklist{},
		&models.AIInsight{},
	)
	if err != nil {
		log.Fatalf("[DB-ERROR] Failed to run migrations: %v", err)
	}

	log.Println("[DB-INIT] Auto-Migrations completed successfully.")
	DB = db
}

// IsIPBlacklisted checks if an IP is in the active blacklist and not expired
func IsIPBlacklisted(ip string) bool {
	if DB == nil {
		return false
	}

	var blacklist models.IntelBlacklist
	now := time.Now()
	
	// Strip port from IP if present (e.g. "127.0.0.1:12345" -> "127.0.0.1")
	if idx := strings.Index(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	result := DB.Where("ip_address = ? AND is_active = true AND (expires_at IS NULL OR expires_at > ?)", ip, now).First(&blacklist)
	return result.Error == nil
}

// SaveAIInsight persists the reasoning result from Llama/Qwen
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
