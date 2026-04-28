# Promotion Expansion Implementation Plan

## 1. Goal And Scope

Implement incremental support for three additional promotion shapes in `rh-pos` without replacing the current campaign flow with a generic rule engine:

1. `buy_x_qty_get_discount_amount`
2. `buy_x_qty_get_discount_percentage`
3. `buy_x_product_get_y_product_free`

Scope of this plan:

- Extend existing discount campaign management so admins can define the new promo shapes.
- Refactor transaction pricing in `internal/usecase/transaction_service.go` so the backend evaluates all eligible promos, applies the most beneficial result for the customer, and supports repeatable applications.
- Preserve current clean architecture boundaries: handler -> usecase -> repository -> database.
- Preserve the existing rule that transaction-level discount excludes items already affected by a campaign.
- Support the agreed free-item flow where frontend sends the free item and backend validates it.

Out of scope for this phase:

- A reusable rules engine.
- Multi-campaign stacking on the same purchased units.
- Cross-category, brand-wide, cart-threshold, or customer-segment promotions.
- Retroactive recomputation of historical transactions.

## 2. Current State Summary

Current implementation is narrow and product-scoped:

- `internal/domain/entities/discount_campaign.go`
  - `discount_campaigns` only store `name`, `discount_percentage`, dates, and `tenant_id`.
  - `discount_campaign_products` maps campaigns to products.
- `internal/usecase/discount_campaign_service.go`
  - Validates only `discount_percentage` and date range.
- `internal/repository/discount_campaign_repository.go`
  - `GetActiveCampaignsForProduct(ctx, productID)` fetches active campaigns for one product.
- `internal/usecase/transaction_service.go`
  - Loads active campaigns per requested product.
  - Chooses the highest `discount_percentage`.
  - Writes discounted unit price into `transaction_items.price`.
  - Stores `discount_percentage` and `campaign_id` on the item.
  - Excludes campaign-affected items from transaction-level discount by splitting totals into `campaignDiscountedTotal` and `regularTotal`.
- `internal/domain/entities/transaction.go`
  - `transaction_items` currently persist `price`, `discount_percentage`, `campaign_id`, and `warranty_days`.
- `internal/repository/transaction_repository.go`
  - Reporting uses `SUM(ti.price * ti.quantity)`.

Important constraint from the current shape:

- `transaction_items.price` currently acts like final unit price after campaign discount.
- That works for simple per-unit percentage discounts, but it is not expressive enough by itself for amount-off bundles and free-item validation unless we store a little more campaign outcome metadata.

## 3. Proposed Campaign Types And Data Model Changes

### Campaign Type Model

Keep one `discount_campaigns` table and extend it with typed parameters instead of introducing a separate promo engine.

Recommended campaign types:

| Type | Meaning | Example |
|---|---|---|
| `product_percentage_discount` | Existing behavior | Product A 20% off |
| `buy_x_qty_get_discount_amount` | Buy X of the same product, get fixed amount off the matched group | Buy 3 of A, get Rp10.000 off |
| `buy_x_qty_get_discount_percentage` | Buy X of the same product, get percentage off the matched group | Buy 3 of A, get 10% off |
| `buy_x_product_get_y_product_free` | Buy trigger product, receive another product for free | Buy A, get B free |

Use explicit string constants in Go, not free-form strings scattered across handlers and services.

### `discount_campaigns` Table Changes

Add columns to support typed campaigns while keeping existing rows valid:

| Column | Type | Purpose |
|---|---|---|
| `campaign_type` | `varchar(50)` | Distinguishes current and new promo shapes |
| `buy_quantity` | `int unsigned NULL` | Trigger quantity for repeatable qty promos |
| `discount_amount` | `decimal(10,2) NULL` | Fixed amount off per applied group |
| `reward_product_id` | `int unsigned NULL` | Free product for buy-X-get-Y |
| `reward_quantity` | `int unsigned NULL` | Number of free units per application |

Keep existing columns:

- `discount_percentage`
- `start_date`
- `end_date`
- `tenant_id`

Recommended migration behavior:

