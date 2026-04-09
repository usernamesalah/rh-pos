package usecase

import (
	"context"
	"fmt"
	"log/slog"

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
	s.logger.InfoContext(ctx, "creating transaction", "user", req.User)

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
			User:          req.User,
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
			// Read product for price and name — stock check is done atomically below
			product, err := s.productRepo.GetByID(ctx, item.ProductID)
			if err != nil {
				return fmt.Errorf("product not found: %w", err)
			}

			transactionItem := entities.TransactionItem{
				ProductID:    item.ProductID,
				Quantity:     item.Quantity,
				WarrantyDays: item.WarrantyDays,
				Price:        product.HargaJual,
			}

			// Check for active campaign discount
			campaigns, err := s.campaignRepo.GetActiveCampaignsForProduct(ctx, item.ProductID)
			if err != nil {
				return fmt.Errorf("failed to check campaigns for product: %w", err)
			}

			if len(campaigns) > 0 {
				// Use highest discount percentage
				best := campaigns[0]
				for _, c := range campaigns[1:] {
					if c.DiscountPercentage > best.DiscountPercentage {
						best = c
					}
				}
				discountedPrice := product.HargaJual * (1 - best.DiscountPercentage/100)
				transactionItem.Price = discountedPrice
				transactionItem.DiscountPercentage = best.DiscountPercentage
				transactionItem.CampaignID = &best.ID
				campaignDiscountedTotal += discountedPrice * float64(item.Quantity)
			} else {
				regularTotal += product.HargaJual * float64(item.Quantity)
			}

			transaction.Items = append(transaction.Items, transactionItem)

			// Atomically deduct stock only if sufficient quantity exists.
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

		// Apply transaction-level discount only to regular (non-campaign) items
		if transaction.Discount > 0 {
			regularTotal = regularTotal * (1 - transaction.Discount/100)
		}

		calculatedTotal := campaignDiscountedTotal + regularTotal

		// Set the server-calculated total price (authoritative)
		transaction.TotalPrice = calculatedTotal

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

// ListTransactions retrieves transactions with pagination
func (s *transactionService) ListTransactions(ctx context.Context, page, limit int) ([]entities.Transaction, int64, error) {
	s.logger.InfoContext(ctx, "listing transactions", "page", page, "limit", limit)

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	transactions, total, err := s.transactionRepo.List(ctx, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list transactions: %w", err)
	}

	return transactions, total, nil
}
