package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"github.com/usernamesalah/rh-pos/internal/pkg/ctxkey"
	"gorm.io/gorm"
)

type categoryRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewCategoryRepository(db *gorm.DB, logger *slog.Logger) interfaces.CategoryRepository {
	return &categoryRepository{db: db, logger: logger}
}

func (r *categoryRepository) Create(ctx context.Context, category *entities.Category) error {
	r.logger.InfoContext(ctx, "creating category", "name", category.Name)

	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant_id not found in context")
	}
	category.TenantID = &tenantID

	if err := r.db.WithContext(ctx).Create(category).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to create category", "error", err)
		return fmt.Errorf("failed to create category: %w", err)
	}
	return nil
}

func (r *categoryRepository) GetByID(ctx context.Context, id uint) (*entities.Category, error) {
	r.logger.InfoContext(ctx, "getting category", "id", id)

	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant_id not found in context")
	}

	var category entities.Category
	if err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&category).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to get category", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	return &category, nil
}

func (r *categoryRepository) List(ctx context.Context, page, limit int) ([]entities.Category, int64, error) {
	r.logger.InfoContext(ctx, "listing categories", "page", page, "limit", limit)

	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return nil, 0, fmt.Errorf("tenant_id not found in context")
	}

	var categories []entities.Category
	var total int64

	query := r.db.WithContext(ctx).Model(&entities.Category{}).Where("tenant_id = ?", tenantID)

	if err := query.Count(&total).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to count categories", "error", err)
		return nil, 0, fmt.Errorf("failed to count categories: %w", err)
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&categories).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to list categories", "error", err)
		return nil, 0, fmt.Errorf("failed to list categories: %w", err)
	}

	return categories, total, nil
}

func (r *categoryRepository) Update(ctx context.Context, category *entities.Category) error {
	r.logger.InfoContext(ctx, "updating category", "id", category.ID)

	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant_id not found in context")
	}

	if err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", category.ID, tenantID).
		Save(category).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to update category", "error", err, "id", category.ID)
		return fmt.Errorf("failed to update category: %w", err)
	}
	return nil
}

func (r *categoryRepository) Delete(ctx context.Context, id uint) error {
	r.logger.InfoContext(ctx, "deleting category", "id", id)

	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant_id not found in context")
	}

	if err := r.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Delete(&entities.Category{}).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to delete category", "error", err, "id", id)
		return fmt.Errorf("failed to delete category: %w", err)
	}
	return nil
}
