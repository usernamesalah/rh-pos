package entities

import (
	"time"
)

type Transaction struct {
	ID            uint              `json:"id" gorm:"primaryKey"`
	Items         []TransactionItem `json:"items" gorm:"foreignKey:TransactionID"`
	UserID        *uint             `json:"user_id" gorm:"index"`
	User          *User             `json:"user,omitempty" gorm:"foreignKey:UserID"`
	PaymentMethod string            `json:"payment_method" gorm:"not null"`
	Discount      float64           `json:"discount" gorm:"default:0"`
	TotalPrice    float64           `json:"total_price" gorm:"not null"`
	TenantID      *uint             `json:"tenant_id" gorm:"index"`
	Tenant        *Tenant           `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	CustomerName  *string           `json:"customer_name,omitempty"`
	CustomerEmail *string           `json:"customer_email,omitempty"`
	CustomerPhone *string           `json:"customer_phone,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	Notes         string            `json:"notes,omitempty" gorm:"type:text"`
}

type TransactionItem struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	TransactionID      uint      `json:"transaction_id" gorm:"not null"`
	ProductID          uint      `json:"product_id" gorm:"not null"`
	Product            Product   `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	Quantity           int       `json:"quantity" gorm:"not null"`
	Price              float64   `json:"price" gorm:"not null"`
	WarrantyDays       int       `json:"warranty_days" gorm:"default:0"`
	DiscountPercentage float64   `json:"discount_percentage" gorm:"default:0"`
	DiscountAmount     float64   `json:"discount_amount" gorm:"default:0"`
	CampaignID         *uint     `json:"campaign_id,omitempty"`
	CampaignType       string    `json:"campaign_type,omitempty"`
	IsFreeItem         bool      `json:"is_free_item" gorm:"default:false"`
	CampaignGroupKey   string    `json:"campaign_group_key,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (Transaction) TableName() string {
	return "transactions"
}

func (TransactionItem) TableName() string {
	return "transaction_items"
}
