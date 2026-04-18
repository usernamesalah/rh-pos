# API Documentation - Kasir PWA

Base URL: `https://your-api-domain.com`

**Catatan Penting:**
- Semua endpoint yang memerlukan autentikasi menggunakan `Authorization: Bearer <token>` header
- ID menggunakan format hashed (base62 encoded) - bukan integer
- Response format konsisten untuk semua endpoint

---

## 1. Authentication (Login)

### POST /auth/login

Login untuk mendapatkan token JWT.

**Request:**
```json
{
  "username": "kasir001",
  "password": "password123"
}
```

**Response (200):**
```json
{
  "status": "success",
  "message": "Login successful",
  "data": {
    "id": "hashed_user_id",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "username": "kasir001",
    "role": "cashier"
  }
}
```

**Error:**
- 401: Invalid credentials

---

## 2. User Profile

### GET /api/profile

Mendapatkan profile user yang sedang login.

**Headers:** `Authorization: Bearer <token>`

**Response (200):**
```json
{
  "status": "success",
  "message": "Profile retrieved successfully",
  "data": {
    "id": "hashed_user_id",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z",
    "username": "kasir001",
    "role": "cashier",
    "tenant_id": "hashed_tenant_id"
  }
}
```

### PUT /api/update-password

Mengganti password user.

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "current_password": "oldpassword123",
  "new_password": "newpassword456"
}
```

**Response (200):**
```json
{
  "status": "success",
  "message": "Password updated successfully"
}
```

**Error:**
- 400: Validation failed
- 401: Invalid current password

---

## 3. Tenant Detail

### GET /api/my-tenant

Mendapatkan informasi tenant/toko saat ini.

**Headers:** `Authorization: Bearer <token>`

**Response (200):**
```json
{
  "status": "success",
  "message": "Tenant information retrieved successfully",
  "data": {
    "id": "hashed_tenant_id",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-15T10:30:00Z",
    "name": "Toko Saya",
    "about": "Toko elektronik terpercaya",
    "address": "Jl. Merdeka No. 123",
    "phone_number": "081234567890",
    "logo": "https://minio.example.com/bucket/logo.png",
    "terms_of_service": "https://example.com/tos"
  }
}
```

### PUT /api/my-tenant

Mengupdate informasi tenant.

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "name": "Toko Elektronik Jaya",
  "about": "Toko elektronik terlengkap",
  "address": "Jl. Sudirman No. 45",
  "phone_number": "081234567890",
  "logo": "https://minio.example.com/bucket/logo.png"
}
```

**Response (200):**
```json
{
  "status": "success",
  "message": "Tenant updated successfully",
  "data": { ... }
}
```

---

## 4. Product List

### GET /api/products

Mendapatkan daftar produk dengan pagination.

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| page | int | 1 | Nomor halaman |
| limit | int | 10 | Jumlah item per halaman (max 100) |

**Response (200):**
```json
{
  "status": "success",
  "message": "Products retrieved successfully",
  "data": [
    {
      "id": "hashed_product_id",
      "created_at": "2024-01-10T08:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z",
      "name": "Laptop ASUS VivoBook",
      "sku": "LAP-ASUS-001",
      "image_url": "https://minio.example.com/bucket/products/image.jpg",
      "harga_modal": 5000000,
      "harga_jual": 5500000,
      "stock": 10
    }
  ],
  "pagination": {
    "total": 50,
    "page": 1,
    "limit": 10
  }
}
```

### GET /api/products/:id

Mendapatkan detail produk berdasarkan ID.

**Headers:** `Authorization: Bearer <token>`

**Response (200):**
```json
{
  "status": "success",
  "message": "Product retrieved successfully",
  "data": {
    "id": "hashed_product_id",
    "created_at": "2024-01-10T08:00:00Z",
    "updated_at": "2024-01-15T10:00:00Z",
    "name": "Laptop ASUS VivoBook",
    "sku": "LAP-ASUS-001",
    "image_url": "https://minio.example.com/bucket/products/image.jpg",
    "harga_modal": 5000000,
    "harga_jual": 5500000,
    "stock": 10
  }
}
```

