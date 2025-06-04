package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
)

type productService struct {
	productRepo interfaces.ProductRepository
	logger      *slog.Logger
}

// NewProductService creates a new product service
func NewProductService(productRepo interfaces.ProductRepository, logger *slog.Logger) interfaces.ProductService {
	return &productService{
		productRepo: productRepo,
		logger:      logger,
	}
}

// GetProduct retrieves a product by ID
func (s *productService) GetProduct(ctx context.Context, id uint) (*entities.Product, error) {
	s.logger.InfoContext(ctx, "getting product", "id", id)

	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

// ListProducts retrieves products with pagination
func (s *productService) ListProducts(ctx context.Context, page, limit int) ([]entities.Product, int64, error) {
	s.logger.InfoContext(ctx, "listing products", "page", page, "limit", limit)

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	products, total, err := s.productRepo.List(ctx, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}

	return products, total, nil
}

// UpdateProduct updates a product with the provided fields
func (s *productService) UpdateProduct(ctx context.Context, id uint, updates map[string]interface{}) (*entities.Product, error) {
	s.logger.InfoContext(ctx, "updating product", "id", id)

	// Get existing product
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Update fields
	if image, ok := updates["image"].(string); ok {
		product.Image = image
	}
	if name, ok := updates["name"].(string); ok {
		product.Name = name
	}
	if sku, ok := updates["sku"].(string); ok {
		product.SKU = sku
	}
	if hargaModal, ok := updates["harga_modal"].(float64); ok {
		product.HargaModal = hargaModal
	}
	if hargaJual, ok := updates["harga_jual"].(float64); ok {
		product.HargaJual = hargaJual
	}
	if stock, ok := updates["stock"].(int); ok {
		product.Stock = stock
	}

	// Save updated product
	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

// UpdateStock updates the stock of a product
func (s *productService) UpdateStock(ctx context.Context, id uint, stock int) (*entities.Product, error) {
	s.logger.InfoContext(ctx, "updating product stock", "id", id, "stock", stock)

	// Validate stock
	if stock < 0 {
		return nil, fmt.Errorf("stock cannot be negative")
	}

	// Update stock
	if err := s.productRepo.UpdateStock(ctx, id, stock); err != nil {
		return nil, fmt.Errorf("failed to update stock: %w", err)
	}

	// Return updated product
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated product: %w", err)
	}

	return product, nil
}
