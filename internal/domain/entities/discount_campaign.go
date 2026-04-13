package entities

import "time"

const (
	CampaignTypeProductPercentageDiscount  = "product_percentage_discount"
	CampaignTypeBuyXQtyGetDiscountAmount   = "buy_x_qty_get_discount_amount"
	CampaignTypeBuyXQtyGetDiscountPercent  = "buy_x_qty_get_discount_percentage"
	CampaignTypeBuyXProductGetYProductFree = "buy_x_product_get_y_product_free"
)

type DiscountCampaign struct {
	ID                 uint                      `json:"id" gorm:"primaryKey"`
	Name               string                    `json:"name" gorm:"not null"`
	CampaignType       string                    `json:"campaign_type" gorm:"not null;default:product_percentage_discount"`
	DiscountPercentage float64                   `json:"discount_percentage" gorm:"not null"`
	BuyQuantity        *int                      `json:"buy_quantity"`
	DiscountAmount     *float64                  `json:"discount_amount"`
	RewardProductID    *uint                     `json:"reward_product_id"`
	RewardProduct      *Product                  `json:"reward_product,omitempty" gorm:"foreignKey:RewardProductID"`
	RewardQuantity     *int                      `json:"reward_quantity"`
	StartDate          time.Time                 `json:"start_date" gorm:"not null"`
	EndDate            time.Time                 `json:"end_date" gorm:"not null"`
	TenantID           *uint                     `json:"tenant_id" gorm:"index"`
	Products           []DiscountCampaignProduct `json:"products,omitempty" gorm:"foreignKey:CampaignID"`
	CreatedAt          time.Time                 `json:"created_at"`
	UpdatedAt          time.Time                 `json:"updated_at"`
}

type DiscountCampaignProduct struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	CampaignID uint      `json:"campaign_id" gorm:"not null"`
	ProductID  uint      `json:"product_id" gorm:"not null"`
	Product    Product   `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	CreatedAt  time.Time `json:"created_at"`
}

func (DiscountCampaign) TableName() string {
	return "discount_campaigns"
}

func (DiscountCampaignProduct) TableName() string {
	return "discount_campaign_products"
}
