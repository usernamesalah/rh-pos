# Feature Updates Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement five incremental product/tenant/RBAC updates: SKU nullable, tenant terms-of-service, cashier discount update permission, product delete endpoint, and nullable price fields with copy-on-null logic.

**Architecture:** Clean Architecture layers touched in order — entity → migration → repository (if needed) → usecase → handler → router. Each task is self-contained and the build must pass after every commit.

**Tech Stack:** Go 1.25, Echo v4, GORM v2, MySQL, Goose migrations, standard `testing` package for unit tests.

---

## File Map

| File | Change |
|------|--------|
| `migrations/015_sku_nullable.sql` | CREATE — make `products.sku` nullable |
| `migrations/016_add_tenant_terms_of_service.sql` | CREATE — add `tenants.terms_of_service` |
| `migrations/017_price_fields_nullable.sql` | CREATE — make `harga_modal`/`harga_jual` nullable |
| `internal/domain/entities/product.go` | MODIFY — `SKU *string`, `HargaModal/HargaJual *float64` |
| `internal/domain/entities/tenant.go` | MODIFY — add `TermsOfService string` |
| `internal/domain/interfaces/services.go` | MODIFY — add `DeleteProduct` to `ProductService` |
| `internal/usecase/product_service.go` | MODIFY — SKU nil guard, copy-on-null prices, `DeleteProduct` |
| `internal/usecase/product_service_test.go` | CREATE — unit tests for new business logic |
| `internal/handler/product_handler.go` | MODIFY — update request structs, add `DeleteProduct` handler |
| `internal/server/router.go` | MODIFY — remove AdminOnly from campaign PUT, add product DELETE |

---

## Task 1: Write the three migrations

**Files:**
- Create: `migrations/015_sku_nullable.sql`
- Create: `migrations/016_add_tenant_terms_of_service.sql`
- Create: `migrations/017_price_fields_nullable.sql`

- [ ] **Step 1: Create migration 015**

```sql
-- migrations/015_sku_nullable.sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE `products` MODIFY COLUMN `sku` VARCHAR(191) NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `products` MODIFY COLUMN `sku` VARCHAR(191) NOT NULL DEFAULT '';
-- +goose StatementEnd
```

- [ ] **Step 2: Create migration 016**

```sql
-- migrations/016_add_tenant_terms_of_service.sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE `tenants` ADD COLUMN `terms_of_service` TEXT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `tenants` DROP COLUMN `terms_of_service`;
-- +goose StatementEnd
```

- [ ] **Step 3: Create migration 017**

```sql
-- migrations/017_price_fields_nullable.sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE `products`
  MODIFY COLUMN `harga_modal` DECIMAL(15,2) NULL,
  MODIFY COLUMN `harga_jual`  DECIMAL(15,2) NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `products`
  MODIFY COLUMN `harga_modal` DECIMAL(15,2) NOT NULL DEFAULT 0,
  MODIFY COLUMN `harga_jual`  DECIMAL(15,2) NOT NULL DEFAULT 0;
-- +goose StatementEnd
```

- [ ] **Step 4: Verify build still passes**

```bash
go build ./...
```

Expected: no output (clean build).

- [ ] **Step 5: Commit**

```bash
git add migrations/015_sku_nullable.sql migrations/016_add_tenant_terms_of_service.sql migrations/017_price_fields_nullable.sql
git commit -m "feat: add migrations 015-017 for nullable SKU, tenant ToS, and nullable prices"
```

---

## Task 2: Tenant Terms of Service — entity + handler

**Files:**
- Modify: `internal/domain/entities/tenant.go`

- [ ] **Step 1: Add `TermsOfService` to Tenant entity**

In `internal/domain/entities/tenant.go`, add one field after `Logo`:

```go
// Tenant represents a tenant in the system
type Tenant struct {
	ID              uint      `json:"id"`
	Name            string    `json:"name"`
	About           string    `json:"about"`
	Address         string    `json:"address"`
	PhoneNumber     string    `json:"phone_number"`
	Logo            string    `json:"logo"`
	TermsOfService  string    `json:"terms_of_service"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
