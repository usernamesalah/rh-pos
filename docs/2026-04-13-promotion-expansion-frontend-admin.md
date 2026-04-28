# Promotion Expansion API Update For Frontend Admin

Summary of backend changes required for frontend updates related to discount campaigns and transaction payloads.

---

## 1. Discount campaign now supports 4 campaign types

`campaign_type` is now part of campaign create, update, and response payloads.

Supported values:

- `product_percentage_discount`
- `buy_x_qty_get_discount_amount`
- `buy_x_qty_get_discount_percentage`
- `buy_x_product_get_y_product_free`

If `campaign_type` is omitted on create, backend defaults to `product_percentage_discount`.

---

## 2. Campaign create/update payload changed

### Create

Endpoint:

```text
POST /api/discount-campaigns
```

Request body:

```json
{
  "name": "Buy 3 Save 10000",
  "campaign_type": "buy_x_qty_get_discount_amount",
  "discount_percentage": 0,
  "buy_quantity": 3,
  "discount_amount": 10000,
  "reward_product_id": null,
  "reward_quantity": null,
  "start_date": "2026-04-13T00:00:00Z",
  "end_date": "2026-04-30T23:59:59Z",
  "product_ids": ["prod_hash_1", "prod_hash_2"]
}
```

Notes:

- `product_ids` uses hashed product IDs.
- `reward_product_id` also uses hashed product ID.
- `start_date` and `end_date` accept RFC3339 (`2026-04-13T00:00:00Z`) or date-only (`2026-04-13`).
- Invalid campaign config returns `400 Bad Request` with backend validation message.

### Update

Endpoint:

```text
PUT /api/discount-campaigns/:id
```

Request body is partial. Only send fields to be changed.

```json
{
  "campaign_type": "buy_x_product_get_y_product_free",
  "buy_quantity": 2,
  "reward_product_id": "prod_hash_reward",
  "reward_quantity": 1
}
```

Notes:

- `:id` is hashed campaign ID.
- `PUT /api/discount-campaigns/:id` is not admin-only in current backend routing.
- Create, delete, add products, and remove products remain admin-only.

---

## 3. Campaign response changed

Campaign responses now include the new promotion fields.

Example response data:

```json
{
  "id": "camp_hash_1",
  "name": "Buy 2 Get 1 Free",
  "campaign_type": "buy_x_product_get_y_product_free",
  "discount_percentage": 0,
  "buy_quantity": 2,
  "reward_product_id": "prod_hash_reward",
  "reward_quantity": 1,
  "reward_product": {
    "id": "prod_hash_reward",
    "name": "Bonus Product",
    "sku": "BONUS-001"
  },
  "start_date": "2026-04-13T00:00:00Z",
  "end_date": "2026-04-30T23:59:59Z",
  "products": [
    {
      "id": "prod_hash_main",
      "product_id": "prod_hash_main",
      "name": "Main Product",
      "sku": "MAIN-001"
    }
  ],
  "created_at": "2026-04-13T10:00:00Z",
  "updated_at": "2026-04-13T10:00:00Z"
}
```

Field behavior:

- `buy_quantity` only appears when relevant.
- `discount_amount` only appears when relevant.
- `reward_product_id` and `reward_quantity` only appear for free-item campaigns.
- `reward_product` is included when reward product is loaded.

---

## 4. Validation rules per campaign type

Frontend admin form should adapt fields based on `campaign_type`.

### `product_percentage_discount`

Required:

- `discount_percentage` must be `> 0` and `<= 100`

### `buy_x_qty_get_discount_amount`

Required:

- `buy_quantity >= 2`
- `discount_amount > 0`

### `buy_x_qty_get_discount_percentage`

Required:

- `buy_quantity >= 2`
- `discount_percentage > 0` and `<= 100`

### `buy_x_product_get_y_product_free`

Required:

- `buy_quantity >= 1`
- `reward_product_id` is required
- `reward_quantity >= 1`

Also:

- `end_date` must be after `start_date`

Recommended frontend behavior:

- Show/hide fields based on selected `campaign_type`
- Clear irrelevant fields before submit to avoid confusing state
- Show backend validation message directly when request fails with `400`

---

## 5. Admin campaign form mapping

Suggested admin UI mapping:

- `name`: text input
- `campaign_type`: select
- `product_ids`: multi-select of target products
- `start_date`, `end_date`: datetime picker
- `discount_percentage`: number input for percentage-based types
- `buy_quantity`: integer input for buy-X types
- `discount_amount`: currency input for amount-discount type
- `reward_product_id`: product selector for free-item type
- `reward_quantity`: integer input for free-item type

Recommended labels:

- `product_percentage_discount`: Product percentage discount
- `buy_x_qty_get_discount_amount`: Buy X qty get discount amount
- `buy_x_qty_get_discount_percentage`: Buy X qty get discount percentage
- `buy_x_product_get_y_product_free`: Buy X product get Y product free

---

## 6. Transaction create payload changed

Transaction item payload now supports free-item lines tied to a campaign.

Endpoint:

```text
POST /api/transactions
```

Request example:

