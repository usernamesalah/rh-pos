package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
)

type ReportHandler struct {
	reportService interfaces.ReportService
	logger        *slog.Logger
}

// NewReportHandler creates a new report handler
func NewReportHandler(reportService interfaces.ReportService, logger *slog.Logger) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
		logger:        logger,
	}
}

// GetSalesReport handles getting sales report
// @Summary Get sales report
// @Description Get sales report for a specific date range
// @Tags Reports
// @Produce json
// @Security bearerAuth
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} interfaces.ReportResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /reports [get]
func (h *ReportHandler) GetSalesReport(c echo.Context) error {
	ctx := c.Request().Context()

	startDateStr := c.QueryParam("start_date")
	endDateStr := c.QueryParam("end_date")

	if startDateStr == "" || endDateStr == "" {
		return ErrorResponse(c, http.StatusBadRequest, "start_date and end_date are required")
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid start_date format, use YYYY-MM-DD")
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid end_date format, use YYYY-MM-DD")
	}

	// Set end date to end of day
	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	if startDate.After(endDate) {
		return ErrorResponse(c, http.StatusBadRequest, "start_date must be before or equal to end_date")
	}

	report, err := h.reportService.GetSalesReport(ctx, startDate, endDate)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get sales report", "error", err)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to get sales report")
	}

	// Convert details to HashIDResponse
	details := make([]HashIDResponse, len(report.Details))
	for i, detail := range report.Details {
		details[i] = WithHashID(
			detail.ID,
			"", // No created_at for report details
			"", // No updated_at for report details
			map[string]interface{}{
				"product_id":   detail.ProductID,
				"product_name": detail.ProductName,
				"total":        detail.Total,
				"total_price":  detail.TotalPrice,
			},
		)
	}

	response := map[string]interface{}{
		"total_revenue":       report.TotalRevenue,
		"items_sold":          report.ItemsSold,
		"average_transaction": report.AverageTransaction,
		"details":             details,
	}

	return SuccessResponse(c, http.StatusOK, "Sales report retrieved successfully", response)
}
