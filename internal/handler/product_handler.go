package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
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
	Image      *string  `json:"image,omitempty"`
	Name       *string  `json:"name,omitempty"`
	SKU        *string  `json:"sku,omitempty"`
	HargaModal *float64 `json:"harga_modal,omitempty"`
	HargaJual  *float64 `json:"harga_jual,omitempty"`
	Stock      *int     `json:"stock,omitempty"`
}

// UpdateStockRequest represents the update stock request
type UpdateStockRequest struct {
	Stock int `json:"stock" validate:"required,min=0"`
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
		items[i] = WithHashID(
			p.ID,
			p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			map[string]interface{}{
				"name":        p.Name,
				"sku":         p.SKU,
				"image":       p.Image,
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
	if req.Image != nil {
		updates["image"] = *req.Image
	}
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
	if req.Stock != nil {
		updates["stock"] = *req.Stock
	}

	product, err := h.productService.UpdateProduct(ctx, id, updates)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to update product", "error", err, "id", id)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to update product")
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

	return SuccessResponse(c, http.StatusOK, "Stock updated successfully", response)
}
