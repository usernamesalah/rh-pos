package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"github.com/usernamesalah/rh-pos/internal/pkg/hash"
)

// CategoryHandler handles HTTP requests for category operations
type CategoryHandler struct {
	categoryService interfaces.CategoryService
	logger          *slog.Logger
}

// NewCategoryHandler creates a new CategoryHandler
func NewCategoryHandler(categoryService interfaces.CategoryService, logger *slog.Logger) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService, logger: logger}
}

// CreateCategoryRequest represents the create category request
type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required"`
}

// UpdateCategoryRequest represents the update category request
type UpdateCategoryRequest struct {
	Name string `json:"name" validate:"required"`
}

// ListCategories handles GET /api/categories
func (h *CategoryHandler) ListCategories(c echo.Context) error {
	ctx := c.Request().Context()

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	categories, total, err := h.categoryService.ListCategories(ctx, page, limit)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to list categories", "error", err)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to list categories")
	}

	items := make([]HashIDResponse, len(categories))
	for i, cat := range categories {
		items[i] = WithHashID(
			cat.ID,
			cat.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			cat.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			map[string]interface{}{
				"name": cat.Name,
			},
		)
	}

	return SuccessPaginatedResponse(c, http.StatusOK, "Categories retrieved successfully", items, total, page, limit)
}

// GetCategory handles GET /api/categories/:id
func (h *CategoryHandler) GetCategory(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := hash.DecodeHashID(c.Param("id"))
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid category ID format")
	}

	cat, err := h.categoryService.GetCategory(ctx, id)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get category", "error", err, "id", id)
		return ErrorResponse(c, http.StatusNotFound, "Category not found")
	}

	return SuccessResponse(c, http.StatusOK, "Category retrieved successfully", WithHashID(
		cat.ID,
		cat.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		cat.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		map[string]interface{}{
			"name": cat.Name,
		},
	))
}

// CreateCategory handles POST /api/categories
func (h *CategoryHandler) CreateCategory(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	cat := &entities.Category{Name: req.Name}
	if err := h.categoryService.CreateCategory(ctx, cat); err != nil {
		h.logger.ErrorContext(ctx, "failed to create category", "error", err)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to create category")
	}

	return SuccessResponse(c, http.StatusCreated, "Category created successfully", WithHashID(
		cat.ID,
		cat.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		cat.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		map[string]interface{}{
			"name": cat.Name,
		},
	))
}

// UpdateCategory handles PUT /api/categories/:id
func (h *CategoryHandler) UpdateCategory(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := hash.DecodeHashID(c.Param("id"))
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid category ID format")
	}

	var req UpdateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	cat, err := h.categoryService.UpdateCategory(ctx, id, req.Name)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to update category", "error", err, "id", id)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to update category")
	}

	return SuccessResponse(c, http.StatusOK, "Category updated successfully", WithHashID(
		cat.ID,
		cat.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		cat.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		map[string]interface{}{
			"name": cat.Name,
		},
	))
}

// DeleteCategory handles DELETE /api/categories/:id
func (h *CategoryHandler) DeleteCategory(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := hash.DecodeHashID(c.Param("id"))
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid category ID format")
	}

	if err := h.categoryService.DeleteCategory(ctx, id); err != nil {
		h.logger.ErrorContext(ctx, "failed to delete category", "error", err, "id", id)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to delete category")
	}

	return SuccessResponse(c, http.StatusOK, "Category deleted successfully", nil)
}
