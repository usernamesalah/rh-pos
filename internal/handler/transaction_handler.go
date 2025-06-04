package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
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

// TransactionListResponse represents the paginated transaction list response
type TransactionListResponse struct {
	Items []entities.Transaction `json:"items"`
	Total int64                  `json:"total"`
	Page  int                    `json:"page"`
	Limit int                    `json:"limit"`
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
// @Success 201 {object} entities.Transaction
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /transactions [post]
func (h *TransactionHandler) CreateTransaction(c echo.Context) error {
	ctx := c.Request().Context()

	var req CreateTransactionRequest
	if err := c.Bind(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid request body", "error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	if err := c.Validate(req); err != nil {
		h.logger.WarnContext(ctx, "validation failed", "error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Validation failed"})
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
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create transaction"})
	}

	return c.JSON(http.StatusCreated, transaction)
}

// ListTransactions handles listing transactions with pagination
// @Summary List all transactions
// @Description Get a paginated list of transactions
// @Tags Transactions
// @Produce json
// @Security bearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} TransactionListResponse
// @Failure 401 {object} ErrorResponse
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
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to list transactions"})
	}

	response := TransactionListResponse{
		Items: transactions,
		Total: total,
		Page:  page,
		Limit: limit,
	}

	return c.JSON(http.StatusOK, response)
}

// GetTransaction handles getting a single transaction by ID
// @Summary Get a transaction by ID
// @Description Get detailed information about a specific transaction
// @Tags Transactions
// @Produce json
// @Security bearerAuth
// @Param id path int true "Transaction ID"
// @Success 200 {object} entities.Transaction
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /transactions/{id} [get]
func (h *TransactionHandler) GetTransaction(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid transaction ID"})
	}

	transaction, err := h.transactionService.GetTransaction(ctx, uint(id))
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get transaction", "error", err, "id", id)
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "Transaction not found"})
	}

	return c.JSON(http.StatusOK, transaction)
}
