# API Changes — 2026-04-13

Summary of backend changes that require frontend updates.

---

## 1. Product SKU is now optional

- `sku` field in `POST /api/products` and `PUT /api/products/:id` is no longer required.
- Omit `sku` or send `null` to create a product without an SKU.
- Responses return `"sku": null` when no SKU is set.

**Create request example (no SKU):**

```json
{
  "name": "Widget",
  "harga_jual": 5000,
  "stock": 10
}
```

---

## 2. Tenant has a new `terms_of_service` field

- `GET /admin/tenants/:id`, `POST /admin/tenants`, `PUT /admin/tenants/:id` now include `terms_of_service` (string, nullable).
- Send as plain text or HTML string in create/update requests.

**Example:**

```json
{
  "name": "My Store",
  "terms_of_service": "All sales are final. Warranty applies for 30 days."
}
```

---

## 3. Cashier can now update discount campaigns

- `PUT /api/discount-campaigns/:id` no longer requires admin role.
- Any authenticated user (including cashier) can update a campaign.
- Other campaign write operations (create, delete, add/remove products) still require admin.

---

## 4. New endpoint: Delete product (admin-only)

```
DELETE /api/products/:id
```

- Requires admin role.
- `id` is the hashed product ID.
- Returns `200` on success:

```json
{
  "status": "success",
  "message": "Product deleted successfully",
  "data": null
}
```

---

## 5. Price fields (`harga_modal`, `harga_jual`) are now optional

- Both fields accept `null` or can be omitted in `POST /api/products`.
- **At least one** of `harga_jual` or `harga_modal` must be provided — sending both as `null`/omitted returns a `500` validation error.
- **Copy-on-null logic:** if only one price is provided, the other is automatically set to the same value.
- Responses may return `null` for either field if the database value is NULL (unlikely for new products due to copy-on-null, but possible for legacy data before migration).

**Examples:**

```json
// Only harga_jual provided — harga_modal auto-set to 2000
{ "name": "Widget", "harga_jual": 2000, "stock": 5 }

// Only harga_modal provided — harga_jual auto-set to 1500
{ "name": "Widget", "harga_modal": 1500, "stock": 5 }

// Both provided — used as-is
{ "name": "Widget", "harga_jual": 3000, "harga_modal": 2000, "stock": 5 }

// Both omitted — returns error
{ "name": "Widget", "stock": 5 }
```

---

## Migration note

Backend requires running three new migrations before these changes take effect:

```bash
./bin/rh-pos migrate up
```