---

## 5. Campaign / Promo

### GET /api/discount-campaigns

Mendapatkan daftar campaign/diskon yang aktif.

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| page | int | 1 | Nomor halaman |
| limit | int | 10 | Jumlah item per halaman |

**Response (200):**
```json
{
  "status": "success",
  "message": "Campaigns retrieved successfully",
  "data": [
    {
      "id": "hashed_campaign_id",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-15T00:00:00Z",
      "name": "Diskon Ramadan 50%",
      "discount_percentage": 50,
      "start_date": "2024-03-01T00:00:00Z",
      "end_date": "2024-03-31T23:59:59Z",
      "products": [
        {
          "id": "hashed_product_id",
          "product_id": "hashed_product_id",
          "name": "Laptop ASUS",
          "sku": "LAP-ASUS-001"
        }
      ]
    }
  ],
  "pagination": {
    "total": 5,
    "page": 1,
    "limit": 10
  }
}
```

### GET /api/discount-campaigns/:id

Mendapatkan detail campaign berdasarkan ID.

**Headers:** `Authorization: Bearer <token>`

**Response (200):**
```json
{
  "status": "success",
  "message": "Campaign retrieved successfully",
  "data": {
    "id": "hashed_campaign_id",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-15T00:00:00Z",
    "name": "Diskon Ramadan 50%",
    "discount_percentage": 50,
    "start_date": "2024-03-01T00:00:00Z",
    "end_date": "2024-03-31T23:59:59Z",
    "products": [...]
  }
}
```

---

## 6. Transaction (POS)

### POST /api/transactions

Membuat transaksi baru (penjualan).

**Headers:** `Authorization: Bearer <token>`

**Request:**
```json
{
  "items": [
    {
      "product_id": "hashed_product_id_1",
      "quantity": 2,
      "warranty_days": 30
    },
    {
      "product_id": "hashed_product_id_2",
      "quantity": 1,
      "warranty_days": 0
    }
  ],
  "payment_method": "cash",
  "discount": 50000,
  "total_price": 1050000,
  "notes": "Pembelian langsung",
  "customer_name": "Budi Santoso",
  "customer_email": "budi@email.com",
  "customer_phone": "081234567890"
}
```

**Catatan:**
- `warranty_days`: optional, jumlah hari garansi untuk item tersebut
- `payment_method`: bisa `cash`, `transfer`, `qris`, dll
- `discount`: total diskon yang diterapkan

**Response (201):**
```json
{
  "status": "success",
  "message": "Transaction created successfully",
  "data": {
    "id": "hashed_transaction_id",
    "created_at": "2024-01-15T14:30:00Z",
    "updated_at": "2024-01-15T14:30:00Z",
    "items": [
      {
        "product_id": "hashed_product_id_1",
        "quantity": 2,
        "price": 550000,
        "warranty_days": 30,
        "discount_percentage": 0,
        "product": {
          "id": "hashed_product_id_1",
          "name": "Laptop ASUS VivoBook",
          "sku": "LAP-ASUS-001",
          "image": "...",
          "harga_modal": 5000000,
          "harga_jual": 5500000,
          "stock": 8
        }
      }
    ],
    "payment_method": "cash",
    "discount": 50000,
    "total_price": 1050000,
    "notes": "Pembelian langsung",
    "customer_name": "Budi Santoso",
    "customer_email": "budi@email.com",
    "customer_phone": "081234567890",
    "user": {
      "id": "hashed_user_id",
      "username": "kasir001"
    }
  }
}
```

---

## 7. Transaction List

### GET /api/transactions

Mendapatkan daftar transaksi dengan pagination.

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| page | int | 1 | Nomor halaman |
| limit | int | 10 | Jumlah item per halaman |

