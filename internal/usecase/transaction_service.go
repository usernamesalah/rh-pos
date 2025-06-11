package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"gorm.io/gorm"
)

type transactionService struct {
	transactionRepo interfaces.TransactionRepository
	productRepo     interfaces.ProductRepository
	db              *gorm.DB
	logger          *slog.Logger
}

// NewTransactionService creates a new transaction service
func NewTransactionService(transactionRepo interfaces.TransactionRepository, productRepo interfaces.ProductRepository, db *gorm.DB, logger *slog.Logger) interfaces.TransactionService {
	return &transactionService{
		transactionRepo: transactionRepo,
		productRepo:     productRepo,
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
		// Create transaction entity
		transaction := &entities.Transaction{
			User:          req.User,
			PaymentMethod: req.PaymentMethod,
			Discount:      req.Discount,
			Notes:         req.Notes,
			Items:         make([]entities.TransactionItem, 0, len(req.Items)),
		}

		// Calculate total price from products
		var calculatedTotal float64

		// Process each item
		for _, item := range req.Items {
			// Validate product exists and has sufficient stock
			product, err := s.productRepo.GetByID(ctx, item.ProductID)
			if err != nil {
				return fmt.Errorf("product not found: %w", err)
			}

			if product.Stock < item.Quantity {
				return fmt.Errorf("insufficient stock for product %s: requested %d, available %d",
					product.Name, item.Quantity, product.Stock)
			}

			// Calculate item total
			itemTotal := product.HargaJual * float64(item.Quantity)
			calculatedTotal += itemTotal

			// Create transaction item
			transactionItem := entities.TransactionItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     product.HargaJual,
			}

			transaction.Items = append(transaction.Items, transactionItem)

			// Update product stock within the transaction
			newStock := product.Stock - item.Quantity
			if err := tx.Model(&entities.Product{}).Where("id = ?", item.ProductID).Update("stock", newStock).Error; err != nil {
				return fmt.Errorf("failed to update product stock: %w", err)
			}
		}

		// Apply discount if any
		if transaction.Discount > 0 {
			calculatedTotal = calculatedTotal * (1 - transaction.Discount/100)
		}

		// Validate total price matches calculated total
		if req.TotalPrice != calculatedTotal {
			return fmt.Errorf("total price mismatch: provided %.2f, calculated %.2f", req.TotalPrice, calculatedTotal)
		}

		// Set the validated total price
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
