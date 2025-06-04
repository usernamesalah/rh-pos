package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/usernamesalah/rh-pos/internal/config"
	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/pkg/database"
	"github.com/usernamesalah/rh-pos/internal/repository"
	"github.com/usernamesalah/rh-pos/internal/usecase"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Connect to database
	db, err := database.NewConnection(cfg.Database.DSN, logger)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		panic(err)
	}

	// Initialize repositories and services
	userRepo := repository.NewUserRepository(db, logger)
	authService := usecase.NewAuthService(userRepo, cfg.JWT.Secret, logger)

	// Create admin user
	hashedPassword, err := authService.HashPassword("admin123")
	if err != nil {
		logger.Error("failed to hash password", "error", err)
		panic(err)
	}

	adminUser := &entities.User{
		Username: "admin",
		Password: hashedPassword,
		Role:     "admin",
	}

	ctx := context.Background()
	if err := userRepo.Create(ctx, adminUser); err != nil {
		logger.Error("failed to create admin user", "error", err)
		// Don't panic here as user might already exist
	} else {
		logger.Info("admin user created successfully", "username", adminUser.Username)
	}

	fmt.Println("Seeding completed!")
	fmt.Println("Admin credentials:")
	fmt.Println("Username: admin")
	fmt.Println("Password: admin123")
}
