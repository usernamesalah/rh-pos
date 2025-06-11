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
	TenantID  *uint     `json:"tenant_id" gorm:"index"`
	Tenant    *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName sets the table name for GORM
func (User) TableName() string {
	return "users"
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(user *User) error
	GetByID(id uint) (*User, error)
	GetByUsername(username string) (*User, error)
	List() ([]*User, error)
	Update(user *User) error
	Delete(id uint) error
}

// UserUseCase defines the interface for user business logic
type UserUseCase interface {
	Create(user *User) error
	GetByID(id uint) (*User, error)
	GetByUsername(username string) (*User, error)
	List() ([]*User, error)
	Update(user *User) error
	Delete(id uint) error
}
