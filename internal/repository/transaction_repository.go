package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"github.com/usernamesalah/rh-pos/internal/pkg/ctxkey"
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
	r.logger.InfoContext(ctx, "creating transaction", "user_id", transaction.UserID)

	if err := r.db.WithContext(ctx).Create(transaction).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to create transaction", "error", err)
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a transaction by ID
func (r *transactionRepository) GetByID(ctx context.Context, id uint) (*entities.Transaction, error) {
	r.logger.InfoContext(ctx, "getting transaction by ID", "id", id)

	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant_id not found in context")
	}

	var transaction entities.Transaction
	if err := r.db.WithContext(ctx).Preload("Items.Product").Where("id = ? AND tenant_id = ?", id, tenantID).First(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("transaction not found: %w", err)
		}
		r.logger.ErrorContext(ctx, "failed to get transaction", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return &transaction, nil
}

// List retrieves transactions with pagination
func (r *transactionRepository) List(ctx context.Context, page, limit int, startDate, endDate *time.Time, search string) ([]entities.Transaction, int64, error) {
	r.logger.InfoContext(ctx, "listing transactions", "page", page, "limit", limit, "start_date", startDate, "end_date", endDate, "search", search)

	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return nil, 0, fmt.Errorf("tenant_id not found in context")
	}

	var transactions []entities.Transaction
	var total int64

	baseQuery := r.db.WithContext(ctx).Model(&entities.Transaction{}).Where("tenant_id = ?", tenantID)

	if startDate != nil && endDate != nil {
		endDateWithTime := endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		baseQuery = baseQuery.Where("created_at BETWEEN ? AND ?", startDate, endDateWithTime)
	}

	if search != "" {
		searchPattern := "%" + search + "%"
		subQuery := r.db.WithContext(ctx).
			Table("transactions").
			Select("transactions.id").
			Joins("LEFT JOIN users ON users.id = transactions.user_id").
			Joins("LEFT JOIN transaction_items ON transaction_items.transaction_id = transactions.id").
			Joins("LEFT JOIN products ON products.id = transaction_items.product_id").
			Where("transactions.tenant_id = ?", tenantID).
			Where("transactions.id LIKE ? OR users.username LIKE ? OR products.name LIKE ?", searchPattern, searchPattern, searchPattern)

		if startDate != nil && endDate != nil {
			endDateWithTime := endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			subQuery = subQuery.Where("transactions.created_at BETWEEN ? AND ?", startDate, endDateWithTime)
		}

		subQuery = subQuery.Group("transactions.id")
		baseQuery = baseQuery.Where("transactions.id IN (?)", subQuery)
	}

	if err := baseQuery.Count(&total).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to count transactions", "error", err)
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	offset := (page - 1) * limit
	if err := baseQuery.Preload("Items.Product").Preload("User").Order("created_at DESC").Offset(offset).Limit(limit).Find(&transactions).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to list transactions", "error", err)
		return nil, 0, fmt.Errorf("failed to list transactions: %w", err)
	}

	return transactions, total, nil
}

// GetReportData retrieves report data for the given date range
func (r *transactionRepository) GetReportData(ctx context.Context, startDate, endDate time.Time) ([]interfaces.ReportDetail, error) {
	r.logger.InfoContext(ctx, "getting report data", "start_date", startDate, "end_date", endDate)

	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant_id not found in context")
	}

	var reportDetails []interfaces.ReportDetail

	query := `
		SELECT
			ti.product_id,
			p.name as product_name,
			SUM(ti.quantity) as total,
			SUM(ti.price * ti.quantity) as total_price
		FROM transaction_items ti
		JOIN transactions t ON ti.transaction_id = t.id
		JOIN products p ON ti.product_id = p.id
		WHERE t.created_at BETWEEN ? AND ? AND t.tenant_id = ?
		GROUP BY ti.product_id, p.name
		ORDER BY total_price DESC
	`

	if err := r.db.WithContext(ctx).Raw(query, startDate, endDate, tenantID).Scan(&reportDetails).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to get report data", "error", err)
		return nil, fmt.Errorf("failed to get report data: %w", err)
	}

	return reportDetails, nil
}

// GetTransactionCount returns the number of distinct transactions in a date range for the tenant.
func (r *transactionRepository) GetTransactionCount(ctx context.Context, startDate, endDate time.Time) (int64, error) {
	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return 0, fmt.Errorf("tenant_id not found in context")
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.Transaction{}).
		Where("created_at BETWEEN ? AND ? AND tenant_id = ?", startDate, endDate, tenantID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	return count, nil
}

// GetByIDWithoutTenant retrieves a transaction by ID without tenant filtering (for public warranty endpoints)
func (r *transactionRepository) GetByIDWithoutTenant(ctx context.Context, id uint) (*entities.Transaction, error) {
	var transaction entities.Transaction
	if err := r.db.WithContext(ctx).Preload("Items.Product").Where("id = ?", id).First(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("transaction not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	return &transaction, nil
}

// SearchByPhone searches transactions by customer phone (no tenant filtering, for public warranty endpoints)
func (r *transactionRepository) SearchByPhone(ctx context.Context, phone string) ([]entities.Transaction, error) {
	var transactions []entities.Transaction
	if err := r.db.WithContext(ctx).Preload("Items.Product").Where("customer_phone = ?", phone).Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("failed to search transactions by phone: %w", err)
	}
	return transactions, nil
}

// Delete deletes a transaction
func (r *transactionRepository) Delete(ctx context.Context, id uint) error {
	r.logger.InfoContext(ctx, "deleting transaction", "id", id)

	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant_id not found in context")
	}

	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&entities.Transaction{}).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to delete transaction", "error", err, "id", id)
		return fmt.Errorf("failed to delete transaction: %w", err)
	}
	return nil
}

// Update updates a transaction
func (r *transactionRepository) Update(ctx context.Context, transaction *entities.Transaction) error {
	r.logger.InfoContext(ctx, "updating transaction", "id", transaction.ID)

	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant_id not found in context")
	}

	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", transaction.ID, tenantID).Save(transaction).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to update transaction", "error", err, "id", transaction.ID)
		return fmt.Errorf("failed to update transaction: %w", err)
	}
	return nil
}
