package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
)

type categoryService struct {
	categoryRepo interfaces.CategoryRepository
	logger       *slog.Logger
}

func NewCategoryService(categoryRepo interfaces.CategoryRepository, logger *slog.Logger) interfaces.CategoryService {
	return &categoryService{categoryRepo: categoryRepo, logger: logger}
}

func (s *categoryService) CreateCategory(ctx context.Context, category *entities.Category) error {
	if category.Name == "" {
		return fmt.Errorf("category name is required")
	}
	s.logger.InfoContext(ctx, "creating category", "name", category.Name)
	return s.categoryRepo.Create(ctx, category)
}

func (s *categoryService) GetCategory(ctx context.Context, id uint) (*entities.Category, error) {
	s.logger.InfoContext(ctx, "getting category", "id", id)
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}
	return category, nil
}

func (s *categoryService) ListCategories(ctx context.Context, page, limit int) ([]entities.Category, int64, error) {
	s.logger.InfoContext(ctx, "listing categories", "page", page, "limit", limit)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.categoryRepo.List(ctx, page, limit)
}

func (s *categoryService) UpdateCategory(ctx context.Context, id uint, name string) (*entities.Category, error) {
	if name == "" {
		return nil, fmt.Errorf("category name is required")
	}
	s.logger.InfoContext(ctx, "updating category", "id", id)

	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	category.Name = name
	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}
	return category, nil
}

func (s *categoryService) DeleteCategory(ctx context.Context, id uint) error {
	s.logger.InfoContext(ctx, "deleting category", "id", id)
	return s.categoryRepo.Delete(ctx, id)
}
