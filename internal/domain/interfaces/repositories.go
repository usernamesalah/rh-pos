package interfaces

import (
	"context"
	"time"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	GetByUsername(ctx context.Context, username string) (*entities.User, error)
	GetByID(ctx context.Context, id uint) (*entities.User, error)
	Create(ctx context.Context, user *entities.User) error
	List(ctx context.Context) ([]*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id uint) error
}

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
	GetByID(ctx context.Context, id uint) (*entities.Product, error)
	List(ctx context.Context, page, limit int) ([]entities.Product, int64, error)
	Update(ctx context.Context, product *entities.Product) error
	UpdateStock(ctx context.Context, id uint, stock int) error
	Create(ctx context.Context, product *entities.Product) error
	GetBySKU(ctx context.Context, sku string) (*entities.Product, error)
	Delete(ctx context.Context, id uint) error
}

// TransactionRepository defines the interface for transaction data operations
type TransactionRepository interface {
	Create(ctx context.Context, transaction *entities.Transaction) error
	GetByID(ctx context.Context, id uint) (*entities.Transaction, error)
	List(ctx context.Context, page, limit int) ([]entities.Transaction, int64, error)
	GetReportData(ctx context.Context, startDate, endDate time.Time) ([]ReportDetail, error)
	Update(ctx context.Context, transaction *entities.Transaction) error
	Delete(ctx context.Context, id uint) error
}

// TenantRepository defines the interface for tenant data operations
type TenantRepository interface {
	Create(ctx context.Context, tenant *entities.Tenant) error
	GetByID(ctx context.Context, id uint) (*entities.Tenant, error)
	List(ctx context.Context) ([]*entities.Tenant, error)
	Update(ctx context.Context, tenant *entities.Tenant) error
	Delete(ctx context.Context, id uint) error
}

// ReportDetail represents report data structure
type ReportDetail struct {
	ID          uint    `json:"id"`
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	Total       int     `json:"total"`
	TotalPrice  float64 `json:"total_price"`
}
