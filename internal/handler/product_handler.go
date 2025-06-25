package handler

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"github.com/usernamesalah/rh-pos/internal/pkg/hash"
)

type ProductHandler struct {
	productService interfaces.ProductService
	logger         *slog.Logger
}

// NewProductHandler creates a new product handler
func NewProductHandler(productService interfaces.ProductService, logger *slog.Logger) *ProductHandler {
	return &ProductHandler{
		productService: productService,
		logger:         logger,
	}
}

// UpdateProductRequest represents the update product request
type UpdateProductRequest struct {
	Name       *string  `json:"name,omitempty"`
	SKU        *string  `json:"sku,omitempty"`
	HargaModal *float64 `json:"harga_modal,omitempty"`
	HargaJual  *float64 `json:"harga_jual,omitempty"`
}

// UpdateStockRequest represents the update stock request
type UpdateStockRequest struct {
	Stock int `json:"stock" validate:"required,min=0"`
}

// CreateProductRequest represents the create product request
type CreateProductRequest struct {
	Name       string  `json:"name" validate:"required"`
	SKU        string  `json:"sku" validate:"required"`
	Image      string  `json:"image,omitempty"`
	HargaModal float64 `json:"harga_modal" validate:"required,min=0"`
	HargaJual  float64 `json:"harga_jual" validate:"required,min=0"`
	Stock      int     `json:"stock" validate:"required,min=0"`
}

// GetUploadURLRequest represents the request for getting an upload URL
type GetUploadURLRequest struct {
	Extension string `json:"extension" validate:"required"`
}

// ListProducts handles listing products with pagination
// @Summary List all products
// @Description Get a paginated list of products
// @Tags Products
// @Produce json
// @Security bearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} Response{data=PaginatedResponse[HashIDResponse]}
// @Failure 401 {object} Response
// @Router /products [get]
func (h *ProductHandler) ListProducts(c echo.Context) error {
	ctx := c.Request().Context()

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	products, total, err := h.productService.ListProducts(ctx, page, limit)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to list products", "error", err)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to list products")
	}

	// Convert products to HashIDResponse
	items := make([]HashIDResponse, len(products))
	for i, p := range products {
		// Get presigned image URL if image exists
		imageURL := ""
		if p.Image != "" {
			imageURL, err = h.productService.GetProductImageURL(ctx, &p)
			if err != nil {
				h.logger.ErrorContext(ctx, "failed to get image URL", "error", err, "product_id", p.ID)
			}
		}

		items[i] = WithHashID(
			p.ID,
			p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			map[string]interface{}{
				"name":        p.Name,
				"sku":         p.SKU,
				"image_url":   imageURL,
				"harga_modal": p.HargaModal,
				"harga_jual":  p.HargaJual,
				"stock":       p.Stock,
			},
		)
	}

	return SuccessPaginatedResponse(
		c,
		http.StatusOK,
		"Products retrieved successfully",
		items,
		total,
		page,
		limit,
	)
}

// GetProduct handles getting a single product by ID
// @Summary Get a product by ID
// @Description Get detailed information about a specific product
// @Tags Products
// @Produce json
// @Security bearerAuth
// @Param id path string true "Product ID"
// @Success 200 {object} Response{data=HashIDResponse}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /products/{id} [get]
func (h *ProductHandler) GetProduct(c echo.Context) error {
	ctx := c.Request().Context()

	// Get hashed ID from URL
	hashedID := c.Param("id")

	// Decode hashed ID to get the actual ID
	id, err := hash.DecodeHashID(hashedID)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid product ID format", "error", err, "hashed_id", hashedID)
		return ErrorResponse(c, http.StatusBadRequest, "Invalid product ID format")
	}

	product, err := h.productService.GetProduct(ctx, id)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get product", "error", err, "id", id)
		return ErrorResponse(c, http.StatusNotFound, "Product not found")
	}

	// Get presigned image URL if image exists
	imageURL := ""
	if product.Image != "" {
		imageURL, err = h.productService.GetProductImageURL(ctx, product)
		if err != nil {
			h.logger.ErrorContext(ctx, "failed to get image URL", "error", err, "product_id", product.ID)
		}
	}

	response := WithHashID(
		product.ID,
		product.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		product.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		map[string]interface{}{
			"name":        product.Name,
			"sku":         product.SKU,
			"image_url":   imageURL,
			"harga_modal": product.HargaModal,
			"harga_jual":  product.HargaJual,
			"stock":       product.Stock,
		},
	)

	return SuccessResponse(c, http.StatusOK, "Product retrieved successfully", response)
}

