package interfaces

import (
	"context"
	"time"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
)

// AuthService defines authentication operations
type AuthService interface {
	Login(ctx context.Context, username, password string) (string, *entities.User, error)
	ValidateToken(tokenString string) (*entities.User, error)
	HashPassword(password string) (string, error)
	GetUserByID(ctx context.Context, id uint) (*entities.User, error)
	CreateUser(ctx context.Context, user *entities.User) error
	UpdatePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error
	ListUsers(ctx context.Context, page, limit int) ([]*entities.User, int64, error)
	UpdateUser(ctx context.Context, user *entities.User) error
	DeleteUser(ctx context.Context, id uint) error
}

// ProductService defines product business operations
type ProductService interface {
	GetProduct(ctx context.Context, id uint) (*entities.Product, error)
	ListProducts(ctx context.Context, page, limit int) ([]entities.Product, int64, error)
	UpdateProduct(ctx context.Context, id uint, updates map[string]interface{}) (*entities.Product, error)
	UpdateStock(ctx context.Context, id uint, stock int) (*entities.Product, error)
	CreateProduct(ctx context.Context, product *entities.Product) error
	DeleteProduct(ctx context.Context, id uint) error
	GetProductImageURL(ctx context.Context, product *entities.Product) (string, error)
	// GetProductUploadURL returns a presigned PUT URL and the generated image key.
	// The caller is responsible for persisting the key after a successful upload.
	GetProductUploadURL(ctx context.Context, product *entities.Product, ext string) (uploadURL string, imageKey string, err error)
	UploadProductImage(ctx context.Context, productID uint, fileData []byte, contentType string) (*entities.Product, error)
	GetProductImageBytes(ctx context.Context, productID uint) ([]byte, string, error)
}

// TransactionService defines transaction business operations
type TransactionService interface {
	CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*entities.Transaction, error)
	GetTransaction(ctx context.Context, id uint) (*entities.Transaction, error)
	ListTransactions(ctx context.Context, page, limit int) ([]entities.Transaction, int64, error)
}

// DiscountCampaignService defines discount campaign business operations
type DiscountCampaignService interface {
	Create(ctx context.Context, req CreateCampaignRequest) (*entities.DiscountCampaign, error)
	GetByID(ctx context.Context, id uint) (*entities.DiscountCampaign, error)
	List(ctx context.Context, page, limit int) ([]entities.DiscountCampaign, int64, error)
	Update(ctx context.Context, id uint, req UpdateCampaignRequest) (*entities.DiscountCampaign, error)
	Delete(ctx context.Context, id uint) error
	AddProducts(ctx context.Context, campaignID uint, productIDs []uint) error
	RemoveProduct(ctx context.Context, campaignID uint, productID uint) error
}

// WarrantyService defines warranty check operations
type WarrantyService interface {
	CheckWarranty(ctx context.Context, hashedTransactionID string) (*WarrantyResponse, error)
	SearchByPhone(ctx context.Context, phone string) ([]WarrantyResponse, error)
}

// ReportService defines reporting operations
type ReportService interface {
	GetSalesReport(ctx context.Context, startDate, endDate time.Time) (*ReportResponse, error)
}

// TenantService defines tenant business operations
type TenantService interface {
	CreateTenant(ctx context.Context, tenant *entities.Tenant) error
	GetTenant(ctx context.Context, id uint) (*entities.Tenant, error)
	ListTenants(ctx context.Context) ([]*entities.Tenant, error)
	UpdateTenant(ctx context.Context, tenant *entities.Tenant) error
	DeleteTenant(ctx context.Context, id uint) error
}

// CreateTransactionRequest represents the request to create a transaction
type CreateTransactionRequest struct {
	Items         []TransactionItemRequest `json:"items"`
	UserID        uint                     `json:"user_id"`
	PaymentMethod string                   `json:"payment_method"`
	Discount      float64                  `json:"discount"`
	TotalPrice    float64                  `json:"total_price"`
	Notes         string                   `json:"notes"`
	CustomerName  *string                  `json:"customer_name"`
	CustomerEmail *string                  `json:"customer_email"`
	CustomerPhone *string                  `json:"customer_phone"`
}

// TransactionItemRequest represents an item in transaction request
type TransactionItemRequest struct {
	ProductID    uint `json:"product_id"`
	Quantity     int  `json:"quantity"`
	WarrantyDays int  `json:"warranty_days"`
}

// CreateCampaignRequest represents the request to create a discount campaign
type CreateCampaignRequest struct {
	Name               string    `json:"name"`
	DiscountPercentage float64   `json:"discount_percentage"`
	StartDate          time.Time `json:"start_date"`
	EndDate            time.Time `json:"end_date"`
	ProductIDs         []uint    `json:"product_ids"`
}

// UpdateCampaignRequest represents the request to update a discount campaign
type UpdateCampaignRequest struct {
	Name               *string    `json:"name"`
	DiscountPercentage *float64   `json:"discount_percentage"`
	StartDate          *time.Time `json:"start_date"`
	EndDate            *time.Time `json:"end_date"`
}

// WarrantyResponse represents the public warranty check response
type WarrantyResponse struct {
	TransactionID   string                 `json:"transaction_id"`
	TransactionDate time.Time              `json:"transaction_date"`
	CustomerName    string                 `json:"customer_name,omitempty"`
	CustomerEmail   string                 `json:"customer_email,omitempty"`
	CustomerPhone   string                 `json:"customer_phone,omitempty"`
	Items           []WarrantyItemResponse `json:"items"`
}

// WarrantyItemResponse represents a single item in warranty response
type WarrantyItemResponse struct {
	ProductName   string    `json:"product_name"`
	Quantity      int       `json:"quantity"`
	WarrantyDays  int       `json:"warranty_days"`
	WarrantyStart time.Time `json:"warranty_start"`
	WarrantyEnd   time.Time `json:"warranty_end"`
	IsActive      bool      `json:"is_active"`
	DaysRemaining int       `json:"days_remaining"`
}

// ReportResponse represents the sales report response
type ReportResponse struct {
	TotalRevenue       float64        `json:"total_revenue"`
	ItemsSold          int            `json:"items_sold"`
	AverageTransaction float64        `json:"average_transaction"`
	Details            []ReportDetail `json:"details"`
}
