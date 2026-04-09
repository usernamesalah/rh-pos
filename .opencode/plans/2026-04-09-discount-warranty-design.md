# POS Enhancement: Discount Campaigns, Warranty System & Customer Data

## Overview

Three interconnected features for the rh-pos system:
1. **Discount Campaign System** -- time-bound percentage discounts applied to groups of products
2. **Warranty System** -- per-item warranty tracking with public lookup
3. **Customer Data on Transactions** -- capture buyer info for warranty verification

## Existing State

- `Product` already has `harga_modal` (capital price) and `harga_jual` (selling price)
- `Transaction` already has a `discount` field (percentage applied to total)
- Stock is an `int` field on `Product`
- Multi-tenancy via `tenant_id` on all entities, filtered via context

---

## Feature 1: Discount Campaign System

### New Entities

**`DiscountCampaign`**

| Column | Type | Constraints |
|--------|------|-------------|
| id | uint | PK |
| name | string | not null |
| discount_percentage | float64 | not null |
| start_date | time.Time | not null |
| end_date | time.Time | not null |
| tenant_id | *uint | FK -> tenants, index |
| created_at | time.Time | |
| updated_at | time.Time | |

**`DiscountCampaignProduct`** (join table)

| Column | Type | Constraints |
|--------|------|-------------|
| id | uint | PK |
| campaign_id | uint | FK -> discount_campaigns, ON DELETE CASCADE |
| product_id | uint | FK -> products |
| created_at | time.Time | |

Unique constraint on (campaign_id, product_id).

### API Endpoints

All under `/api/` (JWT protected, tenant-scoped):

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/discount-campaigns` | Create campaign with product IDs |
| GET | `/api/discount-campaigns` | List campaigns (paginated) |
| GET | `/api/discount-campaigns/:id` | Get campaign detail with products |
| PUT | `/api/discount-campaigns/:id` | Update campaign |
| DELETE | `/api/discount-campaigns/:id` | Delete campaign |
| POST | `/api/discount-campaigns/:id/products` | Add products to campaign |
| DELETE | `/api/discount-campaigns/:id/products/:product_id` | Remove product from campaign |

### Transaction Integration

When creating a transaction, for each item:
1. Query active campaigns for the product (current time between start_date and end_date)
2. If an active campaign exists: apply campaign discount to `harga_jual`, store discounted price in `transaction_items.price`
3. Store `discount_percentage` and `campaign_id` on the transaction item
4. Items with an active campaign discount are **excluded** from the transaction-level discount
5. The transaction-level discount only applies to items without a campaign discount

New fields on `TransactionItem`:
- `discount_percentage` (float64, default 0)
- `campaign_id` (*uint, nullable, FK -> discount_campaigns)

### Price Calculation Example

Products: A (harga_jual=100k, in campaign 20% off), B (harga_jual=200k, no campaign)
Transaction discount: 10%

- Item A price: 100k * (1 - 20/100) = 80k (campaign discount applied)
- Item B price: 200k (no campaign discount, eligible for transaction discount)
- Subtotal: 80k + 200k = 280k
- Transaction discount applies to B only: 200k * 10/100 = 20k
- Total: 280k - 20k = 260k

---

## Feature 2: Warranty System

### Changes to TransactionItem

New fields:
- `warranty_days` (int, default 0) -- warranty duration in days, set by seller at sale time

Warranty start date is implicitly the transaction's `created_at`. Warranty end = `created_at` + `warranty_days` days.

### Transaction Creation

`TransactionItemRequest` gains:
- `warranty_days` (int, optional, default 0)

### Public Warranty Check Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/warranty/:transaction_id` | Check warranty by hashed transaction ID |
| GET | `/warranty/search?phone=081234567890` | Search warranties by customer phone |

These are **public endpoints** (no authentication required).

