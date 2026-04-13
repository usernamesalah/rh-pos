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

// --- minimal mock repository ---

type mockProductRepo struct {
	products map[uint]*entities.Product
	nextID   uint
	bySKU    map[string]*entities.Product
}

func newMockRepo() *mockProductRepo {
	return &mockProductRepo{
		products: make(map[uint]*entities.Product),
		bySKU:    make(map[string]*entities.Product),
	}
}

func (m *mockProductRepo) Create(ctx context.Context, p *entities.Product) error {
	m.nextID++
	p.ID = m.nextID
	m.products[p.ID] = p
	if p.SKU != nil {
		m.bySKU[*p.SKU] = p
	}
	return nil
}

func (m *mockProductRepo) GetByID(ctx context.Context, id uint) (*entities.Product, error) {
	p, ok := m.products[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return p, nil
}

func (m *mockProductRepo) GetBySKU(ctx context.Context, sku string) (*entities.Product, error) {
	p, ok := m.bySKU[sku]
	if !ok {
		return nil, errors.New("not found")
	}
	return p, nil
}

func (m *mockProductRepo) List(ctx context.Context, page, limit int) ([]entities.Product, int64, error) {
	return nil, 0, nil
}

func (m *mockProductRepo) Update(ctx context.Context, p *entities.Product) error {
	m.products[p.ID] = p
	return nil
}

func (m *mockProductRepo) UpdateStock(ctx context.Context, id uint, stock int) error {
	return nil
}

func (m *mockProductRepo) Delete(ctx context.Context, id uint) error {
	delete(m.products, id)
	return nil
}

var _ interfaces.ProductRepository = (*mockProductRepo)(nil)

func ctxWithTenant(tenantID uint) context.Context {
	return ctxkey.WithTenantID(context.Background(), tenantID)
}

func newLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

// --- tests ---

func TestCreateProduct_NilSKU_DoesNotCheckDuplicate(t *testing.T) {
	repo := newMockRepo()
	svc := usecase.NewProductService(repo, nil, "", newLogger())
	ctx := ctxWithTenant(1)

	p := &entities.Product{Name: "Test", SKU: nil}
	if err := svc.CreateProduct(ctx, p); err != nil {
		t.Fatalf("expected nil error for nil SKU product, got: %v", err)
	}
}

func TestDeleteProduct_ExistingProduct_Succeeds(t *testing.T) {
	repo := newMockRepo()
	svc := usecase.NewProductService(repo, nil, "", newLogger())
	ctx := ctxWithTenant(1)

	p := &entities.Product{Name: "ToDelete"}
	if err := svc.CreateProduct(ctx, p); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	if err := svc.DeleteProduct(ctx, p.ID); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestCreateProduct_DuplicateSKU_ReturnsError(t *testing.T) {
	repo := newMockRepo()
	svc := usecase.NewProductService(repo, nil, "", newLogger())
	ctx := ctxWithTenant(1)

	sku := "SKU-001"
	p1 := &entities.Product{Name: "First", SKU: &sku}
	if err := svc.CreateProduct(ctx, p1); err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	p2 := &entities.Product{Name: "Second", SKU: &sku}
	if err := svc.CreateProduct(ctx, p2); err == nil {
		t.Fatal("expected duplicate SKU error, got nil")
	}
}