```json
{
  "payment_method": "cash",
  "discount": 0,
  "notes": "promo transaction",
  "items": [
    {
      "product_id": "prod_hash_main",
      "quantity": 2,
      "warranty_days": 0,
      "is_free_item": false
    },
    {
      "product_id": "prod_hash_reward",
      "quantity": 1,
      "warranty_days": 0,
      "is_free_item": true,
      "campaign_id": "camp_hash_1"
    }
  ]
}
```

New item fields:

- `is_free_item`: boolean
- `campaign_id`: hashed campaign ID, required for free-item lines from frontend flow

Important behavior:

- For free-item campaigns, frontend sends the free line explicitly.
- Backend validates whether the free item matches an active campaign and whether quantity is allowed.
- If frontend sends an invalid free item, backend returns `400`.

---

## 7. Transaction response changed

Transaction item responses now include campaign metadata.

Example item response:

```json
{
  "product_id": "prod_hash_reward",
  "quantity": 1,
  "price": 0,
  "warranty_days": 0,
  "discount_percentage": 0,
  "discount_amount": 5000,
  "campaign_id": "camp_hash_1",
  "campaign_type": "buy_x_product_get_y_product_free",
  "is_free_item": true,
  "campaign_group_key": "c12-p5-r8",
  "product": {
    "id": "prod_hash_reward",
    "name": "Bonus Product",
    "sku": "BONUS-001",
    "image": "products/x.jpg",
    "harga_modal": 2500,
    "harga_jual": 5000,
    "stock": 99
  }
}
```

New response fields per transaction item:

- `discount_amount`
- `campaign_type`
- `is_free_item`
- `campaign_group_key`

Notes:

- `discount_amount` is the authoritative discount value for the line.
- `price` may reflect averaged per-unit pricing after discount distribution.
- `campaign_group_key` can be used by frontend to visually group buy line and free line from the same promotion.

---

## 8. Promotion rules frontend should know

Important business rules already enforced by backend:

- Promotions are repeatable. Example: buy quantity 3 means quantity 6 applies reward twice.
- If multiple campaigns match the same product, backend picks the most beneficial one for the customer.
- Transaction-level `discount` still excludes items affected by campaign pricing.
- For free-item campaigns, no free line sent means no reward is applied.

Implication for frontend:

- Do not try to resolve campaign conflicts on frontend.
- Frontend may preview promotions, but backend remains source of truth.
- For free-item campaigns, frontend should help cashier/admin add the free line explicitly.

---

## 9. Frontend checklist

Admin frontend should update:

- Campaign create form
- Campaign edit form
- Campaign detail view
- Campaign list badges/labels for type
- Product picker for reward product
- Validation messages per type

POS / transaction frontend should update:

- Transaction item payload supports `is_free_item`
- Transaction item payload supports `campaign_id`
- Transaction detail UI can render `discount_amount`, `campaign_type`, `is_free_item`, `campaign_group_key`
- Free-item campaigns should allow sending the reward line explicitly

---

## 10. Routes affected

Discount campaigns:

```text
GET    /api/discount-campaigns
GET    /api/discount-campaigns/:id
POST   /api/discount-campaigns
PUT    /api/discount-campaigns/:id
DELETE /api/discount-campaigns/:id
POST   /api/discount-campaigns/:id/products
DELETE /api/discount-campaigns/:id/products/:product_id
```

Transactions:

```text
POST   /api/transactions
GET    /api/transactions
GET    /api/transactions/:id
```

---

## 11. Example campaign payloads by type

### Product percentage discount

```json
{
  "name": "Ramadan 10%",
  "campaign_type": "product_percentage_discount",
  "discount_percentage": 10,
  "start_date": "2026-04-13T00:00:00Z",
  "end_date": "2026-04-30T23:59:59Z",
  "product_ids": ["prod_hash_1"]
}
```

### Buy X qty get discount amount

```json
{
  "name": "Buy 3 Save 10000",
  "campaign_type": "buy_x_qty_get_discount_amount",
  "buy_quantity": 3,
  "discount_amount": 10000,
  "start_date": "2026-04-13T00:00:00Z",
  "end_date": "2026-04-30T23:59:59Z",
  "product_ids": ["prod_hash_1"]
}
```

### Buy X qty get discount percentage

```json
{
  "name": "Buy 3 Get 10%",
  "campaign_type": "buy_x_qty_get_discount_percentage",
  "buy_quantity": 3,
  "discount_percentage": 10,
  "start_date": "2026-04-13T00:00:00Z",
  "end_date": "2026-04-30T23:59:59Z",
  "product_ids": ["prod_hash_1"]
}
```

### Buy X product get Y product free

```json
{
  "name": "Buy 2 Main Get 1 Bonus",
  "campaign_type": "buy_x_product_get_y_product_free",
  "buy_quantity": 2,
  "reward_product_id": "prod_hash_bonus",
  "reward_quantity": 1,
  "start_date": "2026-04-13T00:00:00Z",
  "end_date": "2026-04-30T23:59:59Z",
  "product_ids": ["prod_hash_main"]
}
```
