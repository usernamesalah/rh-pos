# API Documentation

## Base URL
```
http://localhost:8080
```

## Authentication

### Bearer Token (JWT)
Used for `/api/*` routes. Include in Authorization header:
```
Authorization: Bearer <token>
```

### Basic Auth
Used for `/admin/*` routes. Include in Authorization header:
```
Authorization: Basic <base64(username:password)>
```

---

## Response Format

### Success Response
```json
{
  "success": true,
  "message": "Success message",
  "data": { ... }
}
```

### Paginated Response
```json
{
  "success": true,
  "message": "Success message",
  "data": [ ... ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 100
  }
}
```

### Error Response
```json
{
  "success": false,
  "message": "Error message"
}
```

---

## Endpoints

### 1. Authentication

#### POST /auth/login
Login to the system.

**Request:**
```json
{
  "username": "string",
  "password": "string"
}
```

**Response (200):**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "id": "hashed_id",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "token": "eyJ...",
    "username": "admin",
    "role": "admin"
  }
}
```

---

### 2. User Profile (JWT Required)

#### GET /api/profile
Get current user profile information.

**Response (200):**
```json
{
  "success": true,
  "message": "Profile retrieved successfully",
  "data": {
    "id": "hashed_id",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "username": "admin",
    "role": "admin",
    "tenant_id": "hashed_tenant_id"
  }
}
```

#### PUT /api/update-password
Update current user's password.

**Request:**
```json
{
  "current_password": "string",
  "new_password": "string"
}
```

**Response (200):**
```json
{
  "success": true,
  "message": "Password updated successfully",
  "data": null
}
```

---

### 3. Tenant (JWT Required)

#### GET /api/my-tenant
Get current user's tenant information.

**Response (200):**
```json
{
  "success": true,
  "message": "Tenant information retrieved successfully",
  "data": {
    "id": "hashed_id",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "name": "Tenant Name",
    "about": "About description",
    "address": "Address",
    "phone_number": "1234567890",
    "logo_url": "https://...",
    "terms_of_service": "Terms..."
  }
}
```

#### PUT /api/my-tenant
Update user's tenant information. **Admin only.**

**Request:**
```json
{
  "name": "string",
  "about": "string",
  "address": "string",
  "phone_number": "string",
  "terms_of_service": "string"
}
```

**Response (200):**
```json
{
  "success": true,
  "message": "Tenant updated successfully",
  "data": { ... }
}
```

#### POST /api/my-tenant/logo
Upload tenant logo. **Admin only.**

**Content-Type:** `multipart/form-data`

**Form Data:**
- `file`: Image file

**Response (200):**
```json
{
  "success": true,
  "message": "Logo uploaded successfully",
  "data": { ... }
}
```

---

### 4. Users Management (JWT Required)

#### GET /api/users
List all users for the current tenant.

**Query Parameters:**
- `page` (optional, default: 1): Page number
- `limit` (optional, default: 10, max: 100): Items per page

**Response (200):**
```json
{
  "success": true,
  "message": "Users retrieved successfully",
  "data": [
    {
      "id": "hashed_id",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "username": "cashier1",
      "role": "cashier"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 5
  }
}
```

#### POST /api/users
Create a new user.

**Request:**
```json
{
  "username": "string",
  "password": "string (min 6 chars)",
  "role": "cashier|admin"
}
```

**Response (201):**
```json
{
  "success": true,
  "message": "User created successfully",
  "data": {
    "id": "hashed_id",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "username": "newuser",
    "role": "cashier"
  }
}
```

#### GET /api/users/:id
Get user by ID.

**Path Parameters:**
- `id`: Hashed user ID

**Response (200):**
```json
{
  "success": true,
  "message": "User retrieved successfully",
  "data": { ... }
}
```

#### PUT /api/users/:id
Update user.

**Path Parameters:**
- `id`: Hashed user ID

**Request:**
```json
{
  "username": "string",
  "role": "cashier|admin"
}
```

**Response (200):**
```json
{
  "success": true,
  "message": "User updated successfully",
  "data": { ... }
}
```

#### DELETE /api/users/:id
Delete user.

**Path Parameters:**
- `id`: Hashed user ID

**Response (200):**
```json
{
  "success": true,
  "message": "User deleted successfully",
  "data": null
}
```

---

### 5. Products (JWT Required)

#### GET /api/products
List all products.

**Query Parameters:**
- `page` (optional, default: 1): Page number
- `limit` (optional, default: 10, max: 100): Items per page

**Response (200):**
```json
{
  "success": true,
  "message": "Products retrieved successfully",
  "data": [
    {
      "id": "hashed_id",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "name": "Product Name",
      "sku": "SKU123",
      "image_url": "https://...",
      "harga_modal": 10000,
      "harga_jual": 15000,
      "stock": 100
    }
  ],
  "pagination": { ... }
}
```

#### GET /api/products/:id
Get product by ID.

**Path Parameters:**
- `id`: Hashed product ID

**Response (200):**
```json
{
  "success": true,
  "message": "Product retrieved successfully",
  "data": { ... }
}
```

#### POST /api/products
Create a new product. **Admin only.**

**Request:**
```json
{
  "name": "string (required)",
  "sku": "string (optional)",
  "image": "string (optional)",
  "harga_modal": 10000,
  "harga_jual": 15000,
  "stock": 100
}
```

**Response (201):**
```json
{
  "success": true,
  "message": "Product created successfully",
  "data": { ... }
}
```

#### PUT /api/products/:id
Update product. **Admin only.**

**Path Parameters:**
- `id`: Hashed product ID

**Request:**
```json
{
  "name": "string (optional)",
  "sku": "string (optional)",
  "harga_modal": 10000 (optional),
  "harga_jual": 15000 (optional)
}
```

**Response (200):**
```json
{
  "success": true,
  "message": "Product updated successfully",
  "data": { ... }
}
```

#### PUT /api/products/:id/stock
Update product stock. **Admin only.**

**Path Parameters:**
- `id`: Hashed product ID

**Request:**
```json
{
  "stock": 100
}
```

**Response (200):**
```json
{
  "success": true,
  "message": "Stock updated successfully",
  "data": { ... }
}
```

#### DELETE /api/products/:id
Delete product. **Admin only.**

**Path Parameters:**
- `id`: Hashed product ID

**Response (200):**
```json
{
  "success": true,
  "message": "Product deleted successfully",
  "data": null
}
```

#### GET /api/products/:id/image/bytes
Get product image as binary.

**Path Parameters:**
- `id`: Hashed product ID

**Response (200):** Binary image data

#### POST /api/products/:id/upload-url
Get presigned upload URL for product image. **Admin only.**

**Path Parameters:**
- `id`: Hashed product ID

**Request:**
```json
{
  "extension": "jpg"
}
```

**Response (200):**
```json
{
  "success": true,
  "message": "Upload URL generated successfully",
  "data": {
    "upload_url": "https://...",
    "image_key": "path/to/image"
  }
}
```

#### POST /api/products/:id/image
Upload product image. **Admin only.**

**Path Parameters:**
- `id`: Hashed product ID

**Content-Type:** `multipart/form-data`

**Form Data:**
- `image`: Image file

**Response (200):**
```json
{
  "success": true,
  "message": "Product image uploaded successfully",
  "data": { ... }
}
```

---

### 6. Transactions (JWT Required)

#### POST /api/transactions
Create a new sales transaction.

**Request:**
```json
{
  "items": [
    {
      "product_id": "hashed_product_id",
      "quantity": 2,
      "warranty_days": 30
    }
  ],
  "payment_method": "cash|transfer|qris",
  "discount": 0,
  "total_price": 30000,
  "notes": "string (optional)",
  "customer_name": "string (optional)",
  "customer_email": "string (optional)",
  "customer_phone": "string (optional)"
}
```

**Response (201):**
```json
{
  "success": true,
  "message": "Transaction created successfully",
  "data": {
    "id": "hashed_id",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "items": [
      {
        "product_id": "hashed_id",
        "quantity": 2,
        "price": 15000,
        "warranty_days": 30,
        "discount_percentage": 0,
        "product": {
          "id": "hashed_id",
          "name": "Product Name",
          "sku": "SKU123",
          "harga_modal": 10000,
          "harga_jual": 15000,
          "stock": 98
        }
      }
    ],
    "payment_method": "cash",
    "discount": 0,
    "total_price": 30000,
    "notes": "",
    "user_id": "hashed_id",
    "user": {
      "id": "hashed_id",
      "username": "cashier1"
    },
    "customer_name": "John Doe",
    "customer_email": "john@example.com",
    "customer_phone": "1234567890"
  }
}
```

#### GET /api/transactions
List all transactions.

**Query Parameters:**
- `page` (optional, default: 1): Page number
- `limit` (optional, default: 10, max: 100): Items per page

**Response (200):**
```json
{
  "success": true,
  "message": "Transactions retrieved successfully",
  "data": [ ... ],
  "pagination": { ... }
}
```

#### GET /api/transactions/:id
Get transaction by ID.

**Path Parameters:**
- `id`: Hashed transaction ID

**Response (200):**
```json
{
  "success": true,
  "message": "Transaction retrieved successfully",
  "data": { ... }
}
```

---

### 7. Reports (JWT Required, Admin Only)

#### GET /api/reports
Get sales report.

**Query Parameters:**
- `start_date` (optional): Start date (YYYY-MM-DD)
- `end_date` (optional): End date (YYYY-MM-DD)

If no dates provided, returns all-time report.

**Response (200):**
```json
{
  "success": true,
  "message": "Sales report retrieved successfully",
  "data": {
    "total_revenue": 1000000,
    "items_sold": 100,
    "average_transaction": 10000,
    "details": [
      {
        "id": "hashed_id",
        "product_id": "hashed_id",
        "product_name": "Product Name",
        "total": 50,
        "total_price": 500000
      }
    ]
  }
}
```

---

### 8. Discount Campaigns (JWT Required)

#### GET /api/discount-campaigns
List all discount campaigns.

**Query Parameters:**
- `page` (optional, default: 1): Page number
- `limit` (optional, default: 10, max: 100): Items per page

**Response (200):**
```json
{
  "success": true,
  "message": "Campaigns retrieved successfully",
  "data": [
    {
      "id": "hashed_id",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "name": "Campaign Name",
      "discount_percentage": 10,
      "start_date": "2024-01-01T00:00:00Z",
      "end_date": "2024-12-31T23:59:59Z",
      "products": [
        {
          "id": "hashed_id",
          "product_id": "hashed_id",
          "name": "Product Name",
          "sku": "SKU123"
        }
      ]
    }
  ],
  "pagination": { ... }
}
```

#### GET /api/discount-campaigns/:id
Get discount campaign by ID.

**Path Parameters:**
- `id`: Hashed campaign ID

**Response (200):**
```json
{
  "success": true,
  "message": "Campaign retrieved successfully",
  "data": { ... }
}
```

#### POST /api/discount-campaigns
Create a new discount campaign. **Admin only.**

**Request:**
```json
{
  "name": "string (required)",
  "discount_percentage": 10 (required, 0-100),
  "start_date": "2024-01-01 or 2024-01-01T00:00:00Z",
  "end_date": "2024-12-31 or 2024-12-31T23:59:59Z",
  "product_ids": ["hashed_product_id1", "hashed_product_id2"]
}
```

**Response (201):**
```json
{
  "success": true,
  "message": "Campaign created successfully",
  "data": { ... }
}
```

#### PUT /api/discount-campaigns/:id
Update discount campaign.

**Path Parameters:**
- `id`: Hashed campaign ID

**Request:**
```json
{
  "name": "string (optional)",
  "discount_percentage": 15 (optional),
  "start_date": "2024-01-01 (optional)",
  "end_date": "2024-12-31 (optional)"
}
```

**Response (200):**
```json
{
  "success": true,
  "message": "Campaign updated successfully",
  "data": { ... }
}
```

#### DELETE /api/discount-campaigns/:id
Delete discount campaign. **Admin only.**

**Path Parameters:**
- `id`: Hashed campaign ID

**Response (200):**
```json
{
  "success": true,
  "message": "Campaign deleted successfully",
  "data": null
}
```

#### POST /api/discount-campaigns/:id/products
Add products to a campaign. **Admin only.**

**Path Parameters:**
- `id`: Hashed campaign ID

**Request:**
```json
{
  "product_ids": ["hashed_product_id1", "hashed_product_id2"]
}
```

**Response (200):**
```json
{
  "success": true,
  "message": "Products added to campaign",
  "data": null
}
```

#### DELETE /api/discount-campaigns/:id/products/:product_id
Remove a product from a campaign. **Admin only.**

**Path Parameters:**
- `id`: Hashed campaign ID
- `product_id`: Hashed product ID

**Response (200):**
```json
{
  "success": true,
  "message": "Product removed from campaign",
  "data": null
}
```

---

### 9. Warranty (Public)

#### GET /warranty/:transaction_id
Check warranty status for a transaction.

**Path Parameters:**
- `transaction_id`: Hashed transaction ID

**Response (200):**
```json
{
  "success": true,
  "message": "Warranty information retrieved",
  "data": {
    "transaction_id": "hashed_id",
    "transaction_date": "2024-01-01T00:00:00Z",
    "customer_name": "John Doe",
    "customer_email": "john@example.com",
    "customer_phone": "1234567890",
    "items": [
      {
        "product_name": "Product Name",
        "quantity": 2,
        "warranty_days": 30,
        "warranty_start": "2024-01-01T00:00:00Z",
        "warranty_end": "2024-01-31T00:00:00Z",
        "is_active": true,
        "days_remaining": 15
      }
    ]
  }
}
```

#### GET /warranty/search
Search warranty records by customer phone number.

**Query Parameters:**
- `phone` (required): Customer phone number

**Response (200):**
```json
{
  "success": true,
  "message": "Warranties retrieved",
  "data": {
    "transactions": [
      { ... }
    ]
  }
}
```

---

### 10. Admin Routes (Basic Auth)

#### POST /admin/tenants
Create a new tenant.

**Request:**
```json
{
  "name": "string",
  "about": "string",
  "address": "string",
  "phone_number": "string",
  "terms_of_service": "string"
}
```

**Response (201):**
```json
{
  "id": 1,
  "name": "Tenant Name",
  "about": "...",
  ...
}
```

#### GET /admin/tenants
List all tenants.

**Response (200):**
```json
[
  {
    "id": 1,
    "name": "Tenant Name",
    ...
  }
]
```

#### GET /admin/tenants/:id
Get tenant by ID.

**Path Parameters:**
- `id`: Tenant ID (numeric)

**Response (200):**
```json
{
  "id": 1,
  "name": "Tenant Name",
  "about": "...",
  "address": "...",
  "phone_number": "...",
  "logo_url": "https://...",
  "terms_of_service": "...",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

#### PUT /admin/tenants/:id
Update a tenant.

**Path Parameters:**
- `id`: Tenant ID (numeric)

**Request:**
```json
{
  "name": "string",
  "about": "string",
  ...
}
```

**Response (200):**
```json
{
  "id": 1,
  ...
}
```

#### POST /admin/tenants/:id/logo
Upload tenant logo.

**Path Parameters:**
- `id`: Tenant ID (numeric)

**Content-Type:** `multipart/form-data`

**Form Data:**
- `file`: Image file

**Response (200):**
```json
{
  "id": 1,
  "name": "Tenant Name",
  "logo_url": "https://...",
  ...
}
```

#### POST /admin/users
Create a new user assigned to a tenant.

**Request:**
```json
{
  "username": "string",
  "password": "string",
  "role": "admin|cashier",
  "tenant_id": 1
}
```

**Response (201):**
```json
{
  "id": 1,
  "username": "user1",
  "role": "cashier",
  "tenant_id": 1,
  ...
}
```

---

### 11. Health Check

#### GET /health
Check API health status.

**Response (200):**
```json
{
  "status": "ok"
}
```

---

## Error Codes

| Code | Description |
|------|-------------|
| 400 | Bad Request - Invalid request body or parameters |
| 401 | Unauthorized - Invalid or missing authentication |
| 404 | Not Found - Resource not found |
| 500 | Internal Server Error - Server error |

---

## Notes

- All IDs in responses are hashed using base62 encoding
- Date format in requests: `YYYY-MM-DD` or `YYYY-MM-DDTHH:MM:SSZ`
- Date format in responses: `YYYY-MM-DDTHH:MM:SSZ` (RFC3339)
- Maximum file upload size: 32MB
- Allowed image extensions: jpg, jpeg, png, gif, webp