**Response format for `/warranty/:transaction_id`:**
```json
{
  "transaction_id": "abc123hash",
  "transaction_date": "2026-01-15T10:30:00Z",
  "customer_name": "Re***",
  "customer_email": "re***@gmail.com",
  "customer_phone": "08****7890",
  "items": [
    {
      "product_name": "LCD iPhone 12",
      "quantity": 1,
      "warranty_days": 30,
      "warranty_start": "2026-01-15T10:30:00Z",
      "warranty_end": "2026-02-14T10:30:00Z",
      "is_active": true,
      "days_remaining": 25
    }
  ]
}
```

Only items with `warranty_days > 0` are included in the response.

**Response format for `/warranty/search?phone=...`:**
```json
{
  "transactions": [
    {
      "transaction_id": "abc123hash",
      "transaction_date": "2026-01-15T10:30:00Z",
      "customer_name": "Re***",
      "active_warranties": 2,
      "items": [...]
    }
  ]
}
```

---

## Feature 3: Customer Data on Transactions

### Changes to Transaction

New fields:
- `customer_name` (string, nullable)
- `customer_email` (string, nullable)
- `customer_phone` (string, nullable)

### CreateTransactionRequest

New fields:
- `customer_name` (string, optional)
- `customer_email` (string, optional)
- `customer_phone` (string, optional)

### Data Masking Rules (for public warranty endpoints)

| Field | Input | Masked Output |
|-------|-------|---------------|
| Name | "Reza Harahap" | "Re***" |
| Email | "reza@gmail.com" | "re***@gmail.com" |
| Phone | "081234567890" | "08****7890" |

Rules:
- Name: first 2 chars + "***"
- Email: first 2 chars of local part + "***@" + domain
- Phone: first 2 digits + "****" + last 4 digits
- If field is empty, omit from response

---

## Database Migrations

One new migration file:

**Migration 007: Add discount campaigns, warranty, and customer data**
- Create `discount_campaigns` table
- Create `discount_campaign_products` table with unique constraint on (campaign_id, product_id)
- Add `discount_percentage`, `campaign_id` to `transaction_items`
- Add `warranty_days` to `transaction_items`
- Add `customer_name`, `customer_email`, `customer_phone` to `transactions`

---

## New Files (Clean Architecture)

### Domain Layer
- `internal/domain/entities/discount_campaign.go` -- DiscountCampaign + DiscountCampaignProduct entities
- Update `internal/domain/entities/transaction.go` -- add fields to Transaction and TransactionItem
- Update `internal/domain/interfaces/repositories.go` -- add DiscountCampaignRepository interface
- Update `internal/domain/interfaces/services.go` -- add DiscountCampaignService, WarrantyService interfaces

### Repository Layer
- `internal/repository/discount_campaign.go` -- GORM implementation

### Usecase Layer
- `internal/usecase/discount_campaign.go` -- campaign CRUD logic
- `internal/usecase/warranty.go` -- warranty check + masking logic
- Update `internal/usecase/transaction.go` -- integrate campaign discount at sale time

### Handler Layer
- `internal/handler/discount_campaign.go` -- campaign CRUD endpoints
- `internal/handler/warranty.go` -- public warranty check endpoints
- Update `internal/handler/transaction.go` -- accept customer info + warranty_days

### Router
- Update `internal/server/router.go` -- add campaign routes under `/api/`, warranty routes at root

### Migration
- `migrations/007_add_discount_campaigns_warranty.sql`

---

## Multi-Tenancy

- `DiscountCampaign` and `DiscountCampaignProduct` are tenant-scoped
- Campaign queries filter by `tenant_id` from context
- Warranty public endpoint loads the transaction without tenant filtering (customer has no tenant context) -- use a dedicated repository method that queries by ID only
- Phone search also queries without tenant filtering but only returns masked data

---

## Edge Cases

- Product in multiple active campaigns: use the campaign with the highest discount percentage
- Campaign deleted while active: existing transaction items retain their `campaign_id` and `discount_percentage` as historical record
- Warranty check for non-existent transaction: return 404
- Transaction with no customer info: warranty check returns items but no customer fields
- Phone search with no results: return empty array
