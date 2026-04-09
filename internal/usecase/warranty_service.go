package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"github.com/usernamesalah/rh-pos/internal/pkg/hash"
	"github.com/usernamesalah/rh-pos/internal/pkg/masking"
)

type warrantyService struct {
	transactionRepo interfaces.TransactionRepository
	logger          *slog.Logger
}

// NewWarrantyService creates a new warranty service
func NewWarrantyService(transactionRepo interfaces.TransactionRepository, logger *slog.Logger) interfaces.WarrantyService {
	return &warrantyService{
		transactionRepo: transactionRepo,
		logger:          logger,
	}
}

// CheckWarranty checks warranty for a transaction by hashed transaction ID
func (s *warrantyService) CheckWarranty(ctx context.Context, hashedTransactionID string) (*interfaces.WarrantyResponse, error) {
	id, err := hash.DecodeHashID(hashedTransactionID)
	if err != nil {
		return nil, fmt.Errorf("invalid transaction ID")
	}

	transaction, err := s.transactionRepo.GetByIDWithoutTenant(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	resp := &interfaces.WarrantyResponse{
		TransactionID:   hashedTransactionID,
		TransactionDate: transaction.CreatedAt,
		Items:           buildWarrantyItems(transaction.CreatedAt, now, transaction.Items),
	}

	if transaction.CustomerName != nil && *transaction.CustomerName != "" {
		resp.CustomerName = masking.MaskName(*transaction.CustomerName)
	}
	if transaction.CustomerEmail != nil && *transaction.CustomerEmail != "" {
		resp.CustomerEmail = masking.MaskEmail(*transaction.CustomerEmail)
	}
	if transaction.CustomerPhone != nil && *transaction.CustomerPhone != "" {
		resp.CustomerPhone = masking.MaskPhone(*transaction.CustomerPhone)
	}

	return resp, nil
}

// SearchByPhone searches warranties by customer phone number
func (s *warrantyService) SearchByPhone(ctx context.Context, phone string) ([]interfaces.WarrantyResponse, error) {
	transactions, err := s.transactionRepo.SearchByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	responses := make([]interfaces.WarrantyResponse, 0, len(transactions))

	for _, t := range transactions {
		items := buildWarrantyItems(t.CreatedAt, now, t.Items)
		if len(items) == 0 {
			continue
		}

		resp := interfaces.WarrantyResponse{
			TransactionID:   hash.HashID(t.ID),
			TransactionDate: t.CreatedAt,
			Items:           items,
		}

		if t.CustomerName != nil && *t.CustomerName != "" {
			resp.CustomerName = masking.MaskName(*t.CustomerName)
		}
		if t.CustomerEmail != nil && *t.CustomerEmail != "" {
			resp.CustomerEmail = masking.MaskEmail(*t.CustomerEmail)
		}
		if t.CustomerPhone != nil && *t.CustomerPhone != "" {
			resp.CustomerPhone = masking.MaskPhone(*t.CustomerPhone)
		}

		responses = append(responses, resp)
	}

	return responses, nil
}

// buildWarrantyItems builds warranty item responses filtering only items with warranty_days > 0
func buildWarrantyItems(transactionDate, now time.Time, items []entities.TransactionItem) []interfaces.WarrantyItemResponse {
	result := make([]interfaces.WarrantyItemResponse, 0, len(items))
	for _, item := range items {
		if item.WarrantyDays <= 0 {
			continue
		}
		start := transactionDate
		end := start.AddDate(0, 0, item.WarrantyDays)
		isActive := now.Before(end)
		daysRemaining := 0
		if isActive {
			daysRemaining = int(end.Sub(now).Hours() / 24)
		}
		result = append(result, interfaces.WarrantyItemResponse{
			ProductName:   item.Product.Name,
			Quantity:      item.Quantity,
			WarrantyDays:  item.WarrantyDays,
			WarrantyStart: start,
			WarrantyEnd:   end,
			IsActive:      isActive,
			DaysRemaining: daysRemaining,
		})
	}
	return result
}
