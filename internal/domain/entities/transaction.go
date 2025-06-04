package entities

import (
	"time"
)

// Transaction represents a sales transaction
type Transaction struct {
	ID            uint              `json:"id" gorm:"primaryKey"`
	Items         []TransactionItem `json:"items" gorm:"foreignKey:TransactionID"`
	User          string            `json:"user" gorm:"not null"`
	PaymentMethod string            `json:"payment_method" gorm:"not null"`
	Discount      float64           `json:"discount" gorm:"default:0"`
	TotalPrice    float64           `json:"total_price" gorm:"not null"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// TransactionItem represents an item in a transaction
type TransactionItem struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	TransactionID uint      `json:"transaction_id" gorm:"not null"`
	ProductID     uint      `json:"product_id" gorm:"not null"`
	Product       Product   `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	Quantity      int       `json:"quantity" gorm:"not null"`
	Price         float64   `json:"price" gorm:"not null"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName sets the table name for GORM
func (Transaction) TableName() string {
	return "transactions"
}

// TableName sets the table name for GORM
func (TransactionItem) TableName() string {
	return "transaction_items"
}
