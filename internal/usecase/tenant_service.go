package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"github.com/usernamesalah/rh-pos/internal/pkg/storage"
)

type tenantService struct {
	tenantRepo interfaces.TenantRepository
	storage    storage.StorageClient
	baseURL    string
	logger     *slog.Logger
}

// NewTenantService creates a new tenant service
func NewTenantService(tenantRepo interfaces.TenantRepository, storageClient storage.StorageClient, baseURL string, logger *slog.Logger) interfaces.TenantService {
	return &tenantService{
		tenantRepo: tenantRepo,
		storage:    storageClient,
		baseURL:    baseURL,
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

// UploadTenantLogo uploads a logo for a tenant
func (s *tenantService) UploadTenantLogo(ctx context.Context, tenantID uint, fileData []byte, contentType string) (*entities.Tenant, error) {
	s.logger.InfoContext(ctx, "uploading tenant logo", "tenant_id", tenantID)

	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get tenant", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	ext := "png"
	switch contentType {
	case "image/jpeg":
		ext = "jpg"
	case "image/png":
		ext = "png"
	case "image/gif":
		ext = "gif"
	case "image/webp":
		ext = "webp"
	}

	key := storage.GenerateLogoKey(tenant.ID, ext)

	if err := s.storage.UploadBytes(ctx, key, fileData, contentType); err != nil {
		s.logger.ErrorContext(ctx, "failed to upload logo", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to upload logo: %w", err)
	}

	tenant.Logo = key
	if err := s.tenantRepo.Update(ctx, tenant); err != nil {
		s.logger.ErrorContext(ctx, "failed to update tenant with logo", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return tenant, nil
}

func (s *tenantService) GetTenantLogoURL(ctx context.Context, tenant *entities.Tenant) (string, error) {
	if tenant.Logo == "" {
		return "", nil
	}

	url, err := s.storage.GeneratePresignedURL(ctx, tenant.Logo, time.Hour, false)
	if err != nil {
		return "", fmt.Errorf("failed to generate logo URL: %w", err)
	}

	return url, nil
}
