package server

import (
	"context"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/usernamesalah/rh-pos/docs"
	"github.com/usernamesalah/rh-pos/internal/config"
	"github.com/usernamesalah/rh-pos/internal/handler"
	"github.com/usernamesalah/rh-pos/internal/pkg/ctxkey"
	"github.com/usernamesalah/rh-pos/internal/pkg/hash"
	adminMiddleware "github.com/usernamesalah/rh-pos/internal/pkg/middleware"
)

// CustomValidator wraps the validator
type CustomValidator struct {
	validator *validator.Validate
}

// Validate validates the struct
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// SetupRouter configures the Echo router with all routes and middleware
func SetupRouter(
	cfg *config.Config,
	authHandler *handler.AuthHandler,
	productHandler *handler.ProductHandler,
	transactionHandler *handler.TransactionHandler,
	reportHandler *handler.ReportHandler,
	adminHandler *handler.AdminHandler,
	campaignHandler *handler.DiscountCampaignHandler,
	warrantyHandler *handler.WarrantyHandler,
) *echo.Echo {
	e := echo.New()

	// Set custom validator
	e.Validator = &CustomValidator{validator: validator.New()}

	// Middleware
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: cfg.Server.AllowedOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Swagger documentation
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	docs.RegisterSwaggerHandlers(e)

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "ok",
		})
	})

	// Static file server for local storage
	if cfg.Storage.Type == "local" {
		e.Static("/files", cfg.Storage.LocalPath)
	}

	// Auth routes
	auth := e.Group("/auth")
	auth.POST("/login", authHandler.Login)

	// Admin routes (protected by Basic Auth)
	admin := e.Group("/admin")
	admin.Use(adminMiddleware.AdminAuth(cfg))
	admin.POST("/tenants", adminHandler.CreateTenant)
	admin.GET("/tenants", adminHandler.ListTenants)
	admin.GET("/tenants/:id", adminHandler.GetTenant)
	admin.PUT("/tenants/:id", adminHandler.UpdateTenant)
	admin.POST("/users", adminHandler.CreateUser)

	// Protected routes
	api := e.Group("/api")
	api.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(cfg.JWT.Secret),
		ContextKey: "user",
		SuccessHandler: func(c echo.Context) {
			user := c.Get("user").(*jwt.Token)
			claims := user.Claims.(jwt.MapClaims)

			// Safe user_id extraction
			userIDFloat, ok := claims["user_id"].(float64)
			if !ok {
				c.Logger().Errorf("user_id claim missing or wrong type")
				return
			}
			userID := uint(userIDFloat)
			c.Set("user_id", userID)
			ctx := context.WithValue(c.Request().Context(), ctxkey.UserID, userID)
			c.SetRequest(c.Request().WithContext(ctx))

			// Safe role extraction
			if role, ok := claims["role"].(string); ok {
				c.Set("role", role)
			}

			// Safely handle tenant_id claim
			if tenantIDRaw, ok := claims["tenant_id"]; ok {
				if tenantIDStr, ok := tenantIDRaw.(string); ok {
					decodedTenantID, err := hash.DecodeHashID(tenantIDStr)
					if err == nil {
						c.Set("tenant_id", decodedTenantID)
						// Set tenant_id in the Go context using typed key
						ctx := context.WithValue(c.Request().Context(), ctxkey.TenantID, decodedTenantID)
						c.SetRequest(c.Request().WithContext(ctx))
					} else {
						c.Logger().Errorf("failed to decode tenant_id: %v", err)
					}
				} else {
					c.Logger().Errorf("tenant_id claim is not a string: %T", tenantIDRaw)
				}
			}
		},
	}))

	// User routes
	api.GET("/profile", authHandler.GetProfile)
	api.GET("/my-tenant", authHandler.GetMyTenant)
	api.PUT("/my-tenant", authHandler.UpdateMyTenant, adminMiddleware.AdminOnly)
	api.PUT("/update-password", authHandler.UpdatePassword)
	api.GET("/users", authHandler.ListUsers)
	api.POST("/users", authHandler.CreateUser)
	api.GET("/users/:id", authHandler.GetUser)
	api.PUT("/users/:id", authHandler.UpdateUser)
	api.DELETE("/users/:id", authHandler.DeleteUser)

	// Product routes — read access for all roles, write access for admin only
	products := api.Group("/products")
	products.GET("", productHandler.ListProducts)
	products.GET("/:id", productHandler.GetProduct)
	products.GET("/:id/image/bytes", productHandler.GetProductImageBytes)
	products.POST("", productHandler.CreateProduct, adminMiddleware.AdminOnly)
	products.PUT("/:id", productHandler.UpdateProduct, adminMiddleware.AdminOnly)
	products.PUT("/:id/stock", productHandler.UpdateStock, adminMiddleware.AdminOnly)
	products.DELETE("/:id", productHandler.DeleteProduct, adminMiddleware.AdminOnly)
	products.POST("/:id/upload-url", productHandler.GetUploadURL, adminMiddleware.AdminOnly)
	products.POST("/:id/image", productHandler.UploadProductImage, adminMiddleware.AdminOnly)

	// Transaction routes — cashier and admin
	transactions := api.Group("/transactions")
	transactions.POST("", transactionHandler.CreateTransaction)
	transactions.GET("", transactionHandler.ListTransactions)
	transactions.GET("/:id", transactionHandler.GetTransaction)

	// Report routes — admin only
	reports := api.Group("/reports")
	reports.Use(adminMiddleware.AdminOnly)
	reports.GET("", reportHandler.GetSalesReport)

	// Discount campaign routes — read for all authenticated, write for admin only
	campaigns := api.Group("/discount-campaigns")
	campaigns.GET("", campaignHandler.ListCampaigns)
	campaigns.GET("/:id", campaignHandler.GetCampaign)
	campaigns.POST("", campaignHandler.CreateCampaign, adminMiddleware.AdminOnly)
	campaigns.PUT("/:id", campaignHandler.UpdateCampaign)
	campaigns.DELETE("/:id", campaignHandler.DeleteCampaign, adminMiddleware.AdminOnly)
	campaigns.POST("/:id/products", campaignHandler.AddProducts, adminMiddleware.AdminOnly)
	campaigns.DELETE("/:id/products/:product_id", campaignHandler.RemoveProduct, adminMiddleware.AdminOnly)

	// Public warranty routes (no auth required)
	warranty := e.Group("/warranty")
	warranty.GET("/search", warrantyHandler.SearchByPhone)
	warranty.GET("/:transaction_id", warrantyHandler.CheckWarranty)

	return e
}
