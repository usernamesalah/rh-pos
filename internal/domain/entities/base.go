package entities

import (
	"time"

	"gorm.io/gorm"
)

// Base contains common columns for all tables
type Base struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (base *Base) BeforeCreate(tx *gorm.DB) error {
	base.CreatedAt = time.Now()
	base.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate will update the UpdatedAt timestamp
func (base *Base) BeforeUpdate(tx *gorm.DB) error {
	base.UpdatedAt = time.Now()
	return nil
}
