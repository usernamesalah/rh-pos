package entities

import "time"

type Category struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	TenantID  *uint     `json:"tenant_id" gorm:"index"`
	Tenant    *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Category) TableName() string {
	return "categories"
}