- Backfill existing campaigns with `campaign_type = 'product_percentage_discount'`.
- Keep `discount_percentage` nullable or defaulted for types that do not use it.
- Add FK from `reward_product_id` to `products.id` with `ON DELETE RESTRICT` or `SET NULL` based on current deletion expectations. `RESTRICT` is safer for active config consistency.

### Product Association Strategy

Do not replace `discount_campaign_products`.

Use it as the trigger product mapping for all campaign types:

- `product_percentage_discount`: associated products are the discounted products.
- `buy_x_qty_*`: associated products are the qualifying products and also the discounted products.
- `buy_x_product_get_y_product_free`: associated products are the trigger products; `reward_product_id` identifies the free product.

This keeps repository changes small and preserves current admin flows for attaching products.

### `transaction_items` Table Changes

Add minimal metadata so transaction records remain auditable and free-item validation is possible without recalculation guesswork.

Recommended new columns:

| Column | Type | Purpose |
|---|---|---|
| `discount_amount` | `decimal(10,2) NOT NULL DEFAULT 0.00` | Actual discount value allocated to this line |
| `campaign_type` | `varchar(50) NULL` | Historical record of applied promo type |
| `is_free_item` | `boolean NOT NULL DEFAULT false` | Marks reward lines in buy-X-get-Y-free |
| `campaign_group_key` | `varchar(64) NULL` | Groups purchased line and free/reward line when one promo affects multiple lines |

Keep current fields:

- `price`
- `discount_percentage`
- `campaign_id`
- `warranty_days`

Recommended interpretation after change:

- `price` stays as the persisted final unit price charged for that line.
- `discount_amount` stores the actual total discount value realized by that line.
- `discount_percentage` remains populated for percentage-based promos and `0` for amount/free-item promos.
- `is_free_item=true` means that line is reward inventory and must not be eligible for transaction-level discount.

Why this is the smallest useful schema expansion:

- Reporting can continue to use `price * quantity` with no immediate rewrite.
- Historical transaction inspection becomes reliable for all promo types.
- We avoid introducing a separate `transaction_item_campaigns` detail table unless future complexity forces it.

## 4. API Request/Response Changes

### Discount Campaign Create/Update API

Files impacted:

- `internal/handler/discount_campaign_handler.go`
- `internal/domain/interfaces/services.go`
- `internal/usecase/discount_campaign_service.go`
- `docs/swagger.json`
- likely generated Swagger docs in `docs/docs.go`

Extend create/update payloads to support typed campaign configuration.

Recommended create request shape:

```json
{
  "name": "Buy 3 A get 10000 off",
  "campaign_type": "buy_x_qty_get_discount_amount",
  "discount_percentage": null,
  "discount_amount": 10000,
  "buy_quantity": 3,
  "reward_product_id": null,
  "reward_quantity": null,
  "start_date": "2026-04-13T00:00:00Z",
  "end_date": "2026-04-30T23:59:59Z",
  "product_ids": ["hashed-trigger-product-id"]
}
```

Validation rules by type:

- `product_percentage_discount`
  - requires `discount_percentage`
  - forbids `discount_amount`, `buy_quantity`, `reward_product_id`, `reward_quantity`
- `buy_x_qty_get_discount_amount`
  - requires `buy_quantity >= 2`
  - requires `discount_amount > 0`
  - forbids `reward_product_id`, `reward_quantity`
- `buy_x_qty_get_discount_percentage`
  - requires `buy_quantity >= 2`
  - requires `discount_percentage > 0 && <= 100`
  - forbids `reward_product_id`, `reward_quantity`
- `buy_x_product_get_y_product_free`
  - requires at least one trigger product in `product_ids`
  - requires `buy_quantity >= 1`
  - requires `reward_product_id`
  - requires `reward_quantity >= 1`
  - forbids `discount_amount`
  - `discount_percentage` should be `0` or omitted

Response shape should expose enough metadata for admin UI to display the campaign cleanly:

- `campaign_type`
- `buy_quantity`
- `discount_amount`
- `reward_product_id`
- `reward_quantity`
- existing `discount_percentage`
- existing associated products list

### Transaction Create API

Files impacted:

- `internal/handler/transaction_handler.go`
- `internal/domain/interfaces/services.go`
- `internal/usecase/transaction_service.go`
- Swagger docs

