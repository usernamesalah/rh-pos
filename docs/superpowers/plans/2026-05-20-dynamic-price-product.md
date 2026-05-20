# Dynamic Price Product Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `is_dynamic_price` bool flag to Product so cashiers can enter price freely at checkout; skip stock deduction and campaign discounts for these items.

**Architecture:** Boolean flag on the Product entity drives branching in `CreateTransaction` — when true, item price comes from the request (default 0.0), stock is not deducted, and no campaign discount is applied. A pure helper function `ResolveItemPrice` isolates the logic for testability.

**Tech Stack:** Go, Echo, GORM, MySQL, goose migrations

---

### Task 1: Add migration file

**Files:**
- Create: `migrations/019_add_is_dynamic_price_to_products.sql`

- [ ] **Step 1: Create migration file**

```sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE `products`
    ADD COLUMN `is_dynamic_price` TINYINT(1) NOT NULL DEFAULT 0 AFTER `stock`;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `products`
    DROP COLUMN `is_dynamic_price`;
-- +goose StatementEnd
```

- [ ] **Step 2: Verify file exists**

```bash
cat migrations/019_add_is_dynamic_price_to_products.sql
```

Expected: file contents shown correctly.

- [ ] **Step 3: Commit**

```bash
git add migrations/019_add_is_dynamic_price_to_products.sql
git commit -m "feat: add is_dynamic_price migration for products"
```

---

### Task 2: Add IsDynamicPrice field to Product entity

**Files:**
- Modify: `internal/domain/entities/product.go`

- [ ] **Step 1: Add field to Product struct**

Replace the full `Product` struct in `internal/domain/entities/product.go`:

```go
type Product struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	Image          string    `json:"image"`
	Name           string    `json:"name" gorm:"not null"`
	SKU            *string   `json:"sku" gorm:"uniqueIndex:idx_tenant_sku"`
	HargaModal     *float64  `json:"harga_modal"`
	HargaJual      *float64  `json:"harga_jual"`
	Stock          int       `json:"stock" gorm:"not null;default:0"`
	IsDynamicPrice bool      `json:"is_dynamic_price" gorm:"default:false"`
	TenantID       *uint     `json:"tenant_id" gorm:"uniqueIndex:idx_tenant_sku;index"`
	Tenant         *Tenant   `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
	CategoryID     *uint     `json:"category_id" gorm:"index"`
	Category       *Category `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
```

- [ ] **Step 2: Build to verify no compile errors**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add internal/domain/entities/product.go
git commit -m "feat: add IsDynamicPrice field to Product entity"
```

---

### Task 3: Add Price field to TransactionItemRequest in service interface

**Files:**
- Modify: `internal/domain/interfaces/services.go`

- [ ] **Step 1: Add Price field to TransactionItemRequest**

Replace the existing `TransactionItemRequest` struct (lines 101–106) with:

```go
// TransactionItemRequest represents an item in transaction request
type TransactionItemRequest struct {
	ProductID    uint     `json:"product_id"`
	Quantity     int      `json:"quantity"`
	WarrantyDays int      `json:"warranty_days"`
	Price        *float64 `json:"price"` // used only when product.IsDynamicPrice = true; defaults to 0.0 if nil
}
```

- [ ] **Step 2: Build to verify no compile errors**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add internal/domain/interfaces/services.go
git commit -m "feat: add Price field to TransactionItemRequest interface"
```

---

### Task 4: Implement dynamic price logic in transaction service (TDD)

**Files:**
- Modify: `internal/usecase/transaction_service.go`
- Create: `internal/usecase/transaction_service_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/usecase/transaction_service_test.go`:

```go
package usecase_test

import (
	"testing"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/usecase"
)

func TestResolveItemPrice_RegularProduct_UsesHargaJual(t *testing.T) {
	price := 50000.0
	product := &entities.Product{HargaJual: &price, IsDynamicPrice: false}
	got := usecase.ResolveItemPrice(product, nil)
	if got != 50000.0 {
		t.Fatalf("expected 50000.0, got %v", got)
	}
}

func TestResolveItemPrice_RegularProduct_NilHargaJual_ReturnsZero(t *testing.T) {
	product := &entities.Product{HargaJual: nil, IsDynamicPrice: false}
	got := usecase.ResolveItemPrice(product, nil)
	if got != 0.0 {
		t.Fatalf("expected 0.0, got %v", got)
	}
}

func TestResolveItemPrice_DynamicProduct_UsesRequestPrice(t *testing.T) {
	product := &entities.Product{IsDynamicPrice: true}
	reqPrice := 75000.0
	got := usecase.ResolveItemPrice(product, &reqPrice)
	if got != 75000.0 {
		t.Fatalf("expected 75000.0, got %v", got)
	}
}

func TestResolveItemPrice_DynamicProduct_NilPrice_ReturnsZero(t *testing.T) {
	product := &entities.Product{IsDynamicPrice: true}
	got := usecase.ResolveItemPrice(product, nil)
	if got != 0.0 {
		t.Fatalf("expected 0.0, got %v", got)
	}
}

func TestResolveItemPrice_DynamicProduct_IgnoresHargaJual(t *testing.T) {
	storedPrice := 99999.0
	product := &entities.Product{HargaJual: &storedPrice, IsDynamicPrice: true}
	reqPrice := 25000.0
	got := usecase.ResolveItemPrice(product, &reqPrice)
	if got != 25000.0 {
		t.Fatalf("expected request price 25000.0, got %v", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/usecase/... -run TestResolveItemPrice -v
```

