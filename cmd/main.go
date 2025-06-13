package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	if err := run(cfg, appLogger); err != nil {
		appLogger.Error("error: shutting down", "error", err)
		os.Exit(1)
	}
}

func run(cfg *config.Config, appLogger *slog.Logger) error {
	// Initialize database connection
	db, err := gorm.Open(mysql.Open(cfg.Database.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		appLogger.Error("Failed to connect to database", "error", err)
		return err
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

	// Create server with timeouts
	srv := &http.Server{
		Addr:         "0.0.0.0:" + port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		appLogger.Info("server listening", "port", port)
		serverErrors <- e.StartServer(srv)
	}()

	// Channel to listen for an interrupt or terminate signal from the OS.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return err

	case <-shutdown:
		appLogger.Info("caught signal, shutting down")

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Asking listener to shut down and shed load.
		if err := srv.Shutdown(ctx); err != nil {
			appLogger.Error("error: gracefully shutting down server", "error", err)
			if err := srv.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}
