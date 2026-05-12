package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base struct for all models to handle UUIDs and Timestamps
type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// ThreatLog maps to threat_logs table (Forensic Audit)
type ThreatLog struct {
	Base
	SourceIP      string `gorm:"type:varchar(45);index"`
	Endpoint      string `gorm:"type:varchar(255)"`
	Method        string `gorm:"type:varchar(10)"`
	Status        string `gorm:"type:varchar(50);index"`
	ThreatType    string `gorm:"type:varchar(100);index"`
	Severity      int    `gorm:"type:int"`
	PayloadSample string `gorm:"type:text"`
	UserAgent     string `gorm:"type:text"`
	LatencyMs     int    `gorm:"type:int"` // Custom extra column based on TelemetryLog
}

// MTDAuditTrail maps to mtd_audit_trail table (MTD State Changes)
type MTDAuditTrail struct {
	Base
	OldPort       int    `gorm:"type:int"`
	NewPort       int    `gorm:"type:int"`
	TriggerReason string `gorm:"type:varchar(100)"`
	Status        string `gorm:"type:varchar(50)"`
}

// IntelBlacklist maps to intel_blacklist table
type IntelBlacklist struct {
	Base
	IPAddress string     `gorm:"type:varchar(45);uniqueIndex"`
	Reason    string     `gorm:"type:varchar(255)"`
	ExpiresAt *time.Time `gorm:"type:timestamp"` // Nullable for permanent bans
	IsActive  bool       `gorm:"type:boolean;default:true"`
}

// AIInsight maps to ai_insights table
type AIInsight struct {
	Base
	ThreatLogID       uuid.UUID `gorm:"type:uuid;index"`
	ThreatLog         ThreatLog `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	AIModel           string    `gorm:"type:varchar(100)"`
	AnalysisText      string    `gorm:"type:text"`
	RecommendedAction string    `gorm:"type:varchar(255)"`
}
