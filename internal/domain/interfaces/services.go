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
}

// ProductService defines product business operations
type ProductService interface {
	GetProduct(ctx context.Context, id uint) (*entities.Product, error)
	ListProducts(ctx context.Context, page, limit int) ([]entities.Product, int64, error)
	UpdateProduct(ctx context.Context, id uint, updates map[string]interface{}) (*entities.Product, error)
	UpdateStock(ctx context.Context, id uint, stock int) (*entities.Product, error)
}

// TransactionService defines transaction business operations
type TransactionService interface {
	CreateTransaction(ctx context.Context, req CreateTransactionRequest) (*entities.Transaction, error)
	GetTransaction(ctx context.Context, id uint) (*entities.Transaction, error)
	ListTransactions(ctx context.Context, page, limit int) ([]entities.Transaction, int64, error)
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
	User          string                   `json:"user"`
	PaymentMethod string                   `json:"payment_method"`
	Discount      float64                  `json:"discount"`
	TotalPrice    float64                  `json:"total_price"`
	Notes         string                   `json:"notes"`
}

// TransactionItemRequest represents an item in transaction request
type TransactionItemRequest struct {
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

// ReportResponse represents the sales report response
type ReportResponse struct {
	TotalRevenue       float64        `json:"total_revenue"`
	ItemsSold          int            `json:"items_sold"`
	AverageTransaction float64        `json:"average_transaction"`
	Details            []ReportDetail `json:"details"`
}