```

No handler change is needed: `admin.go` already binds `c.Bind(&tenant)` directly to `entities.Tenant`, so the new field is automatically included in create, update, and get responses.

- [ ] **Step 2: Verify build passes**

```bash
go build ./...
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add internal/domain/entities/tenant.go
git commit -m "feat: add terms_of_service field to Tenant entity"
```

---

## Task 3: Cashier can update discount campaigns — router change

**Files:**
- Modify: `internal/server/router.go`

- [ ] **Step 1: Remove AdminOnly from campaign PUT route**

In `internal/server/router.go`, find the campaign routes block (around line 165) and change:

```go
campaigns.PUT("/:id", campaignHandler.UpdateCampaign, adminMiddleware.AdminOnly)
```

to:

```go
campaigns.PUT("/:id", campaignHandler.UpdateCampaign)
```

All other campaign routes remain unchanged (POST, DELETE, POST products, DELETE products all keep `adminMiddleware.AdminOnly`).

- [ ] **Step 2: Verify build passes**

```bash
go build ./...
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add internal/server/router.go
git commit -m "feat: allow cashier role to update discount campaigns"
```

---

## Task 4: SKU nullable — entity, usecase, handler

**Files:**
- Modify: `internal/domain/entities/product.go`
- Modify: `internal/usecase/product_service.go`
- Modify: `internal/handler/product_handler.go`

- [ ] **Step 1: Write failing test for SKU nil in CreateProduct**

Create `internal/usecase/product_service_test.go`:

```go
package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"github.com/usernamesalah/rh-pos/internal/pkg/ctxkey"
	"github.com/usernamesalah/rh-pos/internal/usecase"
	"log/slog"
	"os"
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/usecase/... -run TestCreateProduct -v 2>&1 | head -30
```

Expected: compilation error because `Product.SKU` is still `string`, not `*string`.

- [ ] **Step 3: Update Product entity**

In `internal/domain/entities/product.go`, change:

```go
SKU        string    `json:"sku" gorm:"uniqueIndex:idx_tenant_sku;not null"`
```

to:

```go
SKU        *string   `json:"sku" gorm:"uniqueIndex:idx_tenant_sku"`
```

- [ ] **Step 4: Fix product_handler.go request struct**

In `internal/handler/product_handler.go`, change `CreateProductRequest`:

```go
type CreateProductRequest struct {
	Name       string   `json:"name" validate:"required"`
	SKU        *string  `json:"sku,omitempty"`
	Image      string   `json:"image,omitempty"`
	HargaModal *float64 `json:"harga_modal,omitempty"`
	HargaJual  *float64 `json:"harga_jual,omitempty"`
	Stock      int      `json:"stock" validate:"min=0"`
}
```

Then in `CreateProduct` handler, update the product assignment:

```go
product := &entities.Product{
    Name:       req.Name,
    SKU:        req.SKU,
    Image:      req.Image,
    HargaModal: req.HargaModal,
    HargaJual:  req.HargaJual,
    Stock:      req.Stock,
}
```

- [ ] **Step 5: Fix product_service.go for nil SKU**

In `internal/usecase/product_service.go`, replace the `CreateProduct` method with:

```go
// CreateProduct creates a new product
func (s *productService) CreateProduct(ctx context.Context, product *entities.Product) error {
	var skuStr string
	if product.SKU != nil {
		skuStr = *product.SKU
	}
	s.logger.InfoContext(ctx, "creating product", "sku", skuStr)

	// Get tenant_id from context
	tenantID, ok := ctxkey.TenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant_id not found in context")
	}

	// Set tenant_id
	product.TenantID = &tenantID

	// Check for duplicate SKU only when SKU is provided
	if product.SKU != nil {
		existingProduct, err := s.productRepo.GetBySKU(ctx, *product.SKU)
		if err == nil && existingProduct != nil {
			return fmt.Errorf("product with SKU %s already exists", *product.SKU)
		}
	}

	// Create product
	if err := s.productRepo.Create(ctx, product); err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}
```

Also fix the `UpdateProduct` switch in the same file — `sku` case needs to handle `*string` entity field:

```go
case "sku":
    if v, ok := value.(string); ok {
        product.SKU = &v
    }
case "harga_modal":
    product.HargaModal = value.(float64)
case "harga_jual":
    product.HargaJual = value.(float64)
```

Wait — `HargaModal` and `HargaJual` are still `float64` at this point in the plan (they become `*float64` in Task 6). Leave the `harga_modal`/`harga_jual` cases as `float64` for now; Task 6 will update them.

- [ ] **Step 6: Run the tests**

```bash
go test ./internal/usecase/... -run TestCreateProduct -v
```

Expected output:
```
--- PASS: TestCreateProduct_NilSKU_DoesNotCheckDuplicate (0.00s)
--- PASS: TestCreateProduct_DuplicateSKU_ReturnsError (0.00s)
PASS
```

- [ ] **Step 7: Verify full build**

```bash
go build ./...
```

Expected: no output.

- [ ] **Step 8: Commit**

```bash
git add internal/domain/entities/product.go \
        internal/usecase/product_service.go \
        internal/usecase/product_service_test.go \
        internal/handler/product_handler.go
