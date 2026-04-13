# Feature Updates Design — 2026-04-13

## Overview

Five incremental feature updates to the rh-pos backend. All changes are confined to their respective layers per the Clean Architecture convention: entity → migration → usecase → handler → router.

---

## Feature 1: SKU Nullable

### Goal
SKU is no longer required when creating or updating a product.

### Entity Change
`internal/domain/entities/product.go`
- Change `SKU string` to `SKU *string`
- GORM tag: remove `not null`, keep `uniqueIndex:idx_tenant_sku`

MySQL allows multiple `NULL` values in a composite unique index, so two products under the same tenant may both have `NULL` SKU without conflict.

### Migration (015)
```sql
ALTER TABLE products MODIFY COLUMN sku VARCHAR(191) NULL;
```

### Handler Change
Remove `validate:"required"` from the `SKU` field in `CreateProductRequest` and `UpdateProductRequest` in `internal/handler/product_handler.go`.

---

## Feature 2: Tenant Terms of Service (Syarat dan Ketentuan)

### Goal
Tenants can store a terms-of-service text used in print settings.

### Entity Change
`internal/domain/entities/tenant.go`
- Add `TermsOfService string \`json:"terms_of_service"\``

### Migration (016)
```sql
ALTER TABLE tenants ADD COLUMN terms_of_service TEXT NULL;
```

### Service / Handler Changes
- Include `TermsOfService` in `UpdateTenant` binding in `internal/handler/admin.go`
- The field is automatically included in `GET /admin/tenants/:id` and `GET /api/my-tenant` responses since GORM maps it from the entity

---

## Feature 3: Cashier Can Update (But Not Delete) Discount Campaigns

### Goal
Cashiers can edit an existing discount campaign's name, percentage, and dates. Only admins can create, delete, or manage products within a campaign.

### Router Change (`internal/server/router.go`)
Before:
```go
campaigns.PUT("/:id", campaignHandler.UpdateCampaign, adminMiddleware.AdminOnly)
```
After:
```go
campaigns.PUT("/:id", campaignHandler.UpdateCampaign)
```

All other campaign routes (POST create, DELETE campaign, POST products, DELETE product) retain `adminMiddleware.AdminOnly`.

No service or handler layer changes required.

---

## Feature 4: Cashier Stock Visibility + Admin-Only Product Delete

### Goal
- Cashiers can already see stock (it is part of the product response) — no change needed.
- Add `DELETE /api/products/:id` (admin-only) which was previously missing.

### Router Change (`internal/server/router.go`)
Add:
```go
products.DELETE("/:id", productHandler.DeleteProduct, adminMiddleware.AdminOnly)
```

### Handler Change (`internal/handler/product_handler.go`)
Add `DeleteProduct` handler that calls `productService.DeleteProduct(ctx, id)`.

### Service Interface Change (`internal/domain/interfaces/services.go`)
Add to `ProductService`:
```go
DeleteProduct(ctx context.Context, id uint) error
```

### Repository + Usecase Changes
Add `DeleteProduct` to the product repository and usecase implementations.

---

## Feature 5: `harga_jual` / `harga_modal` Nullable with Copy-On-Null

### Goal
Either price field may be omitted. If only one is provided, the other is set to the same value. Both omitted is a validation error.

### Entity Change (`internal/domain/entities/product.go`)
- `HargaModal float64` → `HargaModal *float64`
- `HargaJual float64` → `HargaJual *float64`
- GORM tags: remove `not null` from both

### Migration (017)
```sql
ALTER TABLE products
  MODIFY COLUMN harga_modal DECIMAL(15,2) NULL,
  MODIFY COLUMN harga_jual  DECIMAL(15,2) NULL;
```

### Business Logic (usecase layer, `internal/usecase/product_usecase.go`)
Applied before both `CreateProduct` and `UpdateProduct`:
- `harga_jual` set, `harga_modal` nil → `harga_modal = harga_jual`
- `harga_modal` set, `harga_jual` nil → `harga_jual = harga_modal`
- Both set → use as-is
- Both nil → return validation error: "at least one of harga_jual or harga_modal must be provided"

### Handler Change
Remove `validate:"required"` and `gt=0` (or adjust to allow nil) from `HargaModal` and `HargaJual` in `CreateProductRequest` / `UpdateProductRequest`.

---

## Migration Sequence

| # | File | Description |
|---|------|-------------|
| 015 | `015_sku_nullable.sql` | Make `products.sku` nullable |
| 016 | `016_add_tenant_terms_of_service.sql` | Add `tenants.terms_of_service` |
| 017 | `017_price_fields_nullable.sql` | Make `harga_modal` and `harga_jual` nullable |

---

## Non-Goals

- No changes to existing transaction records
- No retroactive discount recalculation — FE handles price display using `harga_jual` + `discount_percentage`
- No new role beyond `admin` and `cashier`
