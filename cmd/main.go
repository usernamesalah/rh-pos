// @title RH POS API
// @version 1.0
// @description Point of Sale system API with multi-tenant support, product management, transactions, discount campaigns, warranty tracking, and sales reporting.

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey bearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token. Format: "Bearer {token}"

// @securityDefinitions.basic basicAuth
// @description Basic Auth for admin endpoints

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/usernamesalah/rh-pos/docs"
	"github.com/usernamesalah/rh-pos/internal/config"
	"github.com/usernamesalah/rh-pos/internal/handler"
	"github.com/usernamesalah/rh-pos/internal/pkg/hash"
	"github.com/usernamesalah/rh-pos/internal/pkg/storage"
	local "github.com/usernamesalah/rh-pos/internal/pkg/storage/local"
	s3client "github.com/usernamesalah/rh-pos/internal/pkg/storage/s3"
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

	// Initialize hash ID salt from config
	hash.Init(cfg.Hash.Salt)

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

	// Initialize storage client based on STORAGE_TYPE config
	storageClient, err := newStorageClient(cfg, appLogger)
	if err != nil {
		appLogger.Error("Failed to initialize storage client", "error", err)
		return err
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db, appLogger)
	productRepo := repository.NewProductRepository(db, appLogger)
	transactionRepo := repository.NewTransactionRepository(db, appLogger)
	tenantRepo := repository.NewTenantRepository(db, appLogger)
	campaignRepo := repository.NewDiscountCampaignRepository(db, appLogger)
	auditLogRepo := repository.NewAuditLogRepository(db, appLogger)
	categoryRepo := repository.NewCategoryRepository(db, appLogger)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	storageBaseURL := os.Getenv("STORAGE_BASE_URL")
	if storageBaseURL == "" {
		storageBaseURL = "http://localhost:" + port
	}

	// Initialize use cases
	authUseCase := usecase.NewAuthService(userRepo, cfg.JWT.Secret, appLogger)
	productUseCase := usecase.NewProductService(productRepo, storageClient, storageBaseURL, appLogger)
	transactionUseCase := usecase.NewTransactionService(transactionRepo, productRepo, campaignRepo, db, appLogger)
	reportUseCase := usecase.NewReportService(transactionRepo, appLogger)
	tenantUseCase := usecase.NewTenantService(tenantRepo, storageClient, storageBaseURL, appLogger)
	campaignUseCase := usecase.NewDiscountCampaignService(campaignRepo, auditLogRepo, appLogger)
	warrantyUseCase := usecase.NewWarrantyService(transactionRepo, appLogger)
	categoryUseCase := usecase.NewCategoryService(categoryRepo, appLogger)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUseCase, tenantUseCase, appLogger)
	productHandler := handler.NewProductHandler(productUseCase, appLogger)
	transactionHandler := handler.NewTransactionHandler(transactionUseCase, appLogger)
	reportHandler := handler.NewReportHandler(reportUseCase, appLogger)
	adminHandler := handler.NewAdminHandler(tenantUseCase, authUseCase)
	campaignHandler := handler.NewDiscountCampaignHandler(campaignUseCase, appLogger)
	warrantyHandler := handler.NewWarrantyHandler(warrantyUseCase, appLogger)
	categoryHandler := handler.NewCategoryHandler(categoryUseCase, appLogger)

	// Setup router
	e := server.SetupRouter(
		cfg,
		authHandler,
		productHandler,
		transactionHandler,
		reportHandler,
		adminHandler,
		campaignHandler,
		warrantyHandler,
		categoryHandler,
	)

	// Start server
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

func newStorageClient(cfg *config.Config, logger *slog.Logger) (storage.StorageClient, error) {
	switch cfg.Storage.Type {
	case "local":
		logger.Info("using local filesystem storage", "path", cfg.Storage.LocalPath)
		return local.NewClient(cfg.Storage.LocalPath, logger)
	case "minio":
		logger.Info("using S3-compatible storage (deprecated, use STORAGE_TYPE=s3)", "endpoint", cfg.S3.Endpoint)
		fallthrough
	case "s3":
		logger.Info("using S3-compatible storage", "endpoint", cfg.S3.Endpoint)
		s3Config := &s3client.Config{
			Endpoint:        cfg.S3.Endpoint,
			AccessKeyID:     cfg.S3.AccessKeyID,
			SecretAccessKey: cfg.S3.SecretAccessKey,
			UseSSL:          cfg.S3.UseSSL,
			Region:          cfg.S3.Region,
			Bucket:          cfg.S3.Bucket,
			DefaultExpiry:   cfg.S3.DefaultExpiry,
		}
		return s3client.NewClient(s3Config, logger)
	default:
		return nil, fmt.Errorf("unknown STORAGE_TYPE %q: must be 'local' or 's3'", cfg.Storage.Type)
	}
}