git commit -m "feat: make product SKU optional (nullable)"
```

---

## Task 5: Product delete — service interface, usecase, handler, router

**Files:**
- Modify: `internal/domain/interfaces/services.go`
- Modify: `internal/usecase/product_service.go`
- Modify: `internal/handler/product_handler.go`
- Modify: `internal/server/router.go`

Note: `ProductRepository.Delete` already exists in `internal/domain/interfaces/repositories.go` and is implemented in `internal/repository/product_repository.go`. Only the service layer and above need changes.

- [ ] **Step 1: Write failing test for DeleteProduct**

Add to `internal/usecase/product_service_test.go`:

```go
func TestDeleteProduct_ExistingProduct_Succeeds(t *testing.T) {
	repo := newMockRepo()
	svc := usecase.NewProductService(repo, nil, "", newLogger())
	ctx := ctxWithTenant(1)

	// Create a product first
	p := &entities.Product{Name: "ToDelete"}
	if err := svc.CreateProduct(ctx, p); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	if err := svc.DeleteProduct(ctx, p.ID); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/usecase/... -run TestDeleteProduct -v 2>&1 | head -10
```

Expected: compilation error — `DeleteProduct` does not exist on service.

- [ ] **Step 3: Add DeleteProduct to ProductService interface**

In `internal/domain/interfaces/services.go`, add one line to the `ProductService` interface:

```go
type ProductService interface {
    GetProduct(ctx context.Context, id uint) (*entities.Product, error)
    ListProducts(ctx context.Context, page, limit int) ([]entities.Product, int64, error)
    UpdateProduct(ctx context.Context, id uint, updates map[string]interface{}) (*entities.Product, error)
    UpdateStock(ctx context.Context, id uint, stock int) (*entities.Product, error)
    CreateProduct(ctx context.Context, product *entities.Product) error
    DeleteProduct(ctx context.Context, id uint) error
    GetProductImageURL(ctx context.Context, product *entities.Product) (string, error)
    GetProductUploadURL(ctx context.Context, product *entities.Product, ext string) (uploadURL string, imageKey string, err error)
    UploadProductImage(ctx context.Context, productID uint, fileData []byte, contentType string) (*entities.Product, error)
    GetProductImageBytes(ctx context.Context, productID uint) ([]byte, string, error)
}
```

- [ ] **Step 4: Implement DeleteProduct in productService usecase**

Add at the end of `internal/usecase/product_service.go`:

```go
// DeleteProduct deletes a product by ID
func (s *productService) DeleteProduct(ctx context.Context, id uint) error {
	s.logger.InfoContext(ctx, "deleting product", "id", id)

	if err := s.productRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}
```

Also add `Delete` to the mock repository in the test file (it's already there from Task 4 Step 1 — no additional change needed).

- [ ] **Step 5: Run the test**

```bash
go test ./internal/usecase/... -run TestDeleteProduct -v
```

Expected:
```
--- PASS: TestDeleteProduct_ExistingProduct_Succeeds (0.00s)
PASS
```

- [ ] **Step 6: Add DeleteProduct handler**

Add to the end of `internal/handler/product_handler.go`:

```go
// DeleteProduct handles deleting a product
// @Summary Delete a product
// @Description Delete a product by its hashed ID (admin only)
// @Tags Products
// @Produce json
// @Security bearerAuth
// @Param id path string true "Hashed product ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /api/products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := hash.DecodeHashID(c.Param("id"))
	if err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid product ID format")
	}

	if err := h.productService.DeleteProduct(ctx, id); err != nil {
		h.logger.ErrorContext(ctx, "failed to delete product", "error", err, "id", id)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to delete product")
	}

	return SuccessResponse(c, http.StatusOK, "Product deleted successfully", nil)
}
```

- [ ] **Step 7: Register the route in the router**

In `internal/server/router.go`, add after the existing product routes (around line 148):

```go
products.DELETE("/:id", productHandler.DeleteProduct, adminMiddleware.AdminOnly)
```

The full products block should now look like:

```go
products := api.Group("/products")
products.GET("", productHandler.ListProducts)
products.GET("/:id", productHandler.GetProduct)
products.GET("/:id/image/bytes", productHandler.GetProductImageBytes)
products.POST("", productHandler.CreateProduct, adminMiddleware.AdminOnly)
products.PUT("/:id", productHandler.UpdateProduct, adminMiddleware.AdminOnly)
products.PUT("/:id/stock", productHandler.UpdateStock, adminMiddleware.AdminOnly)
products.DELETE("/:id", productHandler.DeleteProduct, adminMiddleware.AdminOnly)
products.POST("/:id/upload-url", productHandler.GetUploadURL, adminMiddleware.AdminOnly)
products.POST("/:id/image", productHandler.UploadProductImage, adminMiddleware.AdminOnly)
```

- [ ] **Step 8: Verify build passes**

```bash
go build ./...
go test ./internal/usecase/... -v
```

Expected: build clean, all tests pass.

- [ ] **Step 9: Commit**

```bash
git add internal/domain/interfaces/services.go \
        internal/usecase/product_service.go \
        internal/usecase/product_service_test.go \
        internal/handler/product_handler.go \
        internal/server/router.go