Expected: FAIL — `usecase.ResolveItemPrice undefined`

- [ ] **Step 3: Add ResolveItemPrice helper to transaction_service.go**

Add this exported helper function at the bottom of `internal/usecase/transaction_service.go`:

```go
// ResolveItemPrice returns the price to use for a transaction item.
// For dynamic price products, uses requestPrice (default 0.0 if nil).
// For regular products, uses HargaJual (default 0.0 if nil).
func ResolveItemPrice(product *entities.Product, requestPrice *float64) float64 {
	if product.IsDynamicPrice {
		if requestPrice != nil {
			return *requestPrice
		}
		return 0.0
	}
	if product.HargaJual != nil {
		return *product.HargaJual
	}
	return 0.0
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/usecase/... -run TestResolveItemPrice -v
```

Expected: all 5 tests PASS.

- [ ] **Step 5: Update CreateTransaction item loop to use ResolveItemPrice and branch on IsDynamicPrice**

In `internal/usecase/transaction_service.go`, replace the item processing loop body (the `for _, item := range req.Items` block, lines 71–125) with:

```go
		for _, item := range req.Items {
			product, err := s.productRepo.GetByID(ctx, item.ProductID)
			if err != nil {
				return fmt.Errorf("product not found: %w", err)
			}

			itemPrice := ResolveItemPrice(product, item.Price)

			transactionItem := entities.TransactionItem{
				ProductID:    item.ProductID,
				Quantity:     item.Quantity,
				WarrantyDays: item.WarrantyDays,
				Price:        itemPrice,
			}

			if product.IsDynamicPrice {
				// Dynamic price product: no campaign discount, no stock deduction
				regularTotal += itemPrice * float64(item.Quantity)
			} else {
				// Regular product: check campaigns, deduct stock
				campaigns, err := s.campaignRepo.GetActiveCampaignsForProduct(ctx, item.ProductID)
				if err != nil {
					return fmt.Errorf("failed to check campaigns for product: %w", err)
				}

				if len(campaigns) > 0 {
					best := campaigns[0]
					for _, c := range campaigns[1:] {
						if c.DiscountPercentage > best.DiscountPercentage {
							best = c
						}
					}
					discountedPrice := itemPrice * (1 - best.DiscountPercentage/100)
					transactionItem.Price = discountedPrice
					transactionItem.DiscountPercentage = best.DiscountPercentage
					transactionItem.CampaignID = &best.ID
					campaignDiscountedTotal += discountedPrice * float64(item.Quantity)
				} else {
					regularTotal += itemPrice * float64(item.Quantity)
				}

				result := tx.Model(&entities.Product{}).
					Where("id = ? AND tenant_id = ? AND stock >= ?", item.ProductID, tenantID, item.Quantity).
					Update("stock", gorm.Expr("stock - ?", item.Quantity))
				if result.Error != nil {
					return fmt.Errorf("failed to update product stock: %w", result.Error)
				}
				if result.RowsAffected == 0 {
					return fmt.Errorf("insufficient stock for product %s", product.Name)
				}
			}

			transaction.Items = append(transaction.Items, transactionItem)
		}
```

- [ ] **Step 6: Build and run all tests**

```bash
go build ./... && go test ./internal/usecase/... -v
```

Expected: build succeeds, all tests pass.

- [ ] **Step 7: Commit**

```bash
git add internal/usecase/transaction_service.go internal/usecase/transaction_service_test.go
git commit -m "feat: implement dynamic price logic in transaction service"
```

---

### Task 5: Update product handler DTOs and responses

**Files:**
- Modify: `internal/handler/product_handler.go`

- [ ] **Step 1: Add IsDynamicPrice to CreateProductRequest**

Replace the `CreateProductRequest` struct (lines 47–55):

```go
// CreateProductRequest represents the create product request
type CreateProductRequest struct {
	Name           string   `json:"name" validate:"required"`
	SKU            *string  `json:"sku,omitempty"`
	Image          string   `json:"image,omitempty"`
	HargaModal     *float64 `json:"harga_modal,omitempty"`
	HargaJual      *float64 `json:"harga_jual,omitempty"`
	Stock          int      `json:"stock" validate:"min=0"`
	CategoryID     *string  `json:"category_id,omitempty"`
	IsDynamicPrice bool     `json:"is_dynamic_price"`
}
```

