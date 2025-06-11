package server

import (
	"github.com/go-playground/validator/v10"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/usernamesalah/rh-pos/internal/config"
	"github.com/usernamesalah/rh-pos/internal/handler"
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
	e.Use(echoMiddleware.CORS())

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
	admin.POST("/users", adminHandler.CreateUser)

	// Protected routes
	api := e.Group("/api")
	api.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(cfg.JWT.Secret),
	}))

	// User routes
	api.GET("/profile", authHandler.GetProfile)

	// Product routes
	products := api.Group("/products")
	products.GET("", productHandler.ListProducts)
	products.GET("/:id", productHandler.GetProduct)
	products.PUT("/:id", productHandler.UpdateProduct)
	products.PUT("/:id/stock", productHandler.UpdateStock)

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
