package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"github.com/usernamesalah/rh-pos/internal/pkg/hash"
	"gorm.io/gorm"
)

// DiscountCampaignHandler handles discount campaign HTTP requests
type DiscountCampaignHandler struct {
	campaignService interfaces.DiscountCampaignService
	logger          *slog.Logger
}

// NewDiscountCampaignHandler creates a new discount campaign handler
func NewDiscountCampaignHandler(campaignService interfaces.DiscountCampaignService, logger *slog.Logger) *DiscountCampaignHandler {
	return &DiscountCampaignHandler{
		campaignService: campaignService,
		logger:          logger,
	}
}

// CreateCampaignRequest is the request body for creating a discount campaign
type CreateCampaignRequest struct {
	Name               string   `json:"name" validate:"required"`
	DiscountPercentage float64  `json:"discount_percentage" validate:"required,gt=0,lte=100"`
	StartDate          string   `json:"start_date" validate:"required"`
	EndDate            string   `json:"end_date" validate:"required"`
	ProductIDs         []string `json:"product_ids"`
}

// UpdateCampaignRequest is the request body for updating a discount campaign
type UpdateCampaignRequest struct {
	Name               *string  `json:"name"`
	DiscountPercentage *float64 `json:"discount_percentage"`
	StartDate          *string  `json:"start_date"`
	EndDate            *string  `json:"end_date"`
}

// AddCampaignProductsRequest is the request body for adding products to a campaign
type AddCampaignProductsRequest struct {
	ProductIDs []string `json:"product_ids" validate:"required,min=1"`
}

func parseDate(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC(), nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

func formatCampaignProducts(products []entities.DiscountCampaignProduct) []map[string]interface{} {
	result := make([]map[string]interface{}, len(products))
	for i, p := range products {
		result[i] = map[string]interface{}{
			"id":         hash.HashID(p.ProductID),
			"product_id": hash.HashID(p.ProductID),
			"name":       p.Product.Name,
			"sku":        p.Product.SKU,
		}
	}
	return result
}

func formatCampaignResponse(campaign *entities.DiscountCampaign) map[string]interface{} {
	return WithHashID(
		campaign.ID,
		campaign.CreatedAt.Format(time.RFC3339),
		campaign.UpdatedAt.Format(time.RFC3339),
		map[string]interface{}{
			"name":                campaign.Name,
			"discount_percentage": campaign.DiscountPercentage,
			"start_date":          campaign.StartDate.Format(time.RFC3339),
			"end_date":            campaign.EndDate.Format(time.RFC3339),
			"products":            formatCampaignProducts(campaign.Products),
		},
	)
}

// CreateCampaign handles POST /api/discount-campaigns
// @Summary Create a discount campaign
// @Description Create a new discount campaign with optional product associations
// @Tags Discount Campaigns
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param request body CreateCampaignRequest true "Create campaign request"
// @Success 201 {object} Response{data=HashIDResponse}
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/discount-campaigns [post]
func (h *DiscountCampaignHandler) CreateCampaign(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateCampaignRequest
	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}
	if err := c.Validate(req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Validation failed")
	}

	startDate, err := parseDate(req.StartDate)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid start_date format")
	}
	endDate, err := parseDate(req.EndDate)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid end_date format")
	}

	productIDs, err := decodeHashIDs(req.ProductIDs)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid product ID format")
	}

	campaign, err := h.campaignService.Create(ctx, interfaces.CreateCampaignRequest{
		Name:               req.Name,
		DiscountPercentage: req.DiscountPercentage,
		StartDate:          startDate,
		EndDate:            endDate,
		ProductIDs:         productIDs,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to create campaign", "error", err)
		return ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return SuccessResponse(c, http.StatusCreated, "Campaign created successfully", formatCampaignResponse(campaign))
}

// ListCampaigns handles GET /api/discount-campaigns
// @Summary List discount campaigns
// @Description List all discount campaigns with pagination
// @Tags Discount Campaigns
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} Response{data=[]HashIDResponse}
// @Failure 500 {object} Response
// @Router /api/discount-campaigns [get]
func (h *DiscountCampaignHandler) ListCampaigns(c echo.Context) error {
	ctx := c.Request().Context()

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	campaigns, total, err := h.campaignService.List(ctx, page, limit)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to list campaigns")
	}

	result := make([]map[string]interface{}, len(campaigns))
	for i := range campaigns {
		result[i] = formatCampaignResponse(&campaigns[i])
	}

	return SuccessPaginatedResponse(c, http.StatusOK, "Campaigns retrieved successfully", result, total, page, limit)
}

