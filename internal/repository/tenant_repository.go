package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"gorm.io/gorm"
)

type tenantRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *gorm.DB, logger *slog.Logger) interfaces.TenantRepository {
	return &tenantRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new tenant
func (r *tenantRepository) Create(ctx context.Context, tenant *entities.Tenant) error {
	r.logger.InfoContext(ctx, "creating tenant", "name", tenant.Name)
	if err := r.db.WithContext(ctx).Create(tenant).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to create tenant", "error", err)
		return fmt.Errorf("failed to create tenant: %w", err)
	}
	return nil
}

// GetByID retrieves a tenant by ID
func (r *tenantRepository) GetByID(ctx context.Context, id uint) (*entities.Tenant, error) {
	r.logger.InfoContext(ctx, "getting tenant by ID", "id", id)
	var tenant entities.Tenant
	if err := r.db.WithContext(ctx).First(&tenant, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tenant not found: %w", err)
		}
		r.logger.ErrorContext(ctx, "failed to get tenant", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	return &tenant, nil
}

// List retrieves all tenants
func (r *tenantRepository) List(ctx context.Context) ([]*entities.Tenant, error) {
	r.logger.InfoContext(ctx, "listing tenants")
	var tenants []*entities.Tenant
	if err := r.db.WithContext(ctx).Find(&tenants).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to list tenants", "error", err)
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	return tenants, nil
}

// Update updates a tenant
func (r *tenantRepository) Update(ctx context.Context, tenant *entities.Tenant) error {
	r.logger.InfoContext(ctx, "updating tenant", "id", tenant.ID)
	if err := r.db.WithContext(ctx).Save(tenant).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to update tenant", "error", err, "id", tenant.ID)
		return fmt.Errorf("failed to update tenant: %w", err)
	}
	return nil
}

// Delete deletes a tenant
func (r *tenantRepository) Delete(ctx context.Context, id uint) error {
	r.logger.InfoContext(ctx, "deleting tenant", "id", id)
	if err := r.db.WithContext(ctx).Delete(&entities.Tenant{}, id).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to delete tenant", "error", err, "id", id)
		return fmt.Errorf("failed to delete tenant: %w", err)
	}
	return nil
}
