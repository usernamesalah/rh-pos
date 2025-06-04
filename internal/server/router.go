package server

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/usernamesalah/rh-pos/internal/handler"
)

// CustomValidator wraps the validator
type CustomValidator struct {
	validator *validator.Validate
}

// Validate validates the struct
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// SetupRouter configures and returns the Echo router
func SetupRouter(
	authHandler *handler.AuthHandler,
	productHandler *handler.ProductHandler,
	transactionHandler *handler.TransactionHandler,
	reportHandler *handler.ReportHandler,
) *echo.Echo {
	e := echo.New()

	// Set custom validator
	e.Validator = &CustomValidator{validator: validator.New()}

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Swagger documentation
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	// Auth routes (no middleware)
	auth := e.Group("/auth")
	auth.POST("/login", authHandler.Login)

	// Protected routes
	api := e.Group("")
	api.Use(authHandler.AuthMiddleware())

	// Profile
	api.GET("/profile", authHandler.GetProfile)

	// Products
	products := api.Group("/products")
	products.GET("", productHandler.ListProducts)
	products.GET("/:id", productHandler.GetProduct)
	products.PUT("/:id", productHandler.UpdateProduct)
	products.PUT("/:id/stock", productHandler.UpdateStock)

	// Transactions
	transactions := api.Group("/transactions")
	transactions.GET("", transactionHandler.ListTransactions)
	transactions.POST("", transactionHandler.CreateTransaction)
	transactions.GET("/:id", transactionHandler.GetTransaction)

	// Reports
	reports := api.Group("/reports")
	reports.GET("", reportHandler.GetSalesReport)

	return e
}
