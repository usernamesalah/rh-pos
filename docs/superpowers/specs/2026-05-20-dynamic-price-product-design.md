# Dynamic Price Product Design

**Date:** 2026-05-20  
**Status:** Approved

## Summary

Add `is_dynamic_price bool` flag to Product entity. When true, the product's price is entered by the cashier at checkout time (not stored on the product). Primary use case: jasa/service items (e.g., ongkos servis, jasa pasang).

## Section 1: Entity & Schema

**`internal/domain/entities/product.go`** — add field:

```go
IsDynamicPrice bool `json:"is_dynamic_price" gorm:"default:false"`
```

**Migration** — add column to `products` table:

```sql
ALTER TABLE products ADD COLUMN is_dynamic_price TINYINT(1) NOT NULL DEFAULT 0;
```

`HargaJual` remains nullable. For dynamic price products, `HargaJual` is left nil — not used at transaction time.

## Section 2: Transaction Flow

**`TransactionItemRequest`** (handler + service interface) — add optional field:

```go
Price *float64 `json:"price"` // used only when product is_dynamic_price = true
```

**`CreateTransaction()` usecase logic:**

```
if product.IsDynamicPrice:
    - itemPrice = *item.Price if provided, else 0.0
    - SKIP stock deduction
    - SKIP campaign discount lookup
else:
    - Existing logic unchanged (HargaJual, stock deduction, campaign)
```

`TotalPrice` on transaction is always calculated server-side (sum of all items) — no change.

## Section 3: Product CRUD

**`CreateProductRequest`:**
```go
IsDynamicPrice bool `json:"is_dynamic_price"`
```

**`UpdateProductRequest`** (pointer for partial update):
```go
IsDynamicPrice *bool `json:"is_dynamic_price"`
```

**`product_service.go` → `CreateProduct()`** — pass `IsDynamicPrice` to entity.

**`product_service.go` → `UpdateProduct()`** — add to update map:
```go
if req.IsDynamicPrice != nil {
    updates["is_dynamic_price"] = *req.IsDynamicPrice
}
```

**List/Get responses** — `is_dynamic_price` included automatically via JSON marshaling of entity.

No changes to image, stock, or category endpoints.

## Section 4: Validation & Error Handling

- `item.Price == nil` + `is_dynamic_price = true` → default price to `0.0` (no error)
- `item.Price != nil` + `is_dynamic_price = false` → silently ignored, use `HargaJual` as normal
- No minimum price validation for dynamic price items (0 is valid for free services)

## Scope

Files to change:
1. `internal/domain/entities/product.go` — add `IsDynamicPrice` field
2. `internal/domain/interfaces/services.go` — add `Price *float64` to `TransactionItemRequest`
3. `internal/handler/product_handler.go` — add `IsDynamicPrice` to create/update DTOs
4. `internal/handler/transaction_handler.go` — add `Price *float64` to `TransactionItemRequest`
5. `internal/usecase/product_service.go` — handle `IsDynamicPrice` in create/update
6. `internal/usecase/transaction_service.go` — branch logic for dynamic price items
7. New migration file — `ALTER TABLE products ADD COLUMN is_dynamic_price`

## Out of Scope

- No reference/suggested price on dynamic price products
- No campaign discounts on dynamic price items
- No stock tracking for dynamic price items
- No min/max price constraints