Keep the request mostly unchanged, but extend item payload so the frontend can explicitly send a free reward line.

Recommended request shape per item:

```json
{
  "product_id": "hashed-product-id",
  "quantity": 1,
  "warranty_days": 0,
  "is_free_item": true,
  "campaign_id": "hashed-campaign-id"
}
```

Notes:

- `campaign_id` on request is only needed for free-item lines. It lets backend validate that the client-selected free product matches an active `buy_x_product_get_y_product_free` campaign.
- Purchased lines should not trust client-supplied discount data.
- Keep `total_price` in request optional/ignored as today; server remains authoritative.

### Transaction Response

Extend `formatTransactionItems` in `internal/handler/transaction_handler.go` to include:

- `discount_amount`
- `campaign_type`
- `is_free_item`
- `campaign_group_key` only if non-empty

This is useful for frontend receipts and for debugging pricing decisions.

## 5. Transaction Calculation / Refactor Plan

### Refactor Direction

Do not leave the promotion logic inline inside the current `for _, item := range req.Items` loop.

Instead, keep `transaction_service.go` as the transaction orchestrator and extract a small internal promotion evaluator in the same package, for example:

- `internal/usecase/promotion_pricing.go`

This file should stay focused on transaction pricing, not a general framework.

### Target Flow In `CreateTransaction`

Refactor `CreateTransaction` into these steps:

1. Validate request basics.
2. Load tenant ID from context.
3. Normalize request items into an internal working model.
   - Merge repeated request lines by `product_id` only for promo evaluation if needed, but preserve original lines or rebuild final lines consistently.
   - Separate requested purchased lines from requested `is_free_item=true` lines.
4. Bulk-load all involved products.
5. Bulk-load all active campaigns relevant to the purchased products and requested free-item campaigns.
6. Compute the best applicable campaign result.
7. Materialize final `entities.TransactionItem` rows with authoritative prices and campaign metadata.
8. Deduct stock for all final lines, including free-item reward stock.
9. Apply transaction-level discount only to lines untouched by campaigns.
10. Persist transaction and items.

### Repository Support For Pricing

Current repository API is too per-product for the new logic. Replace or supplement:

- current: `GetActiveCampaignsForProduct(ctx, productID)`
- proposed: `GetActiveCampaignsForProducts(ctx, productIDs []uint) ([]entities.DiscountCampaign, error)`

Repository behavior:

- return active tenant-scoped campaigns for any of the given trigger products
- preload `Products`
- preload `reward_product_id` relation if a `Product` relation is added for reward product on the entity

This avoids N queries inside transaction creation.

### Internal Pricing Model

Introduce a small internal model in usecase code only, not exposed in handlers:

- input line: product ID, requested quantity, unit price, warranty days, whether it was marked as free item by frontend, requested campaign ID
- candidate result: total charge, total discount, affected purchased units, reward units, applied campaign references
- final line output: one or more `entities.TransactionItem`

### Promotion Evaluation Rules By Type

#### A. `product_percentage_discount`

Current behavior, with one improvement:

- For each product line, calculate effective discount value on the purchased quantity.
- Compare this value against other eligible promo types targeting the same purchased units.

#### B. `buy_x_qty_get_discount_amount`

For a line with quantity `qty`, unit price `price`, and campaign requiring `buy_quantity` with `discount_amount`:

- `applications = qty / buy_quantity`
- total line discount = `applications * discount_amount`
- final line total = `(qty * price) - total line discount`
- final persisted line unit `price` can be allocated as `final line total / qty`

Important implementation note:

- Use decimal-safe rounding to 2 decimals when mapping grouped discount back to unit price.
- Store actual aggregate discount in `transaction_items.discount_amount` so the receipt and audits do not depend on reverse math.

For now, because the confirmed requirement is "buy 3 of the same item get Rp10.000 off", support only same-SKU quantity matching, not mixed product groups.

#### C. `buy_x_qty_get_discount_percentage`

Same structure as amount-off bundle:

- `applications = qty / buy_quantity`
- eligible units = `applications * buy_quantity`
- discount value = `eligible units * unit price * discount_percentage / 100`
- remaining units outside the applied groups stay full price

Persist:

