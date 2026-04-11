package docs

import (
	"github.com/labstack/echo/v4"
)

const SwaggerJSON = `{
    "swagger": "2.0",
    "info": {
        "title": "RH POS API",
        "version": "1.0",
        "description": "Point of Sale system API with multi-tenant support, product management, transactions, discount campaigns, warranty tracking, and sales reporting."
    },
    "host": "localhost:8080",
    "basePath": "/",
    "securityDefinitions": {
        "bearerAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header",
            "description": "JWT Bearer token. Format: 'Bearer {token}'"
        },
        "basicAuth": {
            "type": "basic",
            "description": "Basic Auth for admin endpoints"
        }
    },
    "paths": {
        "/auth/login": {
            "post": {
                "summary": "User login",
                "description": "Authenticate user and get JWT token",
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "username": {"type": "string"},
                                "password": {"type": "string"}
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {"description": "Login successful"},
                    "401": {"description": "Invalid credentials"}
                }
            }
        },
        "/api/profile": {
            "get": {
                "summary": "Get user profile",
                "security": [{"bearerAuth": []}],
                "responses": {
                    "200": {"description": "Profile data"}
                }
            }
        },
        "/api/my-tenant": {
            "get": {
                "summary": "Get tenant info",
                "security": [{"bearerAuth": []}],
                "responses": {
                    "200": {"description": "Tenant data"}
                }
            }
        },
        "/api/update-password": {
            "put": {
                "summary": "Update password",
                "security": [{"bearerAuth": []}],
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "current_password": {"type": "string"},
                                "new_password": {"type": "string"}
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {"description": "Password updated"}
                }
            }
        },
        "/api/products": {
            "get": {
                "summary": "List products",
                "security": [{"bearerAuth": []}],
                "parameters": [
                    {"name": "page", "in": "query", "type": "integer"},
                    {"name": "limit", "in": "query", "type": "integer"},
                    {"name": "search", "in": "query", "type": "string"}
                ],
                "responses": {
                    "200": {"description": "Paginated product list"}
                }
            },
            "post": {
                "summary": "Create product",
                "security": [{"bearerAuth": []}],
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "name": {"type": "string"},
                                "sku": {"type": "string"},
                                "harga_modal": {"type": "number"},
                                "harga_jual": {"type": "number"},
                                "stock": {"type": "integer"}
                            }
                        }
                    }
                ],
                "responses": {
                    "201": {"description": "Product created"}
                }
            }
        },
        "/api/products/{id}": {
            "get": {
                "summary": "Get product",
                "security": [{"bearerAuth": []}],
                "parameters": [
                    {"name": "id", "in": "path", "required": true, "type": "string"}
                ],
                "responses": {
                    "200": {"description": "Product data"}
                }
            },
            "put": {
                "summary": "Update product",
                "security": [{"bearerAuth": []}],
                "parameters": [
                    {"name": "id", "in": "path", "required": true, "type": "string"},
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "name": {"type": "string"},
                                "sku": {"type": "string"},
                                "harga_modal": {"type": "number"},
                                "harga_jual": {"type": "number"},
                                "stock": {"type": "integer"}
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {"description": "Product updated"}
                }
            }
        },
        "/api/products/{id}/stock": {
            "put": {
                "summary": "Update product stock",
                "security": [{"bearerAuth": []}],
                "parameters": [
                    {"name": "id", "in": "path", "required": true, "type": "string"},
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "stock": {"type": "integer"}
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {"description": "Stock updated"}
                }
            }
        },
        "/api/products/{id}/upload-url": {
            "post": {
                "summary": "Get upload URL for product image",
                "security": [{"bearerAuth": []}],
                "parameters": [
                    {"name": "id", "in": "path", "required": true, "type": "string"},
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "ext": {"type": "string"}
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {"description": "Upload URL generated"}
                }
            }
        },
        "/api/products/{id}/image": {
            "post": {
                "summary": "Upload product image",
                "security": [{"bearerAuth": []}],
                "consumes": ["multipart/form-data"],
                "parameters": [
                    {"name": "id", "in": "path", "required": true, "type": "string"},
                    {"name": "file", "in": "formData", "type": "file", "required": true}
                ],
                "responses": {
                    "200": {"description": "Image uploaded"}
                }
            }
        },
        "/api/transactions": {
            "get": {
                "summary": "List transactions",
                "security": [{"bearerAuth": []}],
                "parameters": [
                    {"name": "page", "in": "query", "type": "integer"},
                    {"name": "limit", "in": "query", "type": "integer"}
                ],
                "responses": {
                    "200": {"description": "Paginated transaction list"}
                }
            },
            "post": {
                "summary": "Create transaction",
                "security": [{"bearerAuth": []}],
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "items": {
                                    "type": "array",
                                    "items": {
                                        "type": "object",
                                        "properties": {
                                            "product_id": {"type": "string"},
                                            "quantity": {"type": "integer"},
                                            "harga_saat_transaksi": {"type": "number"}
                                        }
                                    }
                                },
                                "payment_method": {"type": "string"},
                                "customer_phone": {"type": "string"},
                                "discount_campaign_id": {"type": "string"}
                            }
                        }
                    }
                ],
                "responses": {
                    "201": {"description": "Transaction created"}
                }
            }
        },
        "/api/transactions/{id}": {
            "get": {
                "summary": "Get transaction",
                "security": [{"bearerAuth": []}],
                "parameters": [
                    {"name": "id", "in": "path", "required": true, "type": "string"}
                ],
                "responses": {
                    "200": {"description": "Transaction data"}
                }
            }
        },
        "/api/reports": {
            "get": {
                "summary": "Get sales report",
                "security": [{"bearerAuth": []}],
                "parameters": [
                    {"name": "start_date", "in": "query", "type": "string"},
                    {"name": "end_date", "in": "query", "type": "string"}
                ],
                "responses": {
                    "200": {"description": "Sales report"}
                }
            }
        },
        "/api/discount-campaigns": {
            "get": {
                "summary": "List discount campaigns",
                "security": [{"bearerAuth": []}],
                "parameters": [
                    {"name": "page", "in": "query", "type": "integer"},
                    {"name": "limit", "in": "query", "type": "integer"},
                    {"name": "active", "in": "query", "type": "boolean"}
                ],
                "responses": {
                    "200": {"description": "Paginated campaign list"}
                }
            },
            "post": {
                "summary": "Create discount campaign",
                "security": [{"bearerAuth": []}],
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "name": {"type": "string"},
                                "discount_type": {"type": "string"},
                                "discount_value": {"type": "number"},
                                "start_date": {"type": "string"},
                                "end_date": {"type": "string"},
                                "is_active": {"type": "boolean"}
                            }
                        }
                    }
                ],
                "responses": {
                    "201": {"description": "Campaign created"}
                }
            }
        },
        "/api/discount-campaigns/{id}": {
            "get": {
                "summary": "Get discount campaign",
                "security": [{"bearerAuth": []}],
                "parameters": [
                    {"name": "id", "in": "path", "required": true, "type": "string"}
                ],
                "responses": {
                    "200": {"description": "Campaign data"}
                }
            },
            "put": {
                "summary": "Update discount campaign",
                "security": [{"bearerAuth": []}],
                "parameters": [
                    {"name": "id", "in": "path", "required": true, "type": "string"},
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "name": {"type": "string"},
                                "discount_type": {"type": "string"},
                                "discount_value": {"type": "number"},
                                "start_date": {"type": "string"},
                                "end_date": {"type": "string"},
                                "is_active": {"type": "boolean"}
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {"description": "Campaign updated"}
                }
            },
            "delete": {
                "summary": "Delete discount campaign",
                "security": [{"bearerAuth": []}],
                "parameters": [
                    {"name": "id", "in": "path", "required": true, "type": "string"}
                ],
                "responses": {
                    "200": {"description": "Campaign deleted"}
                }
            }
        },
        "/api/discount-campaigns/{id}/products": {
            "post": {
                "summary": "Add products to campaign",
                "security": [{"bearerAuth": []}],
                "parameters": [
                    {"name": "id", "in": "path", "required": true, "type": "string"},
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "product_ids": {"type": "array", "items": {"type": "string"}}
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {"description": "Products added"}
                }
            },
            "delete": {
                "summary": "Remove product from campaign",
                "security": [{"bearerAuth": []}],
                "parameters": [
                    {"name": "id", "in": "path", "required": true, "type": "string"},
                    {"name": "product_id", "in": "path", "required": true, "type": "string"}
                ],
                "responses": {
                    "200": {"description": "Product removed"}
                }
            }
        },
        "/warranty/search": {
            "get": {
                "summary": "Search warranty by phone",
                "parameters": [
                    {"name": "phone", "in": "query", "required": true, "type": "string"}
                ],
                "responses": {
                    "200": {"description": "Warranty list"}
                }
            }
        },
        "/warranty/{transaction_id}": {
            "get": {
                "summary": "Check warranty status",
                "parameters": [
                    {"name": "transaction_id", "in": "path", "required": true, "type": "string"}
                ],
                "responses": {
                    "200": {"description": "Warranty status"}
                }
            }
        },
        "/admin/tenants": {
            "get": {
                "summary": "List tenants",
                "security": [{"basicAuth": []}],
                "responses": {
                    "200": {"description": "Tenant list"}
                }
            },
            "post": {
                "summary": "Create tenant",
                "security": [{"basicAuth": []}],
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "name": {"type": "string"},
                                "domain": {"type": "string"}
                            }
                        }
                    }
                ],
                "responses": {
                    "201": {"description": "Tenant created"}
                }
            }
        },
        "/admin/tenants/{id}": {
            "get": {
                "summary": "Get tenant",
                "security": [{"basicAuth": []}],
                "parameters": [
                    {"name": "id", "in": "path", "required": true, "type": "string"}
                ],
                "responses": {
                    "200": {"description": "Tenant data"}
                }
            },
            "put": {
                "summary": "Update tenant",
                "security": [{"basicAuth": []}],
                "parameters": [
                    {"name": "id", "in": "path", "required": true, "type": "string"},
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "name": {"type": "string"},
                                "domain": {"type": "string"},
                                "is_active": {"type": "boolean"}
                            }
                        }
                    }
                ],
                "responses": {
                    "200": {"description": "Tenant updated"}
                }
            }
        },
        "/admin/users": {
            "post": {
                "summary": "Create user",
                "security": [{"basicAuth": []}],
                "parameters": [
                    {
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "properties": {
                                "username": {"type": "string"},
                                "password": {"type": "string"},
                                "role": {"type": "string"},
                                "tenant_id": {"type": "string"}
                            }
                        }
                    }
                ],
                "responses": {
                    "201": {"description": "User created"}
                }
            }
        },
        "/health": {
            "get": {
                "summary": "Health check",
                "responses": {
                    "200": {"description": "Server is healthy"}
                }
            }
        }
    }
}`

func RegisterSwaggerHandlers(e *echo.Echo) {
	e.GET("/doc.json", func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "application/json")
		return c.String(200, SwaggerJSON)
	})
}
