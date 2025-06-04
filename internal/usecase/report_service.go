package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
)

type reportService struct {
	transactionRepo interfaces.TransactionRepository
	logger          *slog.Logger
}

// NewReportService creates a new report service
func NewReportService(transactionRepo interfaces.TransactionRepository, logger *slog.Logger) interfaces.ReportService {
	return &reportService{
		transactionRepo: transactionRepo,
		logger:          logger,
	}
}

// GetSalesReport generates a sales report for the given date range
func (s *reportService) GetSalesReport(ctx context.Context, startDate, endDate time.Time) (*interfaces.ReportResponse, error) {
	s.logger.InfoContext(ctx, "generating sales report", "start_date", startDate, "end_date", endDate)

	// Get report data from repository
	details, err := s.transactionRepo.GetReportData(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get report data: %w", err)
	}

	// Calculate aggregated metrics
	var totalRevenue float64
	var itemsSold int

	for _, detail := range details {
		totalRevenue += detail.TotalPrice
		itemsSold += detail.Total
	}

	// Calculate average transaction value
	var averageTransaction float64
	if len(details) > 0 {
		averageTransaction = totalRevenue / float64(len(details))
	}

	response := &interfaces.ReportResponse{
		TotalRevenue:       totalRevenue,
		ItemsSold:          itemsSold,
		AverageTransaction: averageTransaction,
		Details:            details,
	}

	return response, nil
}
