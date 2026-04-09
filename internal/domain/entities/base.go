package entities

import (
	"time"

	"gorm.io/gorm"
)

// Base contains common columns for all tables.
// GORM automatically manages CreatedAt and UpdatedAt — no hooks needed.
type Base struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
