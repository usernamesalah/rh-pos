package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/usernamesalah/rh-pos/internal/config"
	"github.com/usernamesalah/rh-pos/internal/handler"
	"github.com/usernamesalah/rh-pos/internal/repository"
	"github.com/usernamesalah/rh-pos/internal/server"
	"github.com/usernamesalah/rh-pos/internal/usecase"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	appLogger := slog.New(logHandler)

	// Initialize database connection
	db, err := gorm.Open(mysql.Open(cfg.Database.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		appLogger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db, appLogger)
	productRepo := repository.NewProductRepository(db, appLogger)
	transactionRepo := repository.NewTransactionRepository(db, appLogger)
	tenantRepo := repository.NewTenantRepository(db, appLogger)

	// Initialize use cases
	authUseCase := usecase.NewAuthService(userRepo, cfg.JWT.Secret, appLogger)
	productUseCase := usecase.NewProductService(productRepo, appLogger)
	transactionUseCase := usecase.NewTransactionService(transactionRepo, productRepo, db, appLogger)
	reportUseCase := usecase.NewReportService(transactionRepo, appLogger)
	tenantUseCase := usecase.NewTenantService(tenantRepo, appLogger)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUseCase, appLogger)
	productHandler := handler.NewProductHandler(productUseCase, appLogger)
	transactionHandler := handler.NewTransactionHandler(transactionUseCase, appLogger)
	reportHandler := handler.NewReportHandler(reportUseCase, appLogger)
	adminHandler := handler.NewAdminHandler(tenantUseCase, authUseCase)

	// Setup router
	e := server.SetupRouter(
		cfg,
		authHandler,
		productHandler,
		transactionHandler,
		reportHandler,
		adminHandler,
	)

	// Start server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}