**Response (200):**
```json
{
  "status": "success",
  "message": "Transactions retrieved successfully",
  "data": [
    {
      "id": "hashed_transaction_id",
      "created_at": "2024-01-15T14:30:00Z",
      "updated_at": "2024-01-15T14:30:00Z",
      "items": [...],
      "payment_method": "cash",
      "discount": 50000,
      "total_price": 1050000,
      "notes": "Pembelian langsung",
      "customer_name": "Budi Santoso",
      "user": {
        "id": "hashed_user_id",
        "username": "kasir001"
      }
    }
  ],
  "pagination": {
    "total": 100,
    "page": 1,
    "limit": 10
  }
}
```

### GET /api/transactions/:id

Mendapatkan detail transaksi berdasarkan ID.

**Headers:** `Authorization: Bearer <token>`

**Response (200):**
```json
{
  "status": "success",
  "message": "Transaction retrieved successfully",
  "data": {
    "id": "hashed_transaction_id",
    "created_at": "2024-01-15T14:30:00Z",
    "updated_at": "2024-01-15T14:30:00Z",
    "items": [...],
    "payment_method": "cash",
    "discount": 50000,
    "total_price": 1050000,
    "notes": "Pembelian langsung",
    "customer_name": "Budi Santoso",
    "customer_email": "budi@email.com",
    "customer_phone": "081234567890",
    "user": {
      "id": "hashed_user_id",
      "username": "kasir001"
    }
  }
}
```

---

## 8. Warranty (Garansi)

### GET /warranty/:transaction_id

Mengecek status garansi transaksi (public - tidak perlu auth).

**Response (200):**
```json
{
  "status": "success",
  "message": "Warranty information retrieved",
  "data": {
    "transaction_id": "hashed_transaction_id",
    "transaction_date": "2024-01-15T14:30:00Z",
    "customer_name": "Budi Santoso",
    "customer_email": "budi@email.com",
    "customer_phone": "081234567890",
    "items": [
      {
        "product_name": "Laptop ASUS VivoBook",
        "quantity": 2,
        "warranty_days": 30,
        "warranty_start": "2024-01-15T14:30:00Z",
        "warranty_end": "2024-02-14T14:30:00Z",
        "is_active": true,
        "days_remaining": 20
      }
    ]
  }
}
```

### GET /warranty/search?phone=...

Mencari garansi berdasarkan nomor telepon pelanggan (public).

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| phone | string | Yes | Nomor telepon pelanggan |

**Response (200):**
```json
{
  "status": "success",
  "message": "Warranties retrieved",
  "data": {
    "transactions": [
      {
        "transaction_id": "hashed_transaction_id_1",
        "transaction_date": "2024-01-15T14:30:00Z",
        "items": [...]
      },
      {
        "transaction_id": "hashed_transaction_id_2",
        "transaction_date": "2024-01-10T10:00:00Z",
        "items": [...]
      }
    ]
  }
}
```

---

## Response Format Standar

Semua response mengikuti format berikut:

```json
{
  "status": "success",
  "message": "Deskripsi response",
  "data": { ... },
  "pagination": {
    "total": 100,
    "page": 1,
    "limit": 10
  }
}
```

**Error Response:**
```json
{
  "status": "error",
  "message": "Pesan error"
}
```

---

## Catatan Penting untuk Frontend

1. **Hashed ID**: Semua ID yang dikirim/diterima menggunakan format hashed (base62). Jangan pernah menggunakan integer langsung.

2. **Image URL**: URL gambar produk bersifat sementara/presigned. Untuk production, pertimbangkan untuk caching atau menggunakan endpoint `/api/products/:id/image/bytes`.

3. **Warranty Public**: Endpoint warranty tidak memerlukan autentikasi - bisa diakses langsung oleh customer.

4. **Pagination**: Untuk list endpoint, selalu sertakan parameter `page` dan `limit`.

5. **Token**: Simpan token dengan aman (localStorage/secure storage) dan sertakan di setiap request yang memerlukan autentikasi.