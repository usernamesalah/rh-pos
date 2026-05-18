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

func (m *mockProductRepo) List(ctx context.Context, page, limit int, search string, categoryID *uint) ([]entities.Product, int64, error) {
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

	p := &entities.Product{Name: "Test", SKU: nil, HargaJual: ptr64(1000)}
	if err := svc.CreateProduct(ctx, p); err != nil {
		t.Fatalf("expected nil error for nil SKU product, got: %v", err)
	}
}

func TestDeleteProduct_ExistingProduct_Succeeds(t *testing.T) {
	repo := newMockRepo()
	svc := usecase.NewProductService(repo, nil, "", newLogger())
	ctx := ctxWithTenant(1)

	p := &entities.Product{Name: "ToDelete", HargaJual: ptr64(500)}
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
	p1 := &entities.Product{Name: "First", SKU: &sku, HargaJual: ptr64(1000)}
	if err := svc.CreateProduct(ctx, p1); err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	p2 := &entities.Product{Name: "Second", SKU: &sku, HargaJual: ptr64(2000)}
	if err := svc.CreateProduct(ctx, p2); err == nil {
		t.Fatal("expected duplicate SKU error, got nil")
	}
}

func ptr64(v float64) *float64 { return &v }

func TestCreateProduct_OnlyHargaJual_CopiesHargaModal(t *testing.T) {
	repo := newMockRepo()
	svc := usecase.NewProductService(repo, nil, "", newLogger())
	ctx := ctxWithTenant(1)

	p := &entities.Product{Name: "Test", HargaJual: ptr64(2000)}
	if err := svc.CreateProduct(ctx, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.HargaModal == nil {
		t.Fatal("expected HargaModal to be set, got nil")
	}
	if *p.HargaModal != 2000 {
		t.Errorf("expected HargaModal=2000, got %v", *p.HargaModal)
	}
}

func TestCreateProduct_OnlyHargaModal_CopiesHargaJual(t *testing.T) {
	repo := newMockRepo()
	svc := usecase.NewProductService(repo, nil, "", newLogger())
	ctx := ctxWithTenant(1)

	p := &entities.Product{Name: "Test", HargaModal: ptr64(1500)}
	if err := svc.CreateProduct(ctx, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.HargaJual == nil {
		t.Fatal("expected HargaJual to be set, got nil")
	}
	if *p.HargaJual != 1500 {
		t.Errorf("expected HargaJual=1500, got %v", *p.HargaJual)
	}
}

func TestCreateProduct_BothPricesNil_ReturnsError(t *testing.T) {
	repo := newMockRepo()
	svc := usecase.NewProductService(repo, nil, "", newLogger())
	ctx := ctxWithTenant(1)

	p := &entities.Product{Name: "Test"}
	err := svc.CreateProduct(ctx, p)
	if err == nil {
		t.Fatal("expected validation error for both nil prices, got nil")
	}
}

func TestCreateProduct_BothPricesSet_UsesAsIs(t *testing.T) {
	repo := newMockRepo()
	svc := usecase.NewProductService(repo, nil, "", newLogger())
	ctx := ctxWithTenant(1)

	p := &entities.Product{Name: "Test", HargaJual: ptr64(3000), HargaModal: ptr64(2000)}
	if err := svc.CreateProduct(ctx, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *p.HargaJual != 3000 || *p.HargaModal != 2000 {
		t.Errorf("expected prices unchanged, got jual=%v modal=%v", *p.HargaJual, *p.HargaModal)
	}
}
