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

type DiscountCampaignHandler struct {
	campaignService interfaces.DiscountCampaignService
	logger          *slog.Logger
}

func NewDiscountCampaignHandler(campaignService interfaces.DiscountCampaignService, logger *slog.Logger) *DiscountCampaignHandler {
	return &DiscountCampaignHandler{
		campaignService: campaignService,
		logger:          logger,
	}
}

type CreateCampaignRequest struct {
	Name               string   `json:"name" validate:"required"`
	CampaignType       string   `json:"campaign_type"`
	DiscountPercentage float64  `json:"discount_percentage"`
	BuyQuantity        *int     `json:"buy_quantity"`
	DiscountAmount     *float64 `json:"discount_amount"`
	RewardProductID    *string  `json:"reward_product_id"`
	RewardQuantity     *int     `json:"reward_quantity"`
	StartDate          string   `json:"start_date" validate:"required"`
	EndDate            string   `json:"end_date" validate:"required"`
	ProductIDs         []string `json:"product_ids"`
}

type UpdateCampaignRequest struct {
	Name               *string  `json:"name"`
	CampaignType       *string  `json:"campaign_type"`
	DiscountPercentage *float64 `json:"discount_percentage"`
	BuyQuantity        *int     `json:"buy_quantity"`
	DiscountAmount     *float64 `json:"discount_amount"`
	RewardProductID    *string  `json:"reward_product_id"`
	RewardQuantity     *int     `json:"reward_quantity"`
	StartDate          *string  `json:"start_date"`
	EndDate            *string  `json:"end_date"`
}

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
	data := map[string]interface{}{
		"name":                campaign.Name,
		"campaign_type":       campaign.CampaignType,
		"discount_percentage": campaign.DiscountPercentage,
		"start_date":          campaign.StartDate.Format(time.RFC3339),
		"end_date":            campaign.EndDate.Format(time.RFC3339),
		"products":            formatCampaignProducts(campaign.Products),
	}
	if campaign.BuyQuantity != nil {
		data["buy_quantity"] = *campaign.BuyQuantity
	}
	if campaign.DiscountAmount != nil {
		data["discount_amount"] = *campaign.DiscountAmount
	}
	if campaign.RewardProductID != nil {
		data["reward_product_id"] = hash.HashID(*campaign.RewardProductID)
	}
	if campaign.RewardQuantity != nil {
		data["reward_quantity"] = *campaign.RewardQuantity
	}
	if campaign.RewardProduct != nil {
		data["reward_product"] = map[string]interface{}{
			"id":   hash.HashID(campaign.RewardProduct.ID),
			"name": campaign.RewardProduct.Name,
			"sku":  campaign.RewardProduct.SKU,
		}
	}
	return WithHashID(
		campaign.ID,
		campaign.CreatedAt.Format(time.RFC3339),
		campaign.UpdatedAt.Format(time.RFC3339),
		data,
	)
}

// @Summary Create a discount campaign
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

	var rewardProductID *uint
	if req.RewardProductID != nil {
		decoded, err := hash.DecodeHashID(*req.RewardProductID)
		if err != nil {
			return ErrorResponse(c, http.StatusBadRequest, "Invalid reward_product_id format")
		}
		rewardProductID = &decoded
	}

	campaign, err := h.campaignService.Create(ctx, interfaces.CreateCampaignRequest{
		Name:               req.Name,
		CampaignType:       req.CampaignType,
		DiscountPercentage: req.DiscountPercentage,
		BuyQuantity:        req.BuyQuantity,
		DiscountAmount:     req.DiscountAmount,
		RewardProductID:    rewardProductID,
		RewardQuantity:     req.RewardQuantity,
		StartDate:          startDate,
		EndDate:            endDate,
		ProductIDs:         productIDs,
	})
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to create campaign", "error", err)
		return ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return SuccessResponse(c, http.StatusCreated, "Campaign created successfully", formatCampaignResponse(campaign))
}

// @Summary List discount campaigns
// @Tags Discount Campaigns
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

// @Summary Get a discount campaign
// @Tags Discount Campaigns
// @Produce json
// @Security bearerAuth
// @Param id path string true "Hashed campaign ID"
// @Success 200 {object} Response{data=HashIDResponse}
// @Failure 400,404,500 {object} Response
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

// @Summary Update a discount campaign
// @Tags Discount Campaigns
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Hashed campaign ID"
// @Param request body UpdateCampaignRequest true "Update campaign request"
// @Success 200 {object} Response{data=HashIDResponse}
// @Failure 400,500 {object} Response
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
		CampaignType:       req.CampaignType,
		DiscountPercentage: req.DiscountPercentage,
		BuyQuantity:        req.BuyQuantity,
		DiscountAmount:     req.DiscountAmount,
		RewardQuantity:     req.RewardQuantity,
	}

	if req.RewardProductID != nil {
		decoded, err := hash.DecodeHashID(*req.RewardProductID)
		if err != nil {
			return ErrorResponse(c, http.StatusBadRequest, "Invalid reward_product_id format")
		}
		svcReq.RewardProductID = &decoded
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
		return ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return SuccessResponse(c, http.StatusOK, "Campaign updated successfully", formatCampaignResponse(campaign))
}

// @Summary Delete a discount campaign
// @Tags Discount Campaigns
// @Produce json
// @Security bearerAuth
// @Param id path string true "Hashed campaign ID"
// @Success 200 {object} Response
// @Failure 400,500 {object} Response
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

// @Summary Add products to a campaign
// @Tags Discount Campaigns
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "Hashed campaign ID"
// @Param request body AddCampaignProductsRequest true "Add products request"
// @Success 200 {object} Response
// @Failure 400,500 {object} Response
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

// @Summary Remove a product from a campaign
// @Tags Discount Campaigns
// @Produce json
// @Security bearerAuth
// @Param id path string true "Hashed campaign ID"
// @Param product_id path string true "Hashed product ID"
// @Success 200 {object} Response
// @Failure 400,500 {object} Response
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
