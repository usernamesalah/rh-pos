package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"github.com/usernamesalah/rh-pos/internal/pkg/hash"
	"gorm.io/gorm"
)

type TransactionHandler struct {
	transactionService interfaces.TransactionService
	logger             *slog.Logger
}

// NewTransactionHandler creates a new transaction handler
func NewTransactionHandler(transactionService interfaces.TransactionService, logger *slog.Logger) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
		logger:             logger,
	}
}

// CreateTransactionRequest represents the create transaction request
type CreateTransactionRequest struct {
	Items         []TransactionItemRequest `json:"items" validate:"required,min=1"`
	User          string                   `json:"user" validate:"required"`
	PaymentMethod string                   `json:"payment_method" validate:"required"`
	Discount      float64                  `json:"discount"`
	TotalPrice    float64                  `json:"total_price" validate:"required,min=0"`
}

// TransactionItemRequest represents an item in transaction request
type TransactionItemRequest struct {
	ProductID uint `json:"product_id" validate:"required"`
	Quantity  int  `json:"quantity" validate:"required,min=1"`
}

// CreateTransaction handles creating a new transaction
// @Summary Create a new transaction
// @Description Create a new sales transaction
// @Tags Transactions
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param request body CreateTransactionRequest true "Create transaction request"
// @Success 201 {object} Response{data=HashIDResponse}
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /transactions [post]
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

	// Convert to service request
	serviceReq := interfaces.CreateTransactionRequest{
		User:          req.User,
		PaymentMethod: req.PaymentMethod,
		Discount:      req.Discount,
		TotalPrice:    req.TotalPrice,
		Items:         make([]interfaces.TransactionItemRequest, len(req.Items)),
	}

	for i, item := range req.Items {
		serviceReq.Items[i] = interfaces.TransactionItemRequest{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	transaction, err := h.transactionService.CreateTransaction(ctx, serviceReq)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to create transaction", "error", err)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to create transaction")
	}

	// Convert items to flattened structure
	items := make([]map[string]interface{}, len(transaction.Items))
	for i, item := range transaction.Items {
		items[i] = map[string]interface{}{
			"product_id": item.ProductID,
			"quantity":   item.Quantity,
			"price":      item.Price,
			"product": map[string]interface{}{
				"name":        item.Product.Name,
				"sku":         item.Product.SKU,
				"image":       item.Product.Image,
				"harga_modal": item.Product.HargaModal,
				"harga_jual":  item.Product.HargaJual,
				"stock":       item.Product.Stock,
			},
		}
	}

	response := WithHashID(
		transaction.ID,
		transaction.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		transaction.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		map[string]interface{}{
			"items":          items,
			"user":           transaction.User,
			"payment_method": transaction.PaymentMethod,
			"discount":       transaction.Discount,
			"total_price":    transaction.TotalPrice,
		},
	)

	return SuccessResponse(c, http.StatusCreated, "Transaction created successfully", response)
}

// ListTransactions handles listing transactions with pagination
// @Summary List all transactions
// @Description Get a paginated list of transactions
// @Tags Transactions
// @Produce json
// @Security bearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} Response{data=[]HashIDResponse}
// @Failure 401 {object} Response
// @Router /transactions [get]
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

	// Convert transactions to HashIDResponse
	items := make([]HashIDResponse, len(transactions))
	for i, t := range transactions {
		// Convert items to flattened structure
		transactionItems := make([]map[string]interface{}, len(t.Items))
		for j, item := range t.Items {
			transactionItems[j] = map[string]interface{}{
				"product_id": item.ProductID,
				"quantity":   item.Quantity,
				"price":      item.Price,
				"product": map[string]interface{}{
					"name":        item.Product.Name,
					"sku":         item.Product.SKU,
					"image":       item.Product.Image,
					"harga_modal": item.Product.HargaModal,
					"harga_jual":  item.Product.HargaJual,
					"stock":       item.Product.Stock,
				},
			}
		}

		items[i] = WithHashID(
			t.ID,
			t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			map[string]interface{}{
				"items":          transactionItems,
				"user":           t.User,
				"payment_method": t.PaymentMethod,
				"discount":       t.Discount,
				"total_price":    t.TotalPrice,
			},
		)
	}

	return SuccessPaginatedResponse(c, http.StatusOK, "Transactions retrieved successfully", items, total, page, limit)
}

// GetTransaction handles getting a single transaction by ID
// @Summary Get a transaction by ID
// @Description Get detailed information about a specific transaction
// @Tags Transactions
// @Produce json
// @Security bearerAuth
// @Param id path string true "Transaction ID"
// @Success 200 {object} Response{data=HashIDResponse}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /transactions/{id} [get]
func (h *TransactionHandler) GetTransaction(c echo.Context) error {
	ctx := c.Request().Context()

	// Get hashed ID from URL
	hashedID := c.Param("id")

	// Decode hashed ID to get the actual ID
	id, err := hash.DecodeHashID(hashedID)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid transaction ID format", "error", err, "hashed_id", hashedID)
		return ErrorResponse(c, http.StatusBadRequest, "Invalid transaction ID format")
	}

	transaction, err := h.transactionService.GetTransaction(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrorResponse(c, http.StatusNotFound, "Transaction not found")
		}
		h.logger.ErrorContext(ctx, "failed to get transaction", "error", err, "id", id)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to get transaction")
	}

	// Convert items to flattened structure
	items := make([]map[string]interface{}, len(transaction.Items))
	for i, item := range transaction.Items {
		items[i] = map[string]interface{}{
			"product_id": item.ProductID,
			"quantity":   item.Quantity,
			"price":      item.Price,
			"product": map[string]interface{}{
				"name":        item.Product.Name,
				"sku":         item.Product.SKU,
				"image":       item.Product.Image,
				"harga_modal": item.Product.HargaModal,
				"harga_jual":  item.Product.HargaJual,
				"stock":       item.Product.Stock,
			},
		}
	}

	response := WithHashID(
		transaction.ID,
		transaction.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		transaction.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		map[string]interface{}{
			"items":          items,
			"user":           transaction.User,
			"payment_method": transaction.PaymentMethod,
			"discount":       transaction.Discount,
			"total_price":    transaction.TotalPrice,
		},
	)

	return SuccessResponse(c, http.StatusOK, "Transaction retrieved successfully", response)
}
