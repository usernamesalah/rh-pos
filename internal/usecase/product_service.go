package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"github.com/usernamesalah/rh-pos/internal/pkg/storage"
	"github.com/usernamesalah/rh-pos/internal/pkg/storage/minio"
)

type productService struct {
	productRepo interfaces.ProductRepository
	storage     minio.StorageClient
	logger      *slog.Logger
}

// NewProductService creates a new product service
func NewProductService(productRepo interfaces.ProductRepository, storage minio.StorageClient, logger *slog.Logger) interfaces.ProductService {
	return &productService{
		productRepo: productRepo,
		storage:     storage,
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

	// Get tenant_id from context
	tenantID, ok := ctx.Value("tenant_id").(uint)
	if !ok {
		return nil, fmt.Errorf("tenant_id not found in context")
	}

	// Get existing product
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Update fields
	for field, value := range updates {
		switch field {
		case "image":
			product.Image = value.(string)
		case "name":
			product.Name = value.(string)
		case "sku":
			product.SKU = value.(string)
		case "harga_modal":
			product.HargaModal = value.(float64)
		case "harga_jual":
			product.HargaJual = value.(float64)
		case "stock":
			product.Stock = value.(int)
		}
	}

	// Ensure tenant_id is set
	product.TenantID = &tenantID

	// Save changes
	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

// UpdateStock updates product stock
func (s *productService) UpdateStock(ctx context.Context, id uint, stock int) (*entities.Product, error) {
	s.logger.InfoContext(ctx, "updating product stock", "id", id, "stock", stock)

	// Get tenant_id from context
	tenantID, ok := ctx.Value("tenant_id").(uint)
	if !ok {
		return nil, fmt.Errorf("tenant_id not found in context")
	}

	// Get existing product
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Update stock
	product.Stock = stock
	product.TenantID = &tenantID

	// Save changes
	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to update product stock: %w", err)
	}

	return product, nil
}

// CreateProduct creates a new product
func (s *productService) CreateProduct(ctx context.Context, product *entities.Product) error {
	s.logger.InfoContext(ctx, "creating product", "sku", product.SKU)

	// Get tenant_id from context
	tenantID, ok := ctx.Value("tenant_id").(uint)
	if !ok {
		return fmt.Errorf("tenant_id not found in context")
	}

	// Set tenant_id
	product.TenantID = &tenantID

	// Check if SKU already exists
	existingProduct, err := s.productRepo.GetBySKU(ctx, product.SKU)
	if err == nil && existingProduct != nil {
		return fmt.Errorf("product with SKU %s already exists", product.SKU)
	}

	// Create product
	if err := s.productRepo.Create(ctx, product); err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

// GetProductImageURL generates a presigned GET URL for the product image
func (s *productService) GetProductImageURL(ctx context.Context, product *entities.Product) (string, error) {
	if product.Image == "" {
		return "", nil
	}

	// Generate presigned GET URL with 1 hour expiry
	url, err := s.storage.GeneratePresignedURL(ctx, product.Image, time.Hour, false)
	if err != nil {
		return "", fmt.Errorf("failed to generate image URL: %w", err)
	}

	return url, nil
}

// GetProductUploadURL generates a presigned PUT URL for uploading a product image
func (s *productService) GetProductUploadURL(ctx context.Context, product *entities.Product, ext string) (string, error) {
	if product == nil {
		return "", fmt.Errorf("product is required")
	}

	// Generate image key
	key := storage.GenerateImageKey(product.ID, ext)

	// Generate presigned URL with 15 minutes expiry
	url, err := s.storage.GeneratePresignedURL(ctx, key, 15*time.Minute, true)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	// Update product with new image key
	updates := map[string]interface{}{
		"image": key,
	}
	if _, err := s.UpdateProduct(ctx, product.ID, updates); err != nil {
		return "", fmt.Errorf("failed to update product with image key: %w", err)
	}

	return url, nil
}