- `discount_percentage`
- `discount_amount`
- final unit `price` after allocated average

#### D. `buy_x_product_get_y_product_free`

Algorithm:

1. Evaluate trigger quantity for each applicable campaign.
2. `applications = trigger_qty / buy_quantity`
3. allowed free quantity = `applications * reward_quantity`
4. Find request lines flagged `is_free_item=true` that reference the same `campaign_id`.
5. Validate:
   - campaign exists and is active
   - reward product matches campaign `reward_product_id`
   - requested free quantity does not exceed allowed free quantity
   - reward item was actually sent by frontend when campaign is used
6. Set free-item line final `price = 0`
7. Set free-item line `discount_amount = reward unit price * free quantity`
8. Mark line `is_free_item = true`
9. Attach same `campaign_id` and a generated `campaign_group_key` to both trigger and reward lines

If frontend sends an invalid or excessive free line, reject the transaction with `400`, not silent adjustment.

### Where To Apply Transaction-Level Discount

Preserve current business rule with a slightly broader definition of "affected by a campaign":

- any line with `campaign_id != nil` is excluded from transaction-level discount
- any line with `is_free_item = true` is excluded from transaction-level discount
- only untouched regular lines participate in the transaction-level percentage discount

This means:

- trigger lines participating in bundle promos are excluded
- reward/free lines are excluded
- unrelated products still receive the transaction-level discount

### Persistence Approach

Use the same DB transaction already present in `transaction_service.go`.

Within the transaction:

- compute final priced lines first
- deduct stock for every persisted line quantity, including free items
- create the `transactions` row with authoritative `TotalPrice`
- create associated `transaction_items`

## 6. Conflict Resolution Rules

Use a simple deterministic rule set, not stacking.

### Primary Rule

For each overlapping promo opportunity, choose the result with the highest customer benefit measured as total discount value in currency.

Examples:

- Product A has 20% off and buy-3-get-10000-off.
- For quantity 3 at Rp15.000 each:
  - 20% off => Rp9.000 discount
  - buy-3-get-10000-off => Rp10.000 discount
  - choose the bundle amount promo

### Non-Stacking Rule

The same purchased units cannot receive more than one campaign.

That means:

- no percentage promo plus quantity promo on the same units
- no quantity promo plus free-item promo on the same trigger units

### Repeatability Rule

Repeatable applications are allowed per campaign:

- buy 3 get Rp10.000 off, quantity 6 => discount applied twice
- buy 1 A get 1 B free, quantity 2 A with 2 valid B reward lines => free promo applied twice

### Tie-Breaker Rule

If two promos produce the exact same customer benefit:

1. prefer free-item promo over percentage/amount promo only if frontend supplied a valid matching reward line
2. otherwise prefer the simpler same-line discount in this order:
   1. `buy_x_qty_get_discount_amount`
   2. `buy_x_qty_get_discount_percentage`
   3. `product_percentage_discount`
3. if still tied, prefer lower campaign ID for deterministic behavior

This avoids unstable pricing across runs.

### Request Validation Rule For Free Items

Frontend may send free items, but backend remains authoritative:

- if a free line is sent without a valid campaign, reject
- if a valid free-item campaign is available but the reward line is missing, reject or explicitly decide not to apply the free-item campaign and fall back to a different promo

Recommended pragmatic behavior:

- when evaluating a free-item campaign, only consider it a valid candidate if the required free item line is present and valid
- otherwise, ignore that free-item candidate and continue comparing other eligible promos
- reject only if the client explicitly sent an invalid `is_free_item` line

This keeps the system resilient to UI omissions while still validating explicit reward claims.

## 7. File-By-File Implementation Plan

### Database / Migrations

Add a new migration after the current latest migration, for example:

- `migrations/018_expand_discount_campaigns_for_promotions.sql`

Migration contents:

- alter `discount_campaigns` to add:
  - `campaign_type`
  - `buy_quantity`
  - `discount_amount`
  - `reward_product_id`
  - `reward_quantity`
- backfill existing rows to `product_percentage_discount`
- add FK/index for `reward_product_id`
- alter `transaction_items` to add:
  - `discount_amount`
  - `campaign_type`
  - `is_free_item`
  - `campaign_group_key`

