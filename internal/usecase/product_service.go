package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"path"
	"strings"
	"time"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"github.com/usernamesalah/rh-pos/internal/pkg/ctxkey"
	"github.com/usernamesalah/rh-pos/internal/pkg/storage"
)

type productService struct {
	productRepo interfaces.ProductRepository
	storage     storage.StorageClient
	baseURL     string
	logger      *slog.Logger
}

// NewProductService creates a new product service
func NewProductService(productRepo interfaces.ProductRepository, storage storage.StorageClient, baseURL string, logger *slog.Logger) interfaces.ProductService {
	return &productService{
		productRepo: productRepo,
		storage:     storage,
		baseURL:     baseURL,
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

// ListProducts retrieves products with pagination and optional search
func (s *productService) ListProducts(ctx context.Context, page, limit int, search string, categoryID *uint) ([]entities.Product, int64, error) {
	s.logger.InfoContext(ctx, "listing products", "page", page, "limit", limit, "search", search)

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	products, total, err := s.productRepo.List(ctx, page, limit, search, categoryID)
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
	for field, value := range updates {
		switch field {
		case "image":
			product.Image = value.(string)
		case "name":
			product.Name = value.(string)
		case "sku":
			if v, ok := value.(string); ok {
				product.SKU = &v
			}
		case "harga_modal":
			if v, ok := value.(float64); ok {
				product.HargaModal = &v
			}
		case "harga_jual":
			if v, ok := value.(float64); ok {
				product.HargaJual = &v
			}
		case "stock":
			product.Stock = value.(int)
		case "CategoryID":
			if v, ok := value.(uint); ok {
				product.CategoryID = &v
			} else {
				product.CategoryID = nil
			}
		case "is_dynamic_price":
			if v, ok := value.(bool); ok {
				product.IsDynamicPrice = v
			}
		}
	}

	// Save changes
	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

// UpdateStock updates product stock
func (s *productService) UpdateStock(ctx context.Context, id uint, stock int) (*entities.Product, error) {
	s.logger.InfoContext(ctx, "updating product stock", "id", id, "stock", stock)

	// Get existing product
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Update stock
	product.Stock = stock

	// Save changes
	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to update product stock: %w", err)
	}

	return product, nil
}

// CreateProduct creates a new product
func (s *productService) CreateProduct(ctx context.Context, product *entities.Product) error {
	var skuStr string
	if product.SKU != nil {
		skuStr = *product.SKU
	}
	s.logger.InfoContext(ctx, "creating product", "sku", skuStr)

	// Get tenant_id from context
	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant_id not found in context")
	}

	// Set tenant_id
	product.TenantID = &tenantID

	// Check for duplicate SKU only when SKU is provided
	if product.SKU != nil {
		existingProduct, err := s.productRepo.GetBySKU(ctx, *product.SKU)
		if err == nil && existingProduct != nil {
			return fmt.Errorf("product with SKU %s already exists", *product.SKU)
		}
	}

	// Dynamic price products don't require harga_jual/harga_modal
	if !product.IsDynamicPrice {
		switch {
		case product.HargaJual != nil && product.HargaModal == nil:
			v := *product.HargaJual
			product.HargaModal = &v
		case product.HargaModal != nil && product.HargaJual == nil:
			v := *product.HargaModal
			product.HargaJual = &v
		case product.HargaJual == nil && product.HargaModal == nil:
			return fmt.Errorf("at least one of harga_jual or harga_modal must be provided")
		}
	}

	// Create product
	if err := s.productRepo.Create(ctx, product); err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

func (s *productService) DeleteProduct(ctx context.Context, id uint) error {
	s.logger.InfoContext(ctx, "deleting product", "id", id)

	if err := s.productRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

// GetProductImageURL generates a presigned GET URL for the product image
func (s *productService) GetProductImageURL(ctx context.Context, product *entities.Product) (string, error) {
	if product.Image == "" {
		return "", nil
	}

	url, err := s.storage.GeneratePresignedURL(ctx, product.Image, time.Hour, false)
	if err != nil {
		return "", fmt.Errorf("failed to generate image URL: %w", err)
	}

	if s.baseURL != "" && !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = s.baseURL + url
	}

	return url, nil
}

// GetProductUploadURL generates a presigned PUT URL for uploading a product image.
// The image key is returned but NOT persisted — the caller must update the product
// with the key only after confirming a successful upload.
func (s *productService) GetProductUploadURL(ctx context.Context, product *entities.Product, ext string) (uploadURL string, imageKey string, err error) {
	if product == nil {
		return "", "", fmt.Errorf("product is required")
	}

	// Generate image key
	key := storage.GenerateImageKey(product.ID, ext)

	// Generate presigned PUT URL with 15 minutes expiry
	url, err := s.storage.GeneratePresignedURL(ctx, key, 15*time.Minute, true)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url, key, nil
}

// UploadProductImage uploads a product image directly to S3-compatible storage
func (s *productService) UploadProductImage(ctx context.Context, productID uint, fileData []byte, contentType string) (*entities.Product, error) {
	s.logger.InfoContext(ctx, "uploading product image", "product_id", productID)

	// Get existing product
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Determine file extension from content type
	ext := "jpg" // default
	switch contentType {
	case "image/jpeg":
		ext = "jpg"
	case "image/png":
		ext = "png"
	case "image/gif":
		ext = "gif"
	case "image/webp":
		ext = "webp"
	}

	// Generate image key
	key := storage.GenerateImageKey(product.ID, ext)

	// Upload file to S3-compatible storage
	if err := s.storage.UploadBytes(ctx, key, fileData, contentType); err != nil {
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}

	// Update product with new image key
	updates := map[string]interface{}{
		"image": key,
	}
	updatedProduct, err := s.UpdateProduct(ctx, product.ID, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update product with image key: %w", err)
	}

	return updatedProduct, nil
}

// GetProductImageBytes retrieves the image bytes for a product
func (s *productService) GetProductImageBytes(ctx context.Context, productID uint) ([]byte, string, error) {
	s.logger.InfoContext(ctx, "getting product image bytes", "product_id", productID)

	// Get existing product
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get product: %w", err)
	}

	if product.Image == "" {
		return nil, "", fmt.Errorf("product has no image")
	}

	// Download image bytes from MinIO
	imageBytes, err := s.storage.DownloadBytes(ctx, product.Image)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download image: %w", err)
	}

	// Determine content type based on file extension
	contentType := "image/jpeg" // default
	if len(product.Image) > 4 {
		ext := path.Ext(product.Image)
		switch ext {
		case ".png":
			contentType = "image/png"
		case ".gif":
			contentType = "image/gif"
		case ".webp":
			contentType = "image/webp"
		}
	}

	return imageBytes, contentType, nil
}
