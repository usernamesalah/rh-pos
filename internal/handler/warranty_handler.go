package handler

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
)

// WarrantyHandler handles public warranty check HTTP requests
type WarrantyHandler struct {
	warrantyService interfaces.WarrantyService
	logger          *slog.Logger
}

// NewWarrantyHandler creates a new warranty handler
func NewWarrantyHandler(warrantyService interfaces.WarrantyService, logger *slog.Logger) *WarrantyHandler {
	return &WarrantyHandler{
		warrantyService: warrantyService,
		logger:          logger,
	}
}

// CheckWarranty handles GET /warranty/:transaction_id
func (h *WarrantyHandler) CheckWarranty(c echo.Context) error {
	ctx := c.Request().Context()

	transactionID := c.Param("transaction_id")
	if transactionID == "" {
		return ErrorResponse(c, http.StatusBadRequest, "Transaction ID is required")
	}

	resp, err := h.warrantyService.CheckWarranty(ctx, transactionID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return ErrorResponse(c, http.StatusNotFound, "Transaction not found")
		}
		if strings.Contains(err.Error(), "invalid transaction ID") {
			return ErrorResponse(c, http.StatusBadRequest, "Invalid transaction ID")
		}
		h.logger.ErrorContext(ctx, "failed to check warranty", "error", err)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to check warranty")
	}

	return SuccessResponse(c, http.StatusOK, "Warranty information retrieved", formatWarrantyResponse(resp))
}

// SearchByPhone handles GET /warranty/search?phone=...
func (h *WarrantyHandler) SearchByPhone(c echo.Context) error {
	ctx := c.Request().Context()

	phone := c.QueryParam("phone")
	if phone == "" {
		return ErrorResponse(c, http.StatusBadRequest, "phone query parameter is required")
	}

	responses, err := h.warrantyService.SearchByPhone(ctx, phone)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to search warranties by phone", "error", err)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to search warranties")
	}

	result := make([]map[string]interface{}, len(responses))
	for i := range responses {
		result[i] = formatWarrantyResponse(&responses[i])
	}

	return SuccessResponse(c, http.StatusOK, "Warranties retrieved", map[string]interface{}{
		"transactions": result,
	})
}

func formatWarrantyResponse(resp *interfaces.WarrantyResponse) map[string]interface{} {
	items := make([]map[string]interface{}, len(resp.Items))
	for i, item := range resp.Items {
		items[i] = map[string]interface{}{
			"product_name":   item.ProductName,
			"quantity":       item.Quantity,
			"warranty_days":  item.WarrantyDays,
			"warranty_start": item.WarrantyStart.Format(time.RFC3339),
			"warranty_end":   item.WarrantyEnd.Format(time.RFC3339),
			"is_active":      item.IsActive,
			"days_remaining": item.DaysRemaining,
		}
	}

	result := map[string]interface{}{
		"transaction_id":   resp.TransactionID,
		"transaction_date": resp.TransactionDate.Format(time.RFC3339),
		"items":            items,
	}

	if resp.CustomerName != "" {
		result["customer_name"] = resp.CustomerName
	}
	if resp.CustomerEmail != "" {
		result["customer_email"] = resp.CustomerEmail
	}
	if resp.CustomerPhone != "" {
		result["customer_phone"] = resp.CustomerPhone
	}

	return result
}
