package entities

import (
	"encoding/json"
	"time"
)

// AuditLog records who changed what entity and when.
// Rows are immutable — no UpdatedAt or soft delete.
type AuditLog struct {
	ID          uint            `json:"id" gorm:"primaryKey"`
	TenantID    uint            `json:"tenant_id" gorm:"not null;index"`
	UserID      uint            `json:"user_id" gorm:"not null;index"`
	EntityType  string          `json:"entity_type" gorm:"not null;size:50"`
	EntityID    uint            `json:"entity_id" gorm:"not null"`
	Action      string          `json:"action" gorm:"not null;size:20"`
	BeforeState json.RawMessage `json:"before_state,omitempty" gorm:"type:json"`
	AfterState  json.RawMessage `json:"after_state,omitempty" gorm:"type:json"`
	CreatedAt   time.Time       `json:"created_at"`
}

// TableName sets the table name for GORM.
func (AuditLog) TableName() string {
	return "audit_logs"
}