// UpdateProduct handles updating a product
// @Summary Update a product
// @Description Update product information
// @Tags Products
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path int true "Product ID"
// @Param request body UpdateProductRequest true "Update product request"
// @Success 200 {object} Response{data=HashIDResponse}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /products/{id} [put]
func (h *ProductHandler) UpdateProduct(c echo.Context) error {
	ctx := c.Request().Context()

	// Get hashed ID from URL
	hashedID := c.Param("id")

	// Decode hashed ID to get the actual ID
	id, err := hash.DecodeHashID(hashedID)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid product ID format", "error", err, "hashed_id", hashedID)
		return ErrorResponse(c, http.StatusBadRequest, "Invalid product ID format")
	}

	var req UpdateProductRequest
	if err := c.Bind(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid request body", "error", err)
		return ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	// Convert to updates map
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.SKU != nil {
		updates["sku"] = *req.SKU
	}
	if req.HargaModal != nil {
		updates["harga_modal"] = *req.HargaModal
	}
	if req.HargaJual != nil {
		updates["harga_jual"] = *req.HargaJual
	}

	product, err := h.productService.UpdateProduct(ctx, id, updates)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to update product", "error", err, "id", id)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to update product")
	}

	// Get presigned image URL if image exists
	imageURL := ""
	if product.Image != "" {
		imageURL, err = h.productService.GetProductImageURL(ctx, product)
		if err != nil {
			h.logger.ErrorContext(ctx, "failed to get image URL", "error", err, "product_id", product.ID)
		}
	}

	response := WithHashID(
		product.ID,
		product.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		product.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		map[string]interface{}{
			"name":        product.Name,
			"sku":         product.SKU,
			"image_url":   imageURL,
			"harga_modal": product.HargaModal,
			"harga_jual":  product.HargaJual,
			"stock":       product.Stock,
		},
	)

	return SuccessResponse(c, http.StatusOK, "Product updated successfully", response)
}

// UpdateStock handles updating product stock
// @Summary Update product stock
// @Description Update the stock quantity of a product
// @Tags Products
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path int true "Product ID"
// @Param request body UpdateStockRequest true "Update stock request"
// @Success 200 {object} Response{data=HashIDResponse}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /products/{id}/stock [put]
func (h *ProductHandler) UpdateStock(c echo.Context) error {
	ctx := c.Request().Context()

	// Get hashed ID from URL
	hashedID := c.Param("id")

	// Decode hashed ID to get the actual ID
	id, err := hash.DecodeHashID(hashedID)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid product ID format", "error", err, "hashed_id", hashedID)
		return ErrorResponse(c, http.StatusBadRequest, "Invalid product ID format")
	}

	var req UpdateStockRequest
	if err := c.Bind(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid request body", "error", err)
		return ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(req); err != nil {
		h.logger.WarnContext(ctx, "validation failed", "error", err)
		return ErrorResponse(c, http.StatusBadRequest, "Validation failed")
	}

	product, err := h.productService.UpdateStock(ctx, id, req.Stock)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to update stock", "error", err, "id", id)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to update stock")
	}

	// Get presigned image URL if image exists
	imageURL := ""
	if product.Image != "" {
		imageURL, err = h.productService.GetProductImageURL(ctx, product)
		if err != nil {
			h.logger.ErrorContext(ctx, "failed to get image URL", "error", err, "product_id", product.ID)
		}
	}

	response := WithHashID(
		product.ID,
		product.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		product.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		map[string]interface{}{
			"name":        product.Name,
			"sku":         product.SKU,
			"image":       product.Image,
			"image_url":   imageURL,
			"harga_modal": product.HargaModal,
			"harga_jual":  product.HargaJual,
			"stock":       product.Stock,
		},
	)

	return SuccessResponse(c, http.StatusOK, "Stock updated successfully", response)
}

// CreateProduct handles creating a new product
// @Summary Create a new product
// @Description Create a new product with the provided details
// @Tags Products
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param request body CreateProductRequest true "Create product request"
// @Success 201 {object} Response{data=HashIDResponse}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Router /products [post]
func (h *ProductHandler) CreateProduct(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateProductRequest
	if err := c.Bind(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid request body", "error", err)
		return ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(req); err != nil {
		h.logger.WarnContext(ctx, "validation failed", "error", err)
		return ErrorResponse(c, http.StatusBadRequest, "Validation failed")
	}

	// Create product entity
	product := &entities.Product{
		Name:       req.Name,
		SKU:        req.SKU,
		Image:      req.Image,
		HargaModal: req.HargaModal,
		HargaJual:  req.HargaJual,
		Stock:      req.Stock,
	}

	// Set tenant_id from context
	if tenantID, ok := c.Get("tenant_id").(uint); ok {
		product.TenantID = &tenantID
	}

	if err := h.productService.CreateProduct(ctx, product); err != nil {
		h.logger.ErrorContext(ctx, "failed to create product", "error", err)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to create product")
	}

	response := WithHashID(
		product.ID,
		product.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		product.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		map[string]interface{}{
			"name":        product.Name,
			"sku":         product.SKU,
			"image":       product.Image,
			"harga_modal": product.HargaModal,
			"harga_jual":  product.HargaJual,
			"stock":       product.Stock,
		},
	)

	return SuccessResponse(c, http.StatusCreated, "Product created successfully", response)
}

// GetUploadURL handles getting a presigned URL for uploading a product image
// @Summary Get upload URL for product image
// @Description Get a presigned URL for uploading a product image
// @Tags Products
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Product ID"
// @Param request body GetUploadURLRequest true "Upload URL request"
// @Success 200 {object} Response{data=map[string]string}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /products/{id}/upload-url [get]
func (h *ProductHandler) GetUploadURL(c echo.Context) error {
	ctx := c.Request().Context()

	// Get hashed ID from URL
	hashedID := c.Param("id")

	// Decode hashed ID to get the actual ID
	id, err := hash.DecodeHashID(hashedID)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid product ID format", "error", err, "hashed_id", hashedID)
		return ErrorResponse(c, http.StatusBadRequest, "Invalid product ID format")
	}

	var req GetUploadURLRequest
	if err := c.Bind(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid request body", "error", err)
		return ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(req); err != nil {
		h.logger.WarnContext(ctx, "validation failed", "error", err)
		return ErrorResponse(c, http.StatusBadRequest, "Validation failed")
	}

	// Get product
	product, err := h.productService.GetProduct(ctx, id)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get product", "error", err, "id", id)
		return ErrorResponse(c, http.StatusNotFound, "Product not found")
	}

	// Get presigned upload URL
	uploadURL, err := h.productService.GetProductUploadURL(ctx, product, req.Extension)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get upload URL", "error", err, "product_id", product.ID)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to get upload URL")
	}

	return SuccessResponse(c, http.StatusOK, "Upload URL generated successfully", map[string]string{
		"upload_url": uploadURL,
	})
}