### `internal/domain/entities/discount_campaign.go`

Extend `DiscountCampaign` with:

- `CampaignType string`
- `BuyQuantity *int` or `int` with zero-default
- `DiscountAmount *float64` or `float64` with zero-default
- `RewardProductID *uint`
- optional `RewardProduct *Product`
- `RewardQuantity *int` or `int` with zero-default

Keep `Products []DiscountCampaignProduct` as-is for trigger product mapping.

Add Go constants for campaign types in this file or a nearby domain file.

### `internal/domain/entities/transaction.go`

Extend `TransactionItem` with:

- `DiscountAmount float64`
- `CampaignType *string` or `string`
- `IsFreeItem bool`
- `CampaignGroupKey *string` or `string`

Keep `Price` as final charged unit price.

### `internal/domain/interfaces/services.go`

Update request DTOs:

- `CreateCampaignRequest`
  - add `CampaignType`
  - add `BuyQuantity`
  - add `DiscountAmount`
  - add `RewardProductID`
  - add `RewardQuantity`
- `UpdateCampaignRequest`
  - pointer versions of the same fields
- `TransactionItemRequest`
  - add `IsFreeItem bool`
  - add `CampaignID *uint`

### `internal/domain/interfaces/repositories.go`

Replace or add repository methods:

- add `GetActiveCampaignsForProducts(ctx context.Context, productIDs []uint) ([]entities.DiscountCampaign, error)`

Optionally keep `GetActiveCampaignsForProduct` temporarily during refactor, then remove once transaction service is updated.

### `internal/repository/discount_campaign_repository.go`

Changes:

- update create/list/get/update mappings for new fields
- implement `GetActiveCampaignsForProducts`
- preload `Products` and reward product if relation is added
- ensure tenant-scoped filtering still applies

Query goal:

- one active-campaign fetch for all relevant trigger products

### `internal/usecase/discount_campaign_service.go`

Changes:

- replace single-field percentage validation with per-type validation
- validate incompatible field combinations
- validate reward product presence for free-item campaigns
- preserve current date validation and audit logging

Keep this service focused on campaign configuration, not pricing.

### `internal/handler/discount_campaign_handler.go`

Changes:

- extend create/update request structs for new fields
- decode `reward_product_id` from hashed ID when present
- return `400` for invalid type-specific payloads
- include new fields in `formatCampaignResponse`
- update Swagger annotations and response examples

### `internal/usecase/transaction_service.go`

Main refactor target.

Changes:

- remove per-item repository calls for active campaigns
- normalize item inputs
- call a pricing helper to evaluate promotions across the full transaction
- persist campaign metadata onto final `entities.TransactionItem`
- exclude any campaign-affected line from transaction-level discount
- deduct stock for free reward items too

The outer DB transaction and tenant checks should remain in place.

### `internal/usecase/promotion_pricing.go`

New file recommended.

Responsibilities:

- build promo candidates from loaded campaigns and requested items
- compare candidate discount values
- allocate final line pricing and metadata
- validate frontend-supplied free items

Keep it package-private and narrow to transaction creation.

### `internal/handler/transaction_handler.go`

Changes:

- extend request item struct with `is_free_item` and optional `campaign_id`
- decode hashed `campaign_id` when present
- pass new fields to service request
- include new pricing metadata in formatted response

### `internal/repository/transaction_repository.go`

Likely no structural write-path changes if GORM auto-persists new fields from the entity.

But review report query:

- current revenue query uses `SUM(ti.price * ti.quantity)`
- this remains correct if `price` stays final charged unit price
- free items naturally contribute `0`

No report query rewrite should be required in this phase.

### `internal/server/router.go`

No new endpoints required.

Possible small adjustments:

- verify write protection consistency for `campaigns.PUT("/:id", ...)`
  - currently `PUT` is not admin-protected while create/delete are
  - decide whether to keep as-is or align while touching promo admin flows

This is adjacent but worth reviewing during implementation.

### `cmd/main.go`

Update constructor wiring only if the new pricing helper needs additional dependencies.

Preferred approach:

- keep helper internal to usecase package so `NewTransactionService(...)` signature does not need to change

### Swagger / Docs

Files likely impacted:

