package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"github.com/usernamesalah/rh-pos/internal/pkg/ctxkey"
	"gorm.io/gorm"
)

type transactionService struct {
	transactionRepo interfaces.TransactionRepository
	productRepo     interfaces.ProductRepository
	campaignRepo    interfaces.DiscountCampaignRepository
	db              *gorm.DB
	logger          *slog.Logger
}

// NewTransactionService creates a new transaction service
func NewTransactionService(transactionRepo interfaces.TransactionRepository, productRepo interfaces.ProductRepository, campaignRepo interfaces.DiscountCampaignRepository, db *gorm.DB, logger *slog.Logger) interfaces.TransactionService {
	return &transactionService{
		transactionRepo: transactionRepo,
		productRepo:     productRepo,
		campaignRepo:    campaignRepo,
		db:              db,
		logger:          logger,
	}
}

// CreateTransaction creates a new transaction with database transaction support
func (s *transactionService) CreateTransaction(ctx context.Context, req interfaces.CreateTransactionRequest) (*entities.Transaction, error) {
	s.logger.InfoContext(ctx, "creating transaction", "user_id", req.UserID)

	// Validate request
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("transaction must have at least one item")
	}

	var createdTransaction *entities.Transaction

	// Use database transaction to ensure data consistency
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get tenant_id from context
		tenantID, ok := ctxkey.TenantIDFromContext(ctx)
		if !ok {
			return fmt.Errorf("tenant_id not found in context")
		}

		// Create transaction entity
		transaction := &entities.Transaction{
			UserID:        &req.UserID,
			PaymentMethod: req.PaymentMethod,
			Discount:      req.Discount,
			Notes:         req.Notes,
			TenantID:      &tenantID,
			CustomerName:  req.CustomerName,
			CustomerEmail: req.CustomerEmail,
			CustomerPhone: req.CustomerPhone,
			Items:         make([]entities.TransactionItem, 0, len(req.Items)),
		}

		// Calculate total price from products
		var campaignDiscountedTotal float64 // sum of items with campaign discounts (already discounted)
		var regularTotal float64            // sum of items without campaign discounts (before transaction-level discount)

		// Process each item
		for _, item := range req.Items {
			product, err := s.productRepo.GetByID(ctx, item.ProductID)
			if err != nil {
				return fmt.Errorf("product not found: %w", err)
			}

			// If dynamic price product but no price provided, use 0
			if product.IsDynamicPrice && item.Price == nil {
				defaultPrice := 0.0
				item.Price = &defaultPrice
			}

			if product.IsDynamicPrice && item.Price != nil && *item.Price < 0 {
				zeroPrice := 0.0
				item.Price = &zeroPrice
			}

			itemPrice := ResolveItemPrice(product, item.Price)

			transactionItem := entities.TransactionItem{
				ProductID:    item.ProductID,
				Quantity:     item.Quantity,
				WarrantyDays: item.WarrantyDays,
				Price:        itemPrice,
			}

			if product.IsDynamicPrice {
				// Dynamic price product: no campaign discount, no stock deduction
				regularTotal += itemPrice * float64(item.Quantity)
			} else {
				// Regular product: check campaigns, deduct stock
				campaigns, err := s.campaignRepo.GetActiveCampaignsForProduct(ctx, item.ProductID)
				if err != nil {
					return fmt.Errorf("failed to check campaigns for product: %w", err)
				}

				if len(campaigns) > 0 {
					best := campaigns[0]
					for _, c := range campaigns[1:] {
						if c.DiscountPercentage > best.DiscountPercentage {
							best = c
						}
					}
					discountedPrice := itemPrice * (1 - best.DiscountPercentage/100)
					transactionItem.Price = discountedPrice
					transactionItem.DiscountPercentage = best.DiscountPercentage
					transactionItem.CampaignID = &best.ID
					campaignDiscountedTotal += discountedPrice * float64(item.Quantity)
				} else {
					regularTotal += itemPrice * float64(item.Quantity)
				}

				// Atomically deduct stock only if sufficient quantity exists
				result := tx.Model(&entities.Product{}).
					Where("id = ? AND tenant_id = ? AND stock >= ?", item.ProductID, tenantID, item.Quantity).
					Update("stock", gorm.Expr("stock - ?", item.Quantity))
				if result.Error != nil {
					return fmt.Errorf("failed to update product stock: %w", result.Error)
				}
				if result.RowsAffected == 0 {
					return fmt.Errorf("insufficient stock for product %s", product.Name)
				}
			}

			transaction.Items = append(transaction.Items, transactionItem)
		}

		// Apply transaction-level discount only to regular (non-campaign) items
		if transaction.Discount > 0 {
			regularTotal = regularTotal * (1 - transaction.Discount/100)
		}

calculatedTotal := campaignDiscountedTotal + regularTotal

		// Use user-provided total_price if > 0, otherwise use calculated total
	// This supports dynamic price products where price comes from user input
	if req.TotalPrice > 0 {
		transaction.TotalPrice = req.TotalPrice
	} else {
		transaction.TotalPrice = calculatedTotal
	}

		// Create transaction within the DB transaction
		if err := tx.Create(transaction).Error; err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		createdTransaction = transaction
		return nil
	})

	if err != nil {
		s.logger.ErrorContext(ctx, "transaction failed", "error", err)
		return nil, err
	}

	// Return transaction with populated items
	return s.transactionRepo.GetByID(ctx, createdTransaction.ID)
}

// GetTransaction retrieves a transaction by ID
func (s *transactionService) GetTransaction(ctx context.Context, id uint) (*entities.Transaction, error) {
	s.logger.InfoContext(ctx, "getting transaction", "id", id)

	transaction, err := s.transactionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return transaction, nil
}

// ResolveItemPrice returns the price to use for a transaction item.
// For dynamic price products, uses requestPrice (default 0.0 if nil).
// For regular products, uses HargaJual (default 0.0 if nil).
func ResolveItemPrice(product *entities.Product, requestPrice *float64) float64 {
	if product.IsDynamicPrice {
		if requestPrice != nil {
			return *requestPrice
		}
		return 0.0
	}
	if product.HargaJual != nil {
		return *product.HargaJual
	}
	return 0.0
}

// ListTransactions retrieves transactions with pagination
func (s *transactionService) ListTransactions(ctx context.Context, page, limit int, startDate, endDate *time.Time, search string) ([]entities.Transaction, int64, error) {
	s.logger.InfoContext(ctx, "listing transactions", "page", page, "limit", limit, "start_date", startDate, "end_date", endDate, "search", search)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	transactions, total, err := s.transactionRepo.List(ctx, page, limit, startDate, endDate, search)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list transactions: %w", err)
	}

	return transactions, total, nil
}
