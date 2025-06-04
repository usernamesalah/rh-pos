package entities

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"`
	Role      string    `json:"role" gorm:"not null;default:'user'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName sets the table name for GORM
func (User) TableName() string {
	return "users"
}
