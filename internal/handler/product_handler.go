package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
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

// ProductListResponse represents the paginated product list response
type ProductListResponse struct {
	Items []entities.Product `json:"items"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Limit int                `json:"limit"`
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
// @Success 200 {object} ProductListResponse
// @Failure 401 {object} ErrorResponse
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
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to list products"})
	}

	response := ProductListResponse{
		Items: products,
		Total: total,
		Page:  page,
		Limit: limit,
	}

	return c.JSON(http.StatusOK, response)
}

// GetProduct handles getting a single product by ID
// @Summary Get a product by ID
// @Description Get detailed information about a specific product
// @Tags Products
// @Produce json
// @Security bearerAuth
// @Param id path int true "Product ID"
// @Success 200 {object} entities.Product
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /products/{id} [get]
func (h *ProductHandler) GetProduct(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid product ID"})
	}

	product, err := h.productService.GetProduct(ctx, uint(id))
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get product", "error", err, "id", id)
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "Product not found"})
	}

	return c.JSON(http.StatusOK, product)
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
// @Success 200 {object} entities.Product
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /products/{id} [put]
func (h *ProductHandler) UpdateProduct(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid product ID"})
	}

	var req UpdateProductRequest
	if err := c.Bind(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid request body", "error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
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

	product, err := h.productService.UpdateProduct(ctx, uint(id), updates)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to update product", "error", err, "id", id)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update product"})
	}

	return c.JSON(http.StatusOK, product)
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
// @Success 200 {object} entities.Product
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /products/{id}/stock [put]
func (h *ProductHandler) UpdateStock(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid product ID"})
	}

	var req UpdateStockRequest
	if err := c.Bind(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid request body", "error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	if err := c.Validate(req); err != nil {
		h.logger.WarnContext(ctx, "validation failed", "error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Validation failed"})
	}

	product, err := h.productService.UpdateStock(ctx, uint(id), req.Stock)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to update stock", "error", err, "id", id)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update stock"})
	}

	return c.JSON(http.StatusOK, product)
}
