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

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "ok",
		})
	})

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
	api.PUT("/update-password", authHandler.UpdatePassword)

	// Product routes
	products := api.Group("/products")
	products.GET("", productHandler.ListProducts)
	products.POST("", productHandler.CreateProduct)
	products.GET("/:id", productHandler.GetProduct)
	products.PUT("/:id", productHandler.UpdateProduct)
	products.PUT("/:id/stock", productHandler.UpdateStock)
	products.POST("/:id/upload-url", productHandler.GetUploadURL)
	products.GET("/:id/image/bytes", productHandler.GetProductImageBytes)
	products.POST("/:id/image", productHandler.UploadProductImage)

	// Transaction routes
	transactions := api.Group("/transactions")
	transactions.POST("", transactionHandler.CreateTransaction)
	transactions.GET("", transactionHandler.ListTransactions)
	transactions.GET("/:id", transactionHandler.GetTransaction)

	// Report routes
	reports := api.Group("/reports")
	reports.GET("", reportHandler.GetSalesReport)

	return e
}
