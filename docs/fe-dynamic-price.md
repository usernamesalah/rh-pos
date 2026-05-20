# Dynamic Price Product — Frontend Integration Guide

Product baru bertipe **dynamic price** — harga tidak disimpan di master product, tapi diinput saat checkout.

---

## Perubahan Summary

| Area | Perubahan |
|---|---|
| Admin — Create Product | Tambah field `is_dynamic_price` (bool) |
| Admin — Update Product | Tambah field `is_dynamic_price` (bool, optional) |
| Admin — List/Get Product | Response sekarang include `is_dynamic_price` |
| POS — Create Transaction | Item sekarang bisa kirim `price` (untuk dynamic price product) |

---

## 1. Admin Page — Product Management

### 1.1 Create Product

**POST** `/api/products`

**Request body — tambah field baru:**

```json
{
  "name": "Ongkos Servis",
  "sku": null,
  "harga_modal": null,
  "harga_jual": null,
  "stock": 0,
  "category_id": null,
  "is_dynamic_price": true
}
```

> Untuk product jasa/service: set `is_dynamic_price: true`. `harga_jual` dan `stock` tidak perlu diisi.
>
> Untuk product biasa: `is_dynamic_price` tidak perlu dikirim (default `false`).

**Response:**

```json
{
  "success": true,
  "message": "Product created successfully",
  "data": {
    "id": "abc123",
    "name": "Ongkos Servis",
    "sku": null,
    "image": "",
    "harga_modal": null,
    "harga_jual": null,
    "stock": 0,
    "category_id": null,
    "is_dynamic_price": true,
    "created_at": "2026-05-20T10:00:00+07:00",
    "updated_at": "2026-05-20T10:00:00+07:00"
  }
}
```

---

### 1.2 Update Product

**PUT** `/api/products/:id`

**Request body — field baru (opsional, partial update):**

```json
{
  "is_dynamic_price": false
}
```

> Kirim hanya field yang ingin diubah. `is_dynamic_price` bisa diubah kapan saja.

**Response:** sama seperti Create response di atas (include `is_dynamic_price`).

---

### 1.3 Get Product

**GET** `/api/products/:id`

**Response — sekarang include `is_dynamic_price`:**

```json
{
  "success": true,
  "message": "Product retrieved successfully",
  "data": {
    "id": "abc123",
    "name": "Ongkos Servis",
    "sku": null,
    "image_url": "",
    "harga_modal": null,
    "harga_jual": null,
    "stock": 0,
    "category_id": null,
    "is_dynamic_price": true,
    "created_at": "2026-05-20T10:00:00+07:00",
    "updated_at": "2026-05-20T10:00:00+07:00"
  }
}
```

---

### 1.4 List Products

**GET** `/api/products?page=1&limit=10`

**Response — setiap item sekarang include `is_dynamic_price`:**

```json
{
  "success": true,
  "message": "Products retrieved successfully",
  "data": [
    {
      "id": "abc123",
      "name": "Ongkos Servis",
      "sku": null,
      "image_url": "",
      "harga_modal": null,
      "harga_jual": null,
      "stock": 0,
      "category_id": null,
      "is_dynamic_price": true,
      "created_at": "2026-05-20T10:00:00+07:00",
      "updated_at": "2026-05-20T10:00:00+07:00"
    },
    {
      "id": "def456",
      "name": "Produk Biasa",
      "sku": "SKU-001",
      "image_url": "https://...",
      "harga_modal": 50000,
      "harga_jual": 75000,
      "stock": 10,
      "category_id": "xyz789",
      "is_dynamic_price": false,
      "created_at": "2026-05-20T09:00:00+07:00",
      "updated_at": "2026-05-20T09:00:00+07:00"
    }
  ],
  "total": 2,
  "page": 1,
  "limit": 10
}
```

---

## 2. POS Page — Checkout / Pembayaran

### 2.1 Create Transaction

**POST** `/api/transactions`

**Perubahan di `items`:** tiap item sekarang bisa include field `price` opsional.

#### Skenario A — Mixed (dynamic + regular product)

```json
{
  "payment_method": "cash",
  "discount": 0,
  "notes": "",
  "items": [
    {
      "product_id": "def456",
      "quantity": 2,
      "warranty_days": 0
    },
    {
      "product_id": "abc123",
      "quantity": 1,
      "warranty_days": 0,
      "price": 75000
    }
  ]
}
```

> `price` pada item `abc123` (dynamic price) = harga yang diinput cashier saat checkout.
>
> `price` pada item `def456` (regular product) = diabaikan, server pakai `harga_jual`.

#### Skenario B — Dynamic price, harga tidak diisi (gratis / 0)

```json
{
  "items": [
    {
      "product_id": "abc123",
      "quantity": 1,
      "warranty_days": 0
    }
  ]
}
```

> `price` tidak dikirim → server default ke `0`.

**Response** tidak berubah dari sebelumnya.

---

## 3. Implementasi Frontend

### 3.1 rhpos-fe (Admin Dashboard)

**File yang diubah:**

| File | Perubahan |
|---|---|
| `lib/api.ts` | Tambah `is_dynamic_price?: boolean` ke interface `Product` |
| `app/products/page.tsx` | Toggle checkbox, skip validasi harga, badge "Dinamis" di tabel, pass ke API |

**Behavior:**
- Form tambah/edit product: checkbox **"Harga Dinamis"** — jika dicentang, `harga_jual` dan `harga_modal` dikirim `null`, `stock` dikirim `0`
- Validasi harga (`validatePrices`) diskip jika `is_dynamic_price = true`
- Tabel product: badge **Dinamis** (orange outline) muncul di kolom nama jika `is_dynamic_price = true`
- Payload create/update: field `is_dynamic_price` selalu dikirim

---

### 3.2 pos-mobile (Cashier App)

**File yang diubah:**

| File | Perubahan |
|---|---|
| `types/index.ts` | Tambah `isDynamicPrice?: boolean` ke `Product` dan `CartItem` |
| `lib/api/dto.ts` | `RawProduct` terima `is_dynamic_price`, `mapProduct` map ke `isDynamicPrice` |
| `stores/cart-store.ts` | `add()` set `price: 0` + `isDynamicPrice: true` untuk dynamic product; tambah action `setItemPrice(id, price)` |
| `components/pos/product-card.tsx` | Dynamic product tampil label **"Harga Dinamis"** (orange) ganti harga; tidak dianggap habis stok |
| `app/(pos)/cart/page.tsx` | Untuk item `isDynamicPrice`: tampil input harga (orange row) di atas row garansi |

**Behavior cart:**
- Tap product dinamis → masuk cart dengan `price: 0`
- Di halaman cart, row input harga muncul di bawah nama item (styling orange)
- Cashier ketik harga → update via `setItemPrice`
- Harga terkirim ke server dalam field `price` di payload item transaksi

**Checkout payload untuk dynamic item:**
```json
{
  "product_id": "abc123",
  "quantity": 1,
  "warranty_days": 0,
  "price": 75000
}
```

---

## 4. Catatan Penting

| Aturan | Detail |
|---|---|
| Stock | Dynamic price product **tidak** dikurangi stoknya saat transaksi |
| Discount campaign | Dynamic price product **tidak** terkena campaign discount |
| Transaction-level discount | Tetap berlaku — `discount` di level transaksi masih applied ke dynamic price item |
| `price` untuk regular product | Diabaikan server — aman dikirim, tidak mengubah apapun |
| Default price | Jika cashier tidak isi harga → dikirim `0` → server catat sebagai gratis |
