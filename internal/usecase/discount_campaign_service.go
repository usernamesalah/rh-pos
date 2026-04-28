package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"github.com/usernamesalah/rh-pos/internal/pkg/ctxkey"
)

type discountCampaignService struct {
	campaignRepo interfaces.DiscountCampaignRepository
	auditRepo    interfaces.AuditLogRepository
	logger       *slog.Logger
}

func NewDiscountCampaignService(campaignRepo interfaces.DiscountCampaignRepository, auditRepo interfaces.AuditLogRepository, logger *slog.Logger) interfaces.DiscountCampaignService {
	return &discountCampaignService{
		campaignRepo: campaignRepo,
		auditRepo:    auditRepo,
		logger:       logger,
	}
}

func validateCampaignType(campaignType string, pct float64, buyQty *int, discAmt *float64, rewardPID *uint, rewardQty *int) error {
	switch campaignType {
	case entities.CampaignTypeProductPercentageDiscount:
		if pct <= 0 || pct > 100 {
			return fmt.Errorf("discount_percentage must be between 0 and 100 for %s", campaignType)
		}
	case entities.CampaignTypeBuyXQtyGetDiscountAmount:
		if buyQty == nil || *buyQty < 2 {
			return fmt.Errorf("buy_quantity must be >= 2 for %s", campaignType)
		}
		if discAmt == nil || *discAmt <= 0 {
			return fmt.Errorf("discount_amount must be > 0 for %s", campaignType)
		}
	case entities.CampaignTypeBuyXQtyGetDiscountPercent:
		if buyQty == nil || *buyQty < 2 {
			return fmt.Errorf("buy_quantity must be >= 2 for %s", campaignType)
		}
		if pct <= 0 || pct > 100 {
			return fmt.Errorf("discount_percentage must be between 0 and 100 for %s", campaignType)
		}
	case entities.CampaignTypeBuyXProductGetYProductFree:
		if buyQty == nil || *buyQty < 1 {
			return fmt.Errorf("buy_quantity must be >= 1 for %s", campaignType)
		}
		if rewardPID == nil {
			return fmt.Errorf("reward_product_id is required for %s", campaignType)
		}
		if rewardQty == nil || *rewardQty < 1 {
			return fmt.Errorf("reward_quantity must be >= 1 for %s", campaignType)
		}
	default:
		return fmt.Errorf("unknown campaign_type: %s", campaignType)
	}
	return nil
}

func (s *discountCampaignService) Create(ctx context.Context, req interfaces.CreateCampaignRequest) (*entities.DiscountCampaign, error) {
	if req.CampaignType == "" {
		req.CampaignType = entities.CampaignTypeProductPercentageDiscount
	}

	if err := validateCampaignType(req.CampaignType, req.DiscountPercentage, req.BuyQuantity, req.DiscountAmount, req.RewardProductID, req.RewardQuantity); err != nil {
		return nil, err
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
		CampaignType:       req.CampaignType,
		DiscountPercentage: req.DiscountPercentage,
		BuyQuantity:        req.BuyQuantity,
		DiscountAmount:     req.DiscountAmount,
		RewardProductID:    req.RewardProductID,
		RewardQuantity:     req.RewardQuantity,
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

func (s *discountCampaignService) GetByID(ctx context.Context, id uint) (*entities.DiscountCampaign, error) {
	return s.campaignRepo.GetByID(ctx, id)
}

func (s *discountCampaignService) List(ctx context.Context, page, limit int) ([]entities.DiscountCampaign, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.campaignRepo.List(ctx, page, limit)
}

func (s *discountCampaignService) Update(ctx context.Context, id uint, req interfaces.UpdateCampaignRequest) (*entities.DiscountCampaign, error) {
	campaign, err := s.campaignRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	beforeJSON, err := json.Marshal(campaign)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal before state: %w", err)
	}

	if req.Name != nil {
		campaign.Name = *req.Name
	}
	if req.CampaignType != nil {
		campaign.CampaignType = *req.CampaignType
	}
	if req.DiscountPercentage != nil {
		campaign.DiscountPercentage = *req.DiscountPercentage
	}
	if req.BuyQuantity != nil {
		campaign.BuyQuantity = req.BuyQuantity
	}
	if req.DiscountAmount != nil {
		campaign.DiscountAmount = req.DiscountAmount
	}
	if req.RewardProductID != nil {
		campaign.RewardProductID = req.RewardProductID
	}
	if req.RewardQuantity != nil {
		campaign.RewardQuantity = req.RewardQuantity
	}
	if req.StartDate != nil {
		campaign.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		campaign.EndDate = *req.EndDate
	}

	if err := validateCampaignType(campaign.CampaignType, campaign.DiscountPercentage, campaign.BuyQuantity, campaign.DiscountAmount, campaign.RewardProductID, campaign.RewardQuantity); err != nil {
		return nil, err
	}

	if !campaign.EndDate.After(campaign.StartDate) {
		return nil, fmt.Errorf("end_date must be after start_date")
	}

	if err := s.campaignRepo.Update(ctx, campaign); err != nil {
		return nil, err
	}

	updated, err := s.campaignRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	afterJSON, err := json.Marshal(updated)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal after state: %w", err)
	}

	s.writeAuditLog(ctx, id, "update", beforeJSON, afterJSON)

	return updated, nil
}

func (s *discountCampaignService) Delete(ctx context.Context, id uint) error {
	campaign, err := s.campaignRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	beforeJSON, err := json.Marshal(campaign)
	if err != nil {
		return fmt.Errorf("failed to marshal before state: %w", err)
	}

	if err := s.campaignRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.writeAuditLog(ctx, id, "delete", beforeJSON, nil)

	return nil
}

func (s *discountCampaignService) AddProducts(ctx context.Context, campaignID uint, productIDs []uint) error {
	if _, err := s.campaignRepo.GetByID(ctx, campaignID); err != nil {
		return err
	}
	return s.campaignRepo.AddProducts(ctx, campaignID, productIDs)
}

func (s *discountCampaignService) RemoveProduct(ctx context.Context, campaignID uint, productID uint) error {
	if _, err := s.campaignRepo.GetByID(ctx, campaignID); err != nil {
		return err
	}
	return s.campaignRepo.RemoveProduct(ctx, campaignID, productID)
}

func (s *discountCampaignService) writeAuditLog(ctx context.Context, entityID uint, action string, before, after json.RawMessage) {
	tenantID, _ := ctxkey.TenantIDFromContext(ctx)
	userID, _ := ctxkey.UserIDFromContext(ctx)

	entry := &entities.AuditLog{
		TenantID:    tenantID,
		UserID:      userID,
		EntityType:  "discount_campaign",
		EntityID:    entityID,
		Action:      action,
		BeforeState: before,
		AfterState:  after,
	}

	if err := s.auditRepo.Create(ctx, entry); err != nil {
		s.logger.ErrorContext(ctx, "failed to write audit log",
			"entity_type", "discount_campaign",
			"entity_id", entityID,
			"action", action,
			"error", err,
		)
	}
}
