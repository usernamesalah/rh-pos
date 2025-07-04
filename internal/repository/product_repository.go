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

// Create creates a new product
func (r *productRepository) Create(ctx context.Context, product *entities.Product) error {
	r.logger.InfoContext(ctx, "creating product", "sku", product.SKU)
	if err := r.db.WithContext(ctx).Create(product).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to create product", "error", err)
		return fmt.Errorf("failed to create product: %w", err)
	}
	return nil
}

// Delete deletes a product
func (r *productRepository) Delete(ctx context.Context, id uint) error {
	r.logger.InfoContext(ctx, "deleting product", "id", id)
	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, ctx.Value("tenant_id")).Delete(&entities.Product{}).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to delete product", "error", err, "id", id)
		return fmt.Errorf("failed to delete product: %w", err)
	}
	return nil
}

// GetBySKU retrieves a product by SKU
func (r *productRepository) GetBySKU(ctx context.Context, sku string) (*entities.Product, error) {
	r.logger.InfoContext(ctx, "getting product by SKU", "sku", sku)
	var product entities.Product
	if err := r.db.WithContext(ctx).Where("sku = ? AND tenant_id = ?", sku, ctx.Value("tenant_id")).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("product not found: %w", err)
		}
		r.logger.ErrorContext(ctx, "failed to get product by SKU", "error", err, "sku", sku)
		return nil, fmt.Errorf("failed to get product by SKU: %w", err)
	}
	return &product, nil
}

// GetByID retrieves a product by ID
func (r *productRepository) GetByID(ctx context.Context, id uint) (*entities.Product, error) {
	r.logger.InfoContext(ctx, "getting product by ID", "id", id)

	var product entities.Product
	query := r.db.WithContext(ctx).Where("id = ?", id)

	// Add tenant_id filter if it exists in context
	if tenantID, ok := ctx.Value("tenant_id").(uint); ok {
		query = query.Where("tenant_id = ?", tenantID)
	} else {
		// If no tenant_id in context, only show products with NULL tenant_id
		query = query.Where("tenant_id IS NULL")
	}

	if err := query.First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("product not found: %w", err)
		}
		r.logger.ErrorContext(ctx, "failed to get product", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

// List retrieves all products with pagination
func (r *productRepository) List(ctx context.Context, page, limit int) ([]entities.Product, int64, error) {
	r.logger.InfoContext(ctx, "listing products", "page", page, "limit", limit)

	var products []entities.Product
	var total int64

	query := r.db.WithContext(ctx).Model(&entities.Product{})

	// Add tenant_id filter if it exists in context
	if tenantID, ok := ctx.Value("tenant_id").(uint); ok {
		query = query.Where("tenant_id = ?", tenantID)
	} else {
		// If no tenant_id in context, only show products with NULL tenant_id
		query = query.Where("tenant_id IS NULL")
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to count products", "error", err)
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to list products", "error", err)
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}

	return products, total, nil
}

// Update updates a product
func (r *productRepository) Update(ctx context.Context, product *entities.Product) error {
	r.logger.InfoContext(ctx, "updating product", "id", product.ID)

	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", product.ID, ctx.Value("tenant_id")).Save(product).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to update product", "error", err, "id", product.ID)
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

// UpdateStock updates product stock
func (r *productRepository) UpdateStock(ctx context.Context, id uint, stock int) error {
	r.logger.InfoContext(ctx, "updating product stock", "id", id, "stock", stock)

	if err := r.db.WithContext(ctx).Model(&entities.Product{}).Where("id = ? AND tenant_id = ?", id, ctx.Value("tenant_id")).Update("stock", stock).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to update product stock", "error", err, "id", id)
		return fmt.Errorf("failed to update product stock: %w", err)
	}

	return nil
}
