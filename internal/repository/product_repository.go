package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"github.com/usernamesalah/rh-pos/internal/pkg/ctxkey"
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

	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant_id not found in context")
	}

	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&entities.Product{}).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to delete product", "error", err, "id", id)
		return fmt.Errorf("failed to delete product: %w", err)
	}
	return nil
}

// GetBySKU retrieves a product by SKU
func (r *productRepository) GetBySKU(ctx context.Context, sku string) (*entities.Product, error) {
	r.logger.InfoContext(ctx, "getting product by SKU", "sku", sku)

	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant_id not found in context")
	}

	var product entities.Product
	if err := r.db.WithContext(ctx).Where("sku = ? AND tenant_id = ?", sku, tenantID).Preload("Category").First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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

	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant_id not found in context")
	}

	var product entities.Product
	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Preload("Category").First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("product not found: %w", err)
		}
		r.logger.ErrorContext(ctx, "failed to get product", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

// List retrieves all products with pagination and optional search
func (r *productRepository) List(ctx context.Context, page, limit int, search string, categoryID *uint) ([]entities.Product, int64, error) {
	r.logger.InfoContext(ctx, "listing products", "page", page, "limit", limit, "search", search)

	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return nil, 0, fmt.Errorf("tenant_id not found in context")
	}

	var products []entities.Product
	var total int64

	query := r.db.WithContext(ctx).Model(&entities.Product{}).Where("tenant_id = ?", tenantID)

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name LIKE ?", searchPattern)
	}

	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}

	if err := query.Count(&total).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to count products", "error", err)
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	offset := (page - 1) * limit
	if err := query.Preload("Category").Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to list products", "error", err)
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}

	return products, total, nil
}

// Update updates a product
func (r *productRepository) Update(ctx context.Context, product *entities.Product) error {
	r.logger.InfoContext(ctx, "updating product", "id", product.ID)

	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant_id not found in context")
	}

	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", product.ID, tenantID).
		Select("Image", "Name", "SKU", "HargaModal", "HargaJual", "Stock", "IsDynamicPrice", "CategoryID", "UpdatedAt").
		Updates(product).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to update product", "error", err, "id", product.ID)
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

// UpdateStock updates product stock
func (r *productRepository) UpdateStock(ctx context.Context, id uint, stock int) error {
	r.logger.InfoContext(ctx, "updating product stock", "id", id, "stock", stock)

	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant_id not found in context")
	}

	if err := r.db.WithContext(ctx).Model(&entities.Product{}).Where("id = ? AND tenant_id = ?", id, tenantID).Update("stock", stock).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to update product stock", "error", err, "id", id)
		return fmt.Errorf("failed to update product stock: %w", err)
	}

	return nil
}
