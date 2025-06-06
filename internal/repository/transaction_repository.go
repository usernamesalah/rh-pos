package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"gorm.io/gorm"
)

type transactionRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *gorm.DB, logger *slog.Logger) interfaces.TransactionRepository {
	return &transactionRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new transaction
func (r *transactionRepository) Create(ctx context.Context, transaction *entities.Transaction) error {
	r.logger.InfoContext(ctx, "creating transaction", "user", transaction.User)

	if err := r.db.WithContext(ctx).Create(transaction).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to create transaction", "error", err)
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a transaction by ID
func (r *transactionRepository) GetByID(ctx context.Context, id uint) (*entities.Transaction, error) {
	r.logger.InfoContext(ctx, "getting transaction by ID", "id", id)

	var transaction entities.Transaction
	if err := r.db.WithContext(ctx).Preload("Items.Product").Where("id = ?", id).First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("transaction not found: %w", err)
		}
		r.logger.ErrorContext(ctx, "failed to get transaction", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &transaction, nil
}

// List retrieves transactions with pagination
func (r *transactionRepository) List(ctx context.Context, page, limit int) ([]entities.Transaction, int64, error) {
	r.logger.InfoContext(ctx, "listing transactions", "page", page, "limit", limit)

	var transactions []entities.Transaction
	var total int64

	// Count total transactions
	if err := r.db.WithContext(ctx).Model(&entities.Transaction{}).Count(&total).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to count transactions", "error", err)
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	// Get transactions with pagination
	offset := (page - 1) * limit
	if err := r.db.WithContext(ctx).Preload("Items.Product").Offset(offset).Limit(limit).Find(&transactions).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to list transactions", "error", err)
		return nil, 0, fmt.Errorf("failed to list transactions: %w", err)
	}

	return transactions, total, nil
}

// GetReportData retrieves report data for the given date range
func (r *transactionRepository) GetReportData(ctx context.Context, startDate, endDate time.Time) ([]interfaces.ReportDetail, error) {
	r.logger.InfoContext(ctx, "getting report data", "start_date", startDate, "end_date", endDate)

	var reportDetails []interfaces.ReportDetail

	query := `
		SELECT 
			ti.id,
			ti.product_id,
			p.name as product_name,
			SUM(ti.quantity) as total,
			SUM(ti.price * ti.quantity) as total_price
		FROM transaction_items ti
		JOIN transactions t ON ti.transaction_id = t.id
		JOIN products p ON ti.product_id = p.id
		WHERE t.created_at BETWEEN ? AND ?
		GROUP BY ti.product_id, p.name
		ORDER BY total_price DESC
	`

	if err := r.db.WithContext(ctx).Raw(query, startDate, endDate).Scan(&reportDetails).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to get report data", "error", err)
		return nil, fmt.Errorf("failed to get report data: %w", err)
	}

	return reportDetails, nil
}
