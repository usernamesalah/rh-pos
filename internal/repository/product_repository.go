package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"gorm.io/gorm"
)

type productRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *gorm.DB, logger *slog.Logger) interfaces.ProductRepository {
	return &productRepository{
		db:     db,
		logger: logger,
	}
}

// GetByID retrieves a product by ID
func (r *productRepository) GetByID(ctx context.Context, id uint) (*entities.Product, error) {
	r.logger.InfoContext(ctx, "getting product by ID", "id", id)

	var product entities.Product
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("product not found: %w", err)
		}
		r.logger.ErrorContext(ctx, "failed to get product", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

// List retrieves products with pagination
func (r *productRepository) List(ctx context.Context, page, limit int) ([]entities.Product, int64, error) {
	r.logger.InfoContext(ctx, "listing products", "page", page, "limit", limit)

	var products []entities.Product
	var total int64

	// Count total products
	if err := r.db.WithContext(ctx).Model(&entities.Product{}).Count(&total).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to count products", "error", err)
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Get products with pagination
	offset := (page - 1) * limit
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to list products", "error", err)
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}

	return products, total, nil
}

// Update updates a product
func (r *productRepository) Update(ctx context.Context, product *entities.Product) error {
	r.logger.InfoContext(ctx, "updating product", "id", product.ID)

	if err := r.db.WithContext(ctx).Save(product).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to update product", "error", err, "id", product.ID)
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

// UpdateStock updates product stock
func (r *productRepository) UpdateStock(ctx context.Context, id uint, stock int) error {
	r.logger.InfoContext(ctx, "updating product stock", "id", id, "stock", stock)

	if err := r.db.WithContext(ctx).Model(&entities.Product{}).Where("id = ?", id).Update("stock", stock).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to update product stock", "error", err, "id", id)
		return fmt.Errorf("failed to update product stock: %w", err)
	}

	return nil
}
