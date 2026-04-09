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

type discountCampaignRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewDiscountCampaignRepository creates a new discount campaign repository
func NewDiscountCampaignRepository(db *gorm.DB, logger *slog.Logger) interfaces.DiscountCampaignRepository {
	return &discountCampaignRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new discount campaign
func (r *discountCampaignRepository) Create(ctx context.Context, campaign *entities.DiscountCampaign) error {
	if err := r.db.WithContext(ctx).Create(campaign).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to create discount campaign", "error", err)
		return fmt.Errorf("failed to create discount campaign: %w", err)
	}
	return nil
}

// GetByID retrieves a discount campaign by ID (tenant-scoped)
func (r *discountCampaignRepository) GetByID(ctx context.Context, id uint) (*entities.DiscountCampaign, error) {
	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant_id not found in context")
	}

	var campaign entities.DiscountCampaign
	if err := r.db.WithContext(ctx).
		Preload("Products.Product").
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&campaign).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("discount campaign not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get discount campaign: %w", err)
	}
	return &campaign, nil
}

// List retrieves discount campaigns with pagination (tenant-scoped)
func (r *discountCampaignRepository) List(ctx context.Context, page, limit int) ([]entities.DiscountCampaign, int64, error) {
	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return nil, 0, fmt.Errorf("tenant_id not found in context")
	}

	var campaigns []entities.DiscountCampaign
	var total int64

	if err := r.db.WithContext(ctx).Model(&entities.DiscountCampaign{}).Where("tenant_id = ?", tenantID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count campaigns: %w", err)
	}

	offset := (page - 1) * limit
	if err := r.db.WithContext(ctx).
		Preload("Products.Product").
		Where("tenant_id = ?", tenantID).
		Offset(offset).Limit(limit).
		Find(&campaigns).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list campaigns: %w", err)
	}

	return campaigns, total, nil
}

// Update updates a discount campaign
func (r *discountCampaignRepository) Update(ctx context.Context, campaign *entities.DiscountCampaign) error {
	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant_id not found in context")
	}

	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", campaign.ID, tenantID).Save(campaign).Error; err != nil {
		return fmt.Errorf("failed to update discount campaign: %w", err)
	}
	return nil
}

// Delete deletes a discount campaign
func (r *discountCampaignRepository) Delete(ctx context.Context, id uint) error {
	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant_id not found in context")
	}

	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&entities.DiscountCampaign{}).Error; err != nil {
		return fmt.Errorf("failed to delete discount campaign: %w", err)
	}
	return nil
}

// AddProducts adds products to a campaign
func (r *discountCampaignRepository) AddProducts(ctx context.Context, campaignID uint, productIDs []uint) error {
	records := make([]entities.DiscountCampaignProduct, len(productIDs))
	for i, pid := range productIDs {
		records[i] = entities.DiscountCampaignProduct{
			CampaignID: campaignID,
			ProductID:  pid,
		}
	}
	if err := r.db.WithContext(ctx).Save(&records).Error; err != nil {
		return fmt.Errorf("failed to add products to campaign: %w", err)
	}
	return nil
}

// RemoveProduct removes a product from a campaign
func (r *discountCampaignRepository) RemoveProduct(ctx context.Context, campaignID uint, productID uint) error {
	if err := r.db.WithContext(ctx).
		Where("campaign_id = ? AND product_id = ?", campaignID, productID).
		Delete(&entities.DiscountCampaignProduct{}).Error; err != nil {
		return fmt.Errorf("failed to remove product from campaign: %w", err)
	}
	return nil
}

// GetActiveCampaignsForProduct returns active campaigns for a product (tenant-scoped)
func (r *discountCampaignRepository) GetActiveCampaignsForProduct(ctx context.Context, productID uint) ([]entities.DiscountCampaign, error) {
	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant_id not found in context")
	}

	now := time.Now().UTC()
	var campaigns []entities.DiscountCampaign

	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND start_date <= ? AND end_date >= ?", tenantID, now, now).
		Where("id IN (SELECT campaign_id FROM discount_campaign_products WHERE product_id = ?)", productID).
		Find(&campaigns).Error; err != nil {
		return nil, fmt.Errorf("failed to get active campaigns for product: %w", err)
	}

	return campaigns, nil
}
