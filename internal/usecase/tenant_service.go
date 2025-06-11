package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
)

type tenantService struct {
	tenantRepo interfaces.TenantRepository
	logger     *slog.Logger
}

// NewTenantService creates a new tenant service
func NewTenantService(tenantRepo interfaces.TenantRepository, logger *slog.Logger) interfaces.TenantService {
	return &tenantService{
		tenantRepo: tenantRepo,
		logger:     logger,
	}
}

// CreateTenant creates a new tenant
func (s *tenantService) CreateTenant(ctx context.Context, tenant *entities.Tenant) error {
	s.logger.InfoContext(ctx, "creating tenant", "name", tenant.Name)
	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		s.logger.ErrorContext(ctx, "failed to create tenant", "error", err)
		return fmt.Errorf("failed to create tenant: %w", err)
	}
	return nil
}

// GetTenantByID retrieves a tenant by ID
func (s *tenantService) GetTenantByID(ctx context.Context, id uint) (*entities.Tenant, error) {
	s.logger.InfoContext(ctx, "getting tenant by ID", "id", id)
	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get tenant", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	return tenant, nil
}

// ListTenants retrieves all tenants
func (s *tenantService) ListTenants(ctx context.Context) ([]*entities.Tenant, error) {
	s.logger.InfoContext(ctx, "listing tenants")
	tenants, err := s.tenantRepo.List(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to list tenants", "error", err)
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}
	return tenants, nil
}

// UpdateTenant updates a tenant
func (s *tenantService) UpdateTenant(ctx context.Context, tenant *entities.Tenant) error {
	s.logger.InfoContext(ctx, "updating tenant", "id", tenant.ID)
	if err := s.tenantRepo.Update(ctx, tenant); err != nil {
		s.logger.ErrorContext(ctx, "failed to update tenant", "error", err, "id", tenant.ID)
		return fmt.Errorf("failed to update tenant: %w", err)
	}
	return nil
}

// DeleteTenant deletes a tenant
func (s *tenantService) DeleteTenant(ctx context.Context, id uint) error {
	s.logger.InfoContext(ctx, "deleting tenant", "id", id)
	if err := s.tenantRepo.Delete(ctx, id); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete tenant", "error", err, "id", id)
		return fmt.Errorf("failed to delete tenant: %w", err)
	}
	return nil
}

// GetTenant retrieves a tenant by ID
func (s *tenantService) GetTenant(ctx context.Context, id uint) (*entities.Tenant, error) {
	return s.GetTenantByID(ctx, id)
}