- `docs/swagger.json`
- `docs/docs.go`

Update generated API docs after handler annotation changes so frontend has the new request shape documented.

### Tests

Add focused usecase tests, since current repository has almost no transaction/campaign tests.

Recommended files:

- `internal/usecase/transaction_service_test.go`
- `internal/usecase/discount_campaign_service_test.go`
- optionally `internal/handler/transaction_handler_test.go` if request decoding risk grows

## 8. Test Plan

### Unit Tests: Campaign Validation

`internal/usecase/discount_campaign_service_test.go`

Cover:

- create percentage campaign succeeds with valid payload
- create qty-amount campaign fails without `buy_quantity`
- create qty-percent campaign fails when `discount_percentage > 100`
- create free-item campaign fails without `reward_product_id`
- create free-item campaign fails without `reward_quantity`
- update rejects invalid field combinations

### Unit Tests: Transaction Pricing

`internal/usecase/transaction_service_test.go`

Cover at minimum:

1. Existing product percentage behavior still works.
2. Buy 3 same item get Rp10.000 off applies once.
3. Buy 6 same item get Rp10.000 off applies twice.
4. Buy 3 same item get 10% off applies only to complete groups.
5. Buy A get B free succeeds when frontend sends valid free B line.
6. Buy A get B free rejects when frontend sends wrong free product.
7. Buy A get B free rejects when free quantity exceeds entitlement.
8. Free-item promo is ignored as a candidate when no free line is provided and another valid promo exists.
9. Conflict between percentage promo and qty-amount promo chooses the larger currency discount.
10. Transaction-level discount applies only to unaffected lines.
11. Free items reduce stock.
12. Campaign metadata is persisted on resulting transaction items.

### Repository Tests

If repository test setup is reasonable in this repo, add coverage for:

- `GetActiveCampaignsForProducts` returns only active tenant campaigns
- campaigns preload trigger products correctly

If repository tests are expensive to add, prioritize usecase tests first.

### Handler Tests

Optional but useful if time allows:

- hashed `reward_product_id` decoding
- hashed transaction item `campaign_id` decoding
- invalid free-item request body returns `400`

### Manual Verification

After implementation:

1. Run `go test -v ./...`
2. Run `go build -o bin/rh-pos cmd/main.go`
3. Run migrations
4. Create one campaign of each type via API
5. Create transactions covering:
   - plain percentage promo
   - repeatable buy-3 amount promo
   - buy-X-get-free-Y with valid reward line
   - mixed cart with transaction-level discount on unaffected items
6. Verify list/get transaction responses expose expected campaign metadata
7. Verify report totals still match charged revenue

## 9. Risks / Out Of Scope

### Risks

- Allocating grouped discount back into per-unit `price` can introduce rounding drift.
  - Mitigation: keep `discount_amount` as authoritative stored value per line.
- Free-item promos couple backend validation to frontend request composition.
  - Mitigation: only trust `is_free_item` and `campaign_id` as hints; validate everything server-side.
- Conflict resolution can become ambiguous when several campaign types target the same units.
  - Mitigation: keep explicit non-stacking and tie-break rules.
- Current test coverage around transaction pricing is very thin.
  - Mitigation: add usecase tests before or alongside refactor.
- Existing admin API permissions for campaign update may be inconsistent.
  - Mitigation: confirm desired authorization while touching campaign routes.

### Explicitly Out Of Scope

- Generic expression-based promotion rules
- Stacking multiple campaigns on one line
- Mixed-product bundle qualification beyond the confirmed same-item quantity discount
- Auto-inserting free-item lines on behalf of frontend
- Historical transaction repricing after campaign edits
- Reporting changes beyond ensuring current revenue math still works with final charged unit prices

## Recommended Delivery Order

1. Add migration and entity/interface fields.
2. Update campaign CRUD validation and API payloads.
3. Add repository bulk campaign fetch.
4. Add transaction pricing helper and refactor transaction service.
5. Update transaction handler request/response shape.
6. Add tests for campaign validation and transaction pricing.
7. Regenerate Swagger docs.
8. Run build, tests, and manual promo verification.

This sequence keeps the work incremental and reduces the chance of breaking the current transaction flow while promotions are expanded.
