package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"github.com/usernamesalah/rh-pos/internal/pkg/ctxkey"
)

type discountCampaignService struct {
	campaignRepo interfaces.DiscountCampaignRepository
	logger       *slog.Logger
}

// NewDiscountCampaignService creates a new discount campaign service
func NewDiscountCampaignService(campaignRepo interfaces.DiscountCampaignRepository, logger *slog.Logger) interfaces.DiscountCampaignService {
	return &discountCampaignService{
		campaignRepo: campaignRepo,
		logger:       logger,
	}
}

// Create creates a new discount campaign
func (s *discountCampaignService) Create(ctx context.Context, req interfaces.CreateCampaignRequest) (*entities.DiscountCampaign, error) {
	if req.DiscountPercentage <= 0 || req.DiscountPercentage > 100 {
		return nil, fmt.Errorf("discount_percentage must be between 0 and 100")
	}
	if !req.EndDate.After(req.StartDate) {
		return nil, fmt.Errorf("end_date must be after start_date")
	}

	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant_id not found in context")
	}

	campaign := &entities.DiscountCampaign{
		Name:               req.Name,
		DiscountPercentage: req.DiscountPercentage,
		StartDate:          req.StartDate,
		EndDate:            req.EndDate,
		TenantID:           &tenantID,
	}

	if err := s.campaignRepo.Create(ctx, campaign); err != nil {
		return nil, err
	}

	if len(req.ProductIDs) > 0 {
		if err := s.campaignRepo.AddProducts(ctx, campaign.ID, req.ProductIDs); err != nil {
			return nil, err
		}
	}

	return s.campaignRepo.GetByID(ctx, campaign.ID)
}

// GetByID retrieves a campaign by ID
func (s *discountCampaignService) GetByID(ctx context.Context, id uint) (*entities.DiscountCampaign, error) {
	return s.campaignRepo.GetByID(ctx, id)
}

// List retrieves campaigns with pagination
func (s *discountCampaignService) List(ctx context.Context, page, limit int) ([]entities.DiscountCampaign, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.campaignRepo.List(ctx, page, limit)
}

// Update updates a campaign
func (s *discountCampaignService) Update(ctx context.Context, id uint, req interfaces.UpdateCampaignRequest) (*entities.DiscountCampaign, error) {
	campaign, err := s.campaignRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		campaign.Name = *req.Name
	}
	if req.DiscountPercentage != nil {
		if *req.DiscountPercentage <= 0 || *req.DiscountPercentage > 100 {
			return nil, fmt.Errorf("discount_percentage must be between 0 and 100")
		}
		campaign.DiscountPercentage = *req.DiscountPercentage
	}
	if req.StartDate != nil {
		campaign.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		campaign.EndDate = *req.EndDate
	}

	if !campaign.EndDate.After(campaign.StartDate) {
		return nil, fmt.Errorf("end_date must be after start_date")
	}

	if err := s.campaignRepo.Update(ctx, campaign); err != nil {
		return nil, err
	}

	return s.campaignRepo.GetByID(ctx, id)
}

// Delete deletes a campaign
func (s *discountCampaignService) Delete(ctx context.Context, id uint) error {
	return s.campaignRepo.Delete(ctx, id)
}

// AddProducts adds products to a campaign
func (s *discountCampaignService) AddProducts(ctx context.Context, campaignID uint, productIDs []uint) error {
	if _, err := s.campaignRepo.GetByID(ctx, campaignID); err != nil {
		return err
	}
	return s.campaignRepo.AddProducts(ctx, campaignID, productIDs)
}

// RemoveProduct removes a product from a campaign
func (s *discountCampaignService) RemoveProduct(ctx context.Context, campaignID uint, productID uint) error {
	if _, err := s.campaignRepo.GetByID(ctx, campaignID); err != nil {
		return err
	}
	return s.campaignRepo.RemoveProduct(ctx, campaignID, productID)
}