- [ ] **Step 2: Add IsDynamicPrice to UpdateProductRequest**

Replace the `UpdateProductRequest` struct (lines 33–39):

```go
// UpdateProductRequest represents the update product request
type UpdateProductRequest struct {
	Name           *string  `json:"name,omitempty"`
	SKU            *string  `json:"sku,omitempty"`
	HargaModal     *float64 `json:"harga_modal,omitempty"`
	HargaJual      *float64 `json:"harga_jual,omitempty"`
	CategoryID     *string  `json:"category_id,omitempty"`
	IsDynamicPrice *bool    `json:"is_dynamic_price,omitempty"`
}
```

- [ ] **Step 3: Pass IsDynamicPrice in CreateProduct handler body**

In `CreateProduct` handler (around line 395), update the product construction to include `IsDynamicPrice`:

```go
	product := &entities.Product{
		Name:           req.Name,
		SKU:            req.SKU,
		Image:          req.Image,
		HargaModal:     req.HargaModal,
		HargaJual:      req.HargaJual,
		Stock:          req.Stock,
		CategoryID:     categoryID,
		IsDynamicPrice: req.IsDynamicPrice,
	}
```

- [ ] **Step 4: Add IsDynamicPrice to create response map**

In `CreateProduct` handler (around line 419), update the response map:

```go
	response := WithHashID(
		product.ID,
		product.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		product.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		map[string]interface{}{
			"name":             product.Name,
			"sku":              product.SKU,
			"image":            product.Image,
			"harga_modal":      product.HargaModal,
			"harga_jual":       product.HargaJual,
			"stock":            product.Stock,
			"category_id":      categoryHashedID(product.CategoryID),
			"is_dynamic_price": product.IsDynamicPrice,
		},
	)
```

- [ ] **Step 5: Add IsDynamicPrice to UpdateProduct updates map**

In `UpdateProduct` handler (around line 254, after the CategoryID block), add:

```go
	if req.IsDynamicPrice != nil {
		updates["is_dynamic_price"] = *req.IsDynamicPrice
	}
```

- [ ] **Step 6: Add IsDynamicPrice to update response map**

In `UpdateProduct` handler response (around line 276):

```go
	response := WithHashID(
		product.ID,
		product.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		product.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		map[string]interface{}{
			"name":             product.Name,
			"sku":              product.SKU,
			"image_url":        imageURL,
			"harga_modal":      product.HargaModal,
			"harga_jual":       product.HargaJual,
			"stock":            product.Stock,
			"category_id":      categoryHashedID(product.CategoryID),
			"is_dynamic_price": product.IsDynamicPrice,
		},
	)
```

- [ ] **Step 7: Build to verify**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 8: Commit**

```bash
git add internal/handler/product_handler.go
git commit -m "feat: add is_dynamic_price to product create/update handler"
```

---

### Task 6: Update transaction handler DTO

**Files:**
- Modify: `internal/handler/transaction_handler.go`

- [ ] **Step 1: Add Price to TransactionItemRequest in handler**

Replace `TransactionItemRequest` struct (lines 43–47):

```go
// TransactionItemRequest represents an item in transaction request
type TransactionItemRequest struct {
	ProductID    string   `json:"product_id" validate:"required"`
	Quantity     int      `json:"quantity" validate:"required,min=1"`
	WarrantyDays int      `json:"warranty_days"`
	Price        *float64 `json:"price"`
}
```

- [ ] **Step 2: Forward Price when building service request**

In `CreateTransaction` handler (around lines 102–106), update item mapping:

```go
		serviceReq.Items[i] = interfaces.TransactionItemRequest{
			ProductID:    productID,
			Quantity:     item.Quantity,
			WarrantyDays: item.WarrantyDays,
			Price:        item.Price,
		}
```

- [ ] **Step 3: Build and run all tests**

```bash
go build ./... && go test ./... -v
```

Expected: build succeeds, all tests pass.

- [ ] **Step 4: Commit**

```bash
git add internal/handler/transaction_handler.go
git commit -m "feat: add price field to transaction item handler request"
```

---

### Task 7: Run migration and verify

- [ ] **Step 1: Build the binary**

```bash
go build -o bin/rh-pos cmd/main.go
```

Expected: `bin/rh-pos` produced, no errors.

- [ ] **Step 2: Run migration**

```bash
./bin/rh-pos migrate up
```

Expected: migration `019_add_is_dynamic_price_to_products` applied successfully.

- [ ] **Step 3: Verify column in DB**

Connect to MySQL and run:

```sql
DESCRIBE products;
```

Expected: `is_dynamic_price` column present, type `tinyint(1)`, default `0`.
