package entities

import "time"

// DiscountCampaign represents a time-bound discount campaign for products
type DiscountCampaign struct {
	ID                 uint                      `json:"id" gorm:"primaryKey"`
	Name               string                    `json:"name" gorm:"not null"`
	DiscountPercentage float64                   `json:"discount_percentage" gorm:"not null"`
	StartDate          time.Time                 `json:"start_date" gorm:"not null"`
	EndDate            time.Time                 `json:"end_date" gorm:"not null"`
	TenantID           *uint                     `json:"tenant_id" gorm:"index"`
	Products           []DiscountCampaignProduct `json:"products,omitempty" gorm:"foreignKey:CampaignID"`
	CreatedAt          time.Time                 `json:"created_at"`
	UpdatedAt          time.Time                 `json:"updated_at"`
}

// DiscountCampaignProduct is the join table between campaigns and products
type DiscountCampaignProduct struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	CampaignID uint      `json:"campaign_id" gorm:"not null"`
	ProductID  uint      `json:"product_id" gorm:"not null"`
	Product    Product   `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	CreatedAt  time.Time `json:"created_at"`
}

// TableName sets the table name for GORM
func (DiscountCampaign) TableName() string {
	return "discount_campaigns"
}

// TableName sets the table name for GORM
func (DiscountCampaignProduct) TableName() string {
	return "discount_campaign_products"
}