git commit -m "feat: add admin-only DELETE /api/products/:id endpoint"
```

---

## Task 6: Nullable price fields with copy-on-null logic

**Files:**
- Modify: `internal/domain/entities/product.go`
- Modify: `internal/usecase/product_service.go`
- Modify: `internal/handler/product_handler.go`

- [ ] **Step 1: Write failing tests for copy-on-null**

Add to `internal/usecase/product_service_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/usecase/... -run TestCreateProduct_Only -v 2>&1 | head -20
```

Expected: compilation errors — `HargaModal` and `HargaJual` are still `float64`, not `*float64`.

- [ ] **Step 3: Update Product entity**

In `internal/domain/entities/product.go`, change:

```go
HargaModal float64   `json:"harga_modal" gorm:"not null"`
HargaJual  float64   `json:"harga_jual" gorm:"not null"`
```

to:

```go
HargaModal *float64  `json:"harga_modal"`
HargaJual  *float64  `json:"harga_jual"`
```

- [ ] **Step 4: Fix UpdateProduct switch in usecase**

In `internal/usecase/product_service.go`, update the `harga_modal` and `harga_jual` cases in the `UpdateProduct` switch:

```go
case "harga_modal":
    if v, ok := value.(float64); ok {
        product.HargaModal = &v
    }
case "harga_jual":
    if v, ok := value.(float64); ok {
        product.HargaJual = &v
    }
```

- [ ] **Step 5: Add copy-on-null logic to CreateProduct in usecase**

In `internal/usecase/product_service.go`, add price normalization inside `CreateProduct`, before the `productRepo.Create` call:

```go
// Normalize prices: if one is nil, copy the other; both nil is invalid
switch {
case product.HargaJual != nil && product.HargaModal == nil:
    v := *product.HargaJual
    product.HargaModal = &v
case product.HargaModal != nil && product.HargaJual == nil:
    v := *product.HargaModal
    product.HargaJual = &v
case product.HargaJual == nil && product.HargaModal == nil:
    return fmt.Errorf("at least one of harga_jual or harga_modal must be provided")
}
```

- [ ] **Step 6: Fix handler compilation errors from *float64 entity fields**

The `ListProducts`, `GetProduct`, `UpdateProduct`, `UpdateStock`, `UploadProductImage` handlers all have response maps like:
```go
"harga_modal": product.HargaModal,
"harga_jual":  product.HargaJual,
```

With `*float64` fields, these will now serialize as `null` when nil and as the number when set — that is correct JSON behavior. No handler changes needed for the response maps; Go's JSON marshaller handles `*float64` correctly.

However, the `UpdateProductRequest` already uses `*float64` — no change there.

The `CreateProductRequest` was already updated in Task 4 Step 4 to use `*float64`. No further change needed.

- [ ] **Step 7: Run all tests**

```bash
go test ./internal/usecase/... -v
```

Expected output (all pass):
```
--- PASS: TestCreateProduct_NilSKU_DoesNotCheckDuplicate
--- PASS: TestCreateProduct_DuplicateSKU_ReturnsError
--- PASS: TestDeleteProduct_ExistingProduct_Succeeds
--- PASS: TestCreateProduct_OnlyHargaJual_CopiesHargaModal
--- PASS: TestCreateProduct_OnlyHargaModal_CopiesHargaJual
--- PASS: TestCreateProduct_BothPricesNil_ReturnsError
--- PASS: TestCreateProduct_BothPricesSet_UsesAsIs
PASS
```

- [ ] **Step 8: Verify full build**

```bash
go build ./...
go test ./... 2>&1
```

Expected: clean build, all tests pass.

- [ ] **Step 9: Commit**

```bash
git add internal/domain/entities/product.go \
        internal/usecase/product_service.go \
        internal/usecase/product_service_test.go \
        internal/handler/product_handler.go
git commit -m "feat: make harga_jual and harga_modal optional with copy-on-null logic"
```

---

## Spec Coverage Check

| Spec requirement | Task |
|-----------------|------|
| SKU nullable, no duplicate check when nil | Task 4 |
| Migration 015 | Task 1 |
| Tenant terms_of_service field | Task 2 |
| Migration 016 | Task 1 |
| Cashier can PUT discount campaign | Task 3 |
| Product delete endpoint (admin-only) | Task 5 |
| Migration 017 | Task 1 |
| harga_jual/harga_modal nullable with copy-on-null | Task 6 |
| Both prices nil → validation error | Task 6 |
| Cashier stock visibility (already works) | no task needed |