// GetCampaign handles GET /api/discount-campaigns/:id
// @Summary Get a discount campaign
// @Description Get a discount campaign by its hashed ID
// @Tags Discount Campaigns
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Hashed campaign ID"
// @Success 200 {object} Response{data=HashIDResponse}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /api/discount-campaigns/{id} [get]
func (h *DiscountCampaignHandler) GetCampaign(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := hash.DecodeHashID(c.Param("id"))
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid campaign ID format")
	}

	campaign, err := h.campaignService.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "Campaign not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to get campaign")
	}

	return SuccessResponse(c, http.StatusOK, "Campaign retrieved successfully", formatCampaignResponse(campaign))
}

// UpdateCampaign handles PUT /api/discount-campaigns/:id
// @Summary Update a discount campaign
// @Description Update an existing discount campaign by its hashed ID
// @Tags Discount Campaigns
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Hashed campaign ID"
// @Param request body UpdateCampaignRequest true "Update campaign request"
// @Success 200 {object} Response{data=HashIDResponse}
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/discount-campaigns/{id} [put]
func (h *DiscountCampaignHandler) UpdateCampaign(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := hash.DecodeHashID(c.Param("id"))
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid campaign ID format")
	}

	var req UpdateCampaignRequest
	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	svcReq := interfaces.UpdateCampaignRequest{
		Name:               req.Name,
		DiscountPercentage: req.DiscountPercentage,
	}
	if req.StartDate != nil {
		t, err := parseDate(*req.StartDate)
		if err != nil {
			return ErrorResponse(c, http.StatusBadRequest, "Invalid start_date format")
		}
		svcReq.StartDate = &t
	}
	if req.EndDate != nil {
		t, err := parseDate(*req.EndDate)
		if err != nil {
			return ErrorResponse(c, http.StatusBadRequest, "Invalid end_date format")
		}
		svcReq.EndDate = &t
	}

	campaign, err := h.campaignService.Update(ctx, id, svcReq)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return SuccessResponse(c, http.StatusOK, "Campaign updated successfully", formatCampaignResponse(campaign))
}

// DeleteCampaign handles DELETE /api/discount-campaigns/:id
// @Summary Delete a discount campaign
// @Description Delete a discount campaign by its hashed ID
// @Tags Discount Campaigns
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Hashed campaign ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/discount-campaigns/{id} [delete]
func (h *DiscountCampaignHandler) DeleteCampaign(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := hash.DecodeHashID(c.Param("id"))
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid campaign ID format")
	}

	if err := h.campaignService.Delete(ctx, id); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to delete campaign")
	}

	return SuccessResponse(c, http.StatusOK, "Campaign deleted successfully", nil)
}

// AddProducts handles POST /api/discount-campaigns/:id/products
// @Summary Add products to a campaign
// @Description Add one or more products to a discount campaign
// @Tags Discount Campaigns
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Hashed campaign ID"
// @Param request body AddCampaignProductsRequest true "Add products request"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/discount-campaigns/{id}/products [post]
func (h *DiscountCampaignHandler) AddProducts(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := hash.DecodeHashID(c.Param("id"))
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid campaign ID format")
	}

	var req AddCampaignProductsRequest
	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}
	if err := c.Validate(req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Validation failed")
	}

	productIDs, err := decodeHashIDs(req.ProductIDs)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid product ID format")
	}

	if err := h.campaignService.AddProducts(ctx, id, productIDs); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return SuccessResponse(c, http.StatusOK, "Products added to campaign", nil)
}

// RemoveProduct handles DELETE /api/discount-campaigns/:id/products/:product_id
// @Summary Remove a product from a campaign
// @Description Remove a specific product from a discount campaign
// @Tags Discount Campaigns
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Hashed campaign ID"
// @Param product_id path string true "Hashed product ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/discount-campaigns/{id}/products/{product_id} [delete]
func (h *DiscountCampaignHandler) RemoveProduct(c echo.Context) error {
	ctx := c.Request().Context()

	campaignID, err := hash.DecodeHashID(c.Param("id"))
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid campaign ID format")
	}

	productID, err := hash.DecodeHashID(c.Param("product_id"))
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid product ID format")
	}

	if err := h.campaignService.RemoveProduct(ctx, campaignID, productID); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return SuccessResponse(c, http.StatusOK, "Product removed from campaign", nil)
}

// decodeHashIDs decodes a slice of hashed IDs to uint
func decodeHashIDs(hashedIDs []string) ([]uint, error) {
	ids := make([]uint, 0, len(hashedIDs))
	for _, hid := range hashedIDs {
		id, err := hash.DecodeHashID(hid)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