// UploadProductImage handles uploading an image for a product
// @Summary Upload product image
// @Description Upload an image for a product (replaces existing image if any)
// @Tags Products
// @Accept multipart/form-data
// @Produce json
// @Security bearerAuth
// @Param id path string true "Product ID"
// @Param image formData file true "Product image"
// @Success 200 {object} Response{data=HashIDResponse}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /products/{id}/image [post]
func (h *ProductHandler) UploadProductImage(c echo.Context) error {
	ctx := c.Request().Context()

	// Get hashed ID from URL
	hashedID := c.Param("id")

	// Decode hashed ID to get the actual ID
	id, err := hash.DecodeHashID(hashedID)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid product ID format", "error", err, "hashed_id", hashedID)
		return ErrorResponse(c, http.StatusBadRequest, "Invalid product ID format")
	}

	// Parse multipart form
	if err := c.Request().ParseMultipartForm(32 << 20); err != nil { // 32MB max
		h.logger.WarnContext(ctx, "failed to parse multipart form", "error", err)
		return ErrorResponse(c, http.StatusBadRequest, "Failed to parse form data")
	}

	// Check if image file is provided
	form := c.Request().MultipartForm
	files, ok := form.File["image"]
	if !ok || len(files) == 0 {
		return ErrorResponse(c, http.StatusBadRequest, "Image file is required")
	}

	file := files[0]

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to open uploaded file", "error", err)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to process uploaded file")
	}
	defer src.Close()

	// Read file data
	fileData, err := io.ReadAll(src)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to read uploaded file", "error", err)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to read uploaded file")
	}

	// Upload image to MinIO
	product, err := h.productService.UploadProductImage(ctx, id, fileData, file.Header.Get("Content-Type"))
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to upload product image", "error", err)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to upload product image")
	}

	// Get presigned image URL
	imageURL, err := h.productService.GetProductImageURL(ctx, product)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get image URL", "error", err, "product_id", product.ID)
	}

	response := WithHashID(
		product.ID,
		product.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		product.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		map[string]interface{}{
			"name":        product.Name,
			"sku":         product.SKU,
			"image_url":   imageURL,
			"harga_modal": product.HargaModal,
			"harga_jual":  product.HargaJual,
			"stock":       product.Stock,
		},
	)

	return SuccessResponse(c, http.StatusOK, "Product image uploaded successfully", response)
}

// GetProductImageBytes handles serving product image bytes directly
// @Summary Get product image
// @Description Get product image as bytes (serves the actual image file)
// @Tags Products
// @Produce image/*
// @Security bearerAuth
// @Param id path string true "Product ID"
// @Success 200 {file} binary "Image file"
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /products/{id}/image/bytes [get]
func (h *ProductHandler) GetProductImageBytes(c echo.Context) error {
	ctx := c.Request().Context()

	// Get hashed ID from URL
	hashedID := c.Param("id")

	// Decode hashed ID to get the actual ID
	id, err := hash.DecodeHashID(hashedID)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid product ID format", "error", err, "hashed_id", hashedID)
		return ErrorResponse(c, http.StatusBadRequest, "Invalid product ID format")
	}

	// Get image bytes from service
	imageBytes, contentType, err := h.productService.GetProductImageBytes(ctx, id)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get product image bytes", "error", err, "product_id", id)
		return ErrorResponse(c, http.StatusNotFound, "Product image not found")
	}

	// Set response headers
	c.Response().Header().Set("Content-Type", contentType)
	c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", len(imageBytes)))
	c.Response().Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour

	// Write image bytes to response
	return c.Blob(http.StatusOK, contentType, imageBytes)
}
