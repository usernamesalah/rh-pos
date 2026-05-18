package usecase_test

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"github.com/usernamesalah/rh-pos/internal/pkg/ctxkey"
	"github.com/usernamesalah/rh-pos/internal/usecase"
)

type mockCategoryRepo struct {
	categories map[uint]*entities.Category
	nextID     uint
}

func newMockCategoryRepo() *mockCategoryRepo {
	return &mockCategoryRepo{categories: make(map[uint]*entities.Category)}
}

func (m *mockCategoryRepo) Create(ctx context.Context, c *entities.Category) error {
	m.nextID++
	c.ID = m.nextID
	cp := *c
	m.categories[c.ID] = &cp
	return nil
}

func (m *mockCategoryRepo) GetByID(ctx context.Context, id uint) (*entities.Category, error) {
	c, ok := m.categories[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return c, nil
}

func (m *mockCategoryRepo) List(ctx context.Context, page, limit int) ([]entities.Category, int64, error) {
	all := make([]entities.Category, 0, len(m.categories))
	for _, c := range m.categories {
		all = append(all, *c)
	}
	return all, int64(len(all)), nil
}

func (m *mockCategoryRepo) Update(ctx context.Context, c *entities.Category) error {
	if _, ok := m.categories[c.ID]; !ok {
		return errors.New("not found")
	}
	cp := *c
	m.categories[c.ID] = &cp
	return nil
}

func (m *mockCategoryRepo) Delete(ctx context.Context, id uint) error {
	if _, ok := m.categories[id]; !ok {
		return errors.New("not found")
	}
	delete(m.categories, id)
	return nil
}

var _ interfaces.CategoryRepository = (*mockCategoryRepo)(nil)

func ctxWithTenantCat(tenantID uint) context.Context {
	return ctxkey.WithTenantID(context.Background(), tenantID)
}

func newTestCategoryLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, nil))
}

func TestCreateCategory_Success(t *testing.T) {
	svc := usecase.NewCategoryService(newMockCategoryRepo(), newTestCategoryLogger())
	ctx := ctxWithTenantCat(1)

	cat := &entities.Category{Name: "Electronics"}
	if err := svc.CreateCategory(ctx, cat); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cat.ID == 0 {
		t.Fatal("expected ID to be set after create")
	}
}

func TestCreateCategory_EmptyName_ReturnsError(t *testing.T) {
	svc := usecase.NewCategoryService(newMockCategoryRepo(), newTestCategoryLogger())
	ctx := ctxWithTenantCat(1)

	if err := svc.CreateCategory(ctx, &entities.Category{Name: ""}); err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

func TestGetCategory_NotFound_ReturnsError(t *testing.T) {
	svc := usecase.NewCategoryService(newMockCategoryRepo(), newTestCategoryLogger())
	ctx := ctxWithTenantCat(1)

	if _, err := svc.GetCategory(ctx, 999); err == nil {
		t.Fatal("expected error for non-existent category, got nil")
	}
}

func TestUpdateCategory_Success(t *testing.T) {
	svc := usecase.NewCategoryService(newMockCategoryRepo(), newTestCategoryLogger())
	ctx := ctxWithTenantCat(1)

	cat := &entities.Category{Name: "Electronics"}
	_ = svc.CreateCategory(ctx, cat)

	updated, err := svc.UpdateCategory(ctx, cat.ID, "Gadgets")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "Gadgets" {
		t.Fatalf("expected name 'Gadgets', got %q", updated.Name)
	}
}

func TestUpdateCategory_EmptyName_ReturnsError(t *testing.T) {
	svc := usecase.NewCategoryService(newMockCategoryRepo(), newTestCategoryLogger())
	ctx := ctxWithTenantCat(1)

	cat := &entities.Category{Name: "Electronics"}
	_ = svc.CreateCategory(ctx, cat)

	if _, err := svc.UpdateCategory(ctx, cat.ID, ""); err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

func TestDeleteCategory_Success(t *testing.T) {
	svc := usecase.NewCategoryService(newMockCategoryRepo(), newTestCategoryLogger())
	ctx := ctxWithTenantCat(1)

	cat := &entities.Category{Name: "Electronics"}
	_ = svc.CreateCategory(ctx, cat)

	if err := svc.DeleteCategory(ctx, cat.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := svc.GetCategory(ctx, cat.ID); err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestListCategories_ReturnsTenantCategories(t *testing.T) {
	svc := usecase.NewCategoryService(newMockCategoryRepo(), newTestCategoryLogger())
	ctx := ctxWithTenantCat(1)

	_ = svc.CreateCategory(ctx, &entities.Category{Name: "A"})
	_ = svc.CreateCategory(ctx, &entities.Category{Name: "B"})

	cats, total, err := svc.ListCategories(ctx, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total=2, got %d", total)
	}
	if len(cats) != 2 {
		t.Fatalf("expected 2 items, got %d", len(cats))
	}
}
