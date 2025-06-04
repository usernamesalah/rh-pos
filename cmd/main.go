package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/usernamesalah/rh-pos/internal/config"
	"github.com/usernamesalah/rh-pos/internal/handler"
	"github.com/usernamesalah/rh-pos/internal/pkg/database"
	"github.com/usernamesalah/rh-pos/internal/repository"
	"github.com/usernamesalah/rh-pos/internal/server"
	"github.com/usernamesalah/rh-pos/internal/usecase"
)

// @title POS System API
// @version 1.0
// @description API for Point of Sale System with user management, product management, and sales reporting
// @host localhost:8080
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey bearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Load configuration from .env file
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("starting application", "config", map[string]interface{}{
		"server_host": cfg.Server.Host,
		"server_port": cfg.Server.Port,
		"db_host":     cfg.Database.Host,
		"db_name":     cfg.Database.Name,
		"log_level":   cfg.Logger.Level,
	})

	// Connect to database
	db, err := database.NewConnection(cfg.Database.DSN, logger)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		panic(err)
	}

	// Run migrations
	if err := database.AutoMigrate(db, logger); err != nil {
		logger.Error("failed to run migrations", "error", err)
		panic(err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db, logger)
	productRepo := repository.NewProductRepository(db, logger)
	transactionRepo := repository.NewTransactionRepository(db, logger)

	// Initialize services (pass db to transaction service for transaction support)
	authService := usecase.NewAuthService(userRepo, cfg.JWT.Secret, logger)
	productService := usecase.NewProductService(productRepo, logger)
	transactionService := usecase.NewTransactionService(transactionRepo, productRepo, db, logger)
	reportService := usecase.NewReportService(transactionRepo, logger)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, logger)
	productHandler := handler.NewProductHandler(productService, logger)
	transactionHandler := handler.NewTransactionHandler(transactionService, logger)
	reportHandler := handler.NewReportHandler(reportService, logger)

	// Setup router
	e := server.SetupRouter(cfg, authHandler, productHandler, transactionHandler, reportHandler)

	// Start server
	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	logger.Info("starting server", "address", serverAddr)

	// Start server in a goroutine
	go func() {
		if err := e.Start(serverAddr); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to start server", "error", err)
			panic(err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	logger.Info("shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		panic(err)
	}

	logger.Info("server exited")
}
