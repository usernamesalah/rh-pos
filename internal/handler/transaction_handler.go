package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"github.com/usernamesalah/rh-pos/internal/pkg/hash"
	"gorm.io/gorm"
)

type TransactionHandler struct {
	transactionService interfaces.TransactionService
	logger             *slog.Logger
}

func NewTransactionHandler(transactionService interfaces.TransactionService, logger *slog.Logger) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
		logger:             logger,
	}
}

type CreateTransactionRequest struct {
	Items         []TransactionItemRequest `json:"items" validate:"required,min=1"`
	PaymentMethod string                   `json:"payment_method" validate:"required"`
	Discount      float64                  `json:"discount"`
	TotalPrice    float64                  `json:"total_price"`
	Notes         string                   `json:"notes"`
	CustomerName  *string                  `json:"customer_name"`
	CustomerEmail *string                  `json:"customer_email"`
	CustomerPhone *string                  `json:"customer_phone"`
}

type TransactionItemRequest struct {
	ProductID    string  `json:"product_id" validate:"required"`
	Quantity     int     `json:"quantity" validate:"required,min=1"`
	WarrantyDays int     `json:"warranty_days"`
	IsFreeItem   bool    `json:"is_free_item"`
	CampaignID   *string `json:"campaign_id"`
}

// @Summary Create a new transaction
// @Tags Transactions
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param request body CreateTransactionRequest true "Create transaction request"
// @Success 201 {object} Response{data=HashIDResponse}
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/transactions [post]
func (h *TransactionHandler) CreateTransaction(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateTransactionRequest
	if err := c.Bind(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid request body", "error", err)
		return ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(req); err != nil {
		h.logger.WarnContext(ctx, "validation failed", "error", err)
		return ErrorResponse(c, http.StatusBadRequest, "Validation failed")
	}

	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return ErrorResponse(c, http.StatusUnauthorized, "Invalid token claims")
	}

	serviceReq := interfaces.CreateTransactionRequest{
		UserID:        userID,
		PaymentMethod: req.PaymentMethod,
		Discount:      req.Discount,
		TotalPrice:    req.TotalPrice,
		Notes:         req.Notes,
		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		CustomerPhone: req.CustomerPhone,
		Items:         make([]interfaces.TransactionItemRequest, len(req.Items)),
	}

	for i, item := range req.Items {
		productID, err := hash.DecodeHashID(item.ProductID)
		if err != nil {
			h.logger.WarnContext(ctx, "invalid product ID format", "error", err, "hashed_id", item.ProductID)
			return ErrorResponse(c, http.StatusBadRequest, "Invalid product ID format")
		}

		svcItem := interfaces.TransactionItemRequest{
			ProductID:    productID,
			Quantity:     item.Quantity,
			WarrantyDays: item.WarrantyDays,
			IsFreeItem:   item.IsFreeItem,
		}

		if item.CampaignID != nil {
			campaignID, err := hash.DecodeHashID(*item.CampaignID)
			if err != nil {
				h.logger.WarnContext(ctx, "invalid campaign ID format", "error", err)
				return ErrorResponse(c, http.StatusBadRequest, "Invalid campaign ID format")
			}
			svcItem.CampaignID = &campaignID
		}

		serviceReq.Items[i] = svcItem
	}

	transaction, err := h.transactionService.CreateTransaction(ctx, serviceReq)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to create transaction", "error", err)
		return ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return SuccessResponse(c, http.StatusCreated, "Transaction created successfully", formatTransaction(transaction))
}

// @Summary List all transactions
// @Tags Transactions
// @Produce json
// @Security bearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} Response{data=[]HashIDResponse}
// @Failure 401 {object} Response
// @Router /api/transactions [get]
func (h *TransactionHandler) ListTransactions(c echo.Context) error {
	ctx := c.Request().Context()

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	transactions, total, err := h.transactionService.ListTransactions(ctx, page, limit)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to list transactions", "error", err)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to list transactions")
	}

	result := make([]HashIDResponse, len(transactions))
	for i := range transactions {
		result[i] = formatTransaction(&transactions[i])
	}

	return SuccessPaginatedResponse(c, http.StatusOK, "Transactions retrieved successfully", result, total, page, limit)
}

// @Summary Get a transaction by ID
// @Tags Transactions
// @Produce json
// @Security bearerAuth
// @Param id path string true "Transaction ID"
// @Success 200 {object} Response{data=HashIDResponse}
// @Failure 400,404 {object} Response
// @Router /api/transactions/{id} [get]
func (h *TransactionHandler) GetTransaction(c echo.Context) error {
	ctx := c.Request().Context()

	hashedID := c.Param("id")
	id, err := hash.DecodeHashID(hashedID)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid transaction ID format", "error", err, "hashed_id", hashedID)
		return ErrorResponse(c, http.StatusBadRequest, "Invalid transaction ID format")
	}

	transaction, err := h.transactionService.GetTransaction(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "Transaction not found")
		}
		h.logger.ErrorContext(ctx, "failed to get transaction", "error", err, "id", id)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to get transaction")
	}

	return SuccessResponse(c, http.StatusOK, "Transaction retrieved successfully", formatTransaction(transaction))
}

func formatTransactionItems(items []entities.TransactionItem) []map[string]interface{} {
	result := make([]map[string]interface{}, len(items))
	for i, item := range items {
		itemMap := map[string]interface{}{
			"product_id":          hash.HashID(item.ProductID),
			"quantity":            item.Quantity,
			"price":               item.Price,
			"warranty_days":       item.WarrantyDays,
			"discount_percentage": item.DiscountPercentage,
			"discount_amount":     item.DiscountAmount,
			"is_free_item":        item.IsFreeItem,
			"product": map[string]interface{}{
				"id":          hash.HashID(item.Product.ID),
				"name":        item.Product.Name,
				"sku":         item.Product.SKU,
				"image":       item.Product.Image,
				"harga_modal": item.Product.HargaModal,
				"harga_jual":  item.Product.HargaJual,
				"stock":       item.Product.Stock,
			},
		}
		if item.CampaignID != nil {
			itemMap["campaign_id"] = hash.HashID(*item.CampaignID)
		}
		if item.CampaignType != "" {
			itemMap["campaign_type"] = item.CampaignType
		}
		if item.CampaignGroupKey != "" {
			itemMap["campaign_group_key"] = item.CampaignGroupKey
		}
		result[i] = itemMap
	}
	return result
}

func formatTransaction(t *entities.Transaction) HashIDResponse {
	data := map[string]interface{}{
		"items":          formatTransactionItems(t.Items),
		"payment_method": t.PaymentMethod,
		"discount":       t.Discount,
		"total_price":    t.TotalPrice,
		"notes":          t.Notes,
	}

	if t.UserID != nil {
		data["user_id"] = hash.HashID(*t.UserID)
	}
	if t.User != nil {
		data["user"] = map[string]interface{}{
			"id":       hash.HashID(t.User.ID),
			"username": t.User.Username,
		}
	}

	if t.CustomerName != nil {
		data["customer_name"] = *t.CustomerName
	}
	if t.CustomerEmail != nil {
		data["customer_email"] = *t.CustomerEmail
	}
	if t.CustomerPhone != nil {
		data["customer_phone"] = *t.CustomerPhone
	}
	return WithHashID(
		t.ID,
		t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		data,
	)
}
