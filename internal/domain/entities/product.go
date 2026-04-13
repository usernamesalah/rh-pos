package entities

import (
	"time"
)

// Product represents a product in the system
type Product struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	Image      string    `json:"image"`
	Name       string    `json:"name" gorm:"not null"`
	SKU        *string   `json:"sku" gorm:"uniqueIndex:idx_tenant_sku"`
	HargaModal *float64  `json:"harga_modal"`
	HargaJual  *float64  `json:"harga_jual"`
	Stock      int       `json:"stock" gorm:"not null;default:0"`
	TenantID   *uint     `json:"tenant_id" gorm:"uniqueIndex:idx_tenant_sku;index"`
	Tenant     *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName sets the table name for GORM
func (Product) TableName() string {
	return "products"
}
