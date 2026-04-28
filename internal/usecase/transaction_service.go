package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"github.com/usernamesalah/rh-pos/internal/pkg/ctxkey"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type transactionService struct {
	transactionRepo interfaces.TransactionRepository
	productRepo     interfaces.ProductRepository
	campaignRepo    interfaces.DiscountCampaignRepository
	db              *gorm.DB
	logger          *slog.Logger
}

func NewTransactionService(transactionRepo interfaces.TransactionRepository, productRepo interfaces.ProductRepository, campaignRepo interfaces.DiscountCampaignRepository, db *gorm.DB, logger *slog.Logger) interfaces.TransactionService {
	return &transactionService{
		transactionRepo: transactionRepo,
		productRepo:     productRepo,
		campaignRepo:    campaignRepo,
		db:              db,
		logger:          logger,
	}
}

func (s *transactionService) CreateTransaction(ctx context.Context, req interfaces.CreateTransactionRequest) (*entities.Transaction, error) {
	s.logger.InfoContext(ctx, "creating transaction", "user_id", req.UserID)

	if len(req.Items) == 0 {
		return nil, fmt.Errorf("transaction must have at least one item")
	}

	productIDs := collectProductIDs(req.Items)

	campaigns, err := s.campaignRepo.GetActiveCampaignsForProducts(ctx, productIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load active campaigns: %w", err)
	}

	var createdTransaction *entities.Transaction

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tenantID, ok := ctxkey.TenantIDFromContext(ctx)
		if !ok {
			return fmt.Errorf("tenant_id not found in context")
		}

		sortedIDs := sortProductIDs(productIDs)

		productMap, err := loadProductsForUpdate(tx, tenantID, sortedIDs)
		if err != nil {
			return err
		}

		purchasedLines, freeLines, warrantyMap, err := buildPricingInputs(req.Items, productMap)
		if err != nil {
			return err
		}

		pricedLines, err := evaluatePromotions(purchasedLines, freeLines, campaigns)
		if err != nil {
			return err
		}

		transaction := &entities.Transaction{
			UserID:        &req.UserID,
			PaymentMethod: req.PaymentMethod,
			Discount:      req.Discount,
			Notes:         req.Notes,
			TenantID:      &tenantID,
			CustomerName:  req.CustomerName,
			CustomerEmail: req.CustomerEmail,
			CustomerPhone: req.CustomerPhone,
			Items:         make([]entities.TransactionItem, 0, len(pricedLines)),
		}

		var campaignTotal float64
		var regularTotal float64

		for _, pl := range pricedLines {
			product := productMap[pl.ProductID]
			lineTotal := pl.UnitPrice * float64(pl.Quantity)

			item := entities.TransactionItem{
				ProductID:          pl.ProductID,
				Quantity:           pl.Quantity,
				Price:              pl.UnitPrice,
				WarrantyDays:       warrantyMap[pl.ProductID],
				DiscountPercentage: pl.DiscountPercentage,
				DiscountAmount:     pl.DiscountAmount,
				CampaignID:         pl.CampaignID,
				CampaignType:       pl.CampaignType,
				IsFreeItem:         pl.IsFreeItem,
				CampaignGroupKey:   pl.CampaignGroupKey,
			}

			if pl.CampaignID != nil || pl.IsFreeItem {
				campaignTotal += lineTotal
			} else {
				regularTotal += lineTotal
			}

			transaction.Items = append(transaction.Items, item)

			if product.Stock < pl.Quantity {
				return fmt.Errorf("insufficient stock for product %s (available: %d, requested: %d)", product.Name, product.Stock, pl.Quantity)
			}

			result := tx.Model(&entities.Product{}).
				Where("id = ? AND tenant_id = ? AND stock >= ?", pl.ProductID, tenantID, pl.Quantity).
				Update("stock", gorm.Expr("stock - ?", pl.Quantity))
			if result.Error != nil {
				return fmt.Errorf("failed to update product stock: %w", result.Error)
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("insufficient stock for product %s", product.Name)
			}

			product.Stock -= pl.Quantity
		}

		if transaction.Discount > 0 {
			regularTotal = regularTotal * (1 - transaction.Discount/100)
		}

		transaction.TotalPrice = roundTo2(campaignTotal + regularTotal)

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

	return s.transactionRepo.GetByID(ctx, createdTransaction.ID)
}

func (s *transactionService) GetTransaction(ctx context.Context, id uint) (*entities.Transaction, error) {
	s.logger.InfoContext(ctx, "getting transaction", "id", id)

	transaction, err := s.transactionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	return transaction, nil
}

func (s *transactionService) ListTransactions(ctx context.Context, page, limit int) ([]entities.Transaction, int64, error) {
	s.logger.InfoContext(ctx, "listing transactions", "page", page, "limit", limit)

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

func collectProductIDs(items []interfaces.TransactionItemRequest) []uint {
	seen := make(map[uint]bool)
	var ids []uint
	for _, item := range items {
		if !seen[item.ProductID] {
			seen[item.ProductID] = true
			ids = append(ids, item.ProductID)
		}
	}
	return ids
}

func sortProductIDs(ids []uint) []uint {
	sorted := make([]uint, len(ids))
	copy(sorted, ids)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	return sorted
}

func loadProductsForUpdate(tx *gorm.DB, tenantID uint, sortedProductIDs []uint) (map[uint]*entities.Product, error) {
	var products []entities.Product
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id IN ? AND tenant_id = ?", sortedProductIDs, tenantID).
		Order("id ASC").
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to load products for update: %w", err)
	}

	m := make(map[uint]*entities.Product, len(products))
	for i := range products {
		m[products[i].ID] = &products[i]
	}

	for _, pid := range sortedProductIDs {
		if _, ok := m[pid]; !ok {
			return nil, fmt.Errorf("product %d not found", pid)
		}
	}

	return m, nil
}

func buildPricingInputs(items []interfaces.TransactionItemRequest, productMap map[uint]*entities.Product) ([]pricingLineInput, []pricingLineInput, map[uint]int, error) {
	var purchased []pricingLineInput
	var free []pricingLineInput
	warrantyMap := make(map[uint]int)

	for _, item := range items {
		product, ok := productMap[item.ProductID]
		if !ok {
			return nil, nil, nil, fmt.Errorf("product %d not found in loaded products", item.ProductID)
		}

		var unitPrice float64
		if product.HargaJual != nil {
			unitPrice = *product.HargaJual
		}

		if item.WarrantyDays > 0 {
			warrantyMap[item.ProductID] = item.WarrantyDays
		}

		line := pricingLineInput{
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  unitPrice,
			IsFreeItem: item.IsFreeItem,
			CampaignID: item.CampaignID,
		}

		if item.IsFreeItem {
			free = append(free, line)
		} else {
			purchased = append(purchased, line)
		}
	}

	return purchased, free, warrantyMap, nil
}
