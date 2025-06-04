package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"gorm.io/gorm"
)

type userRepository struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB, logger *slog.Logger) interfaces.UserRepository {
	return &userRepository{
		db:     db,
		logger: logger,
	}
}

// GetByUsername retrieves a user by username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	r.logger.InfoContext(ctx, "getting user by username", "username", username)

	var user entities.User
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		r.logger.ErrorContext(ctx, "failed to get user", "error", err, "username", username)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uint) (*entities.User, error) {
	r.logger.InfoContext(ctx, "getting user by ID", "id", id)

	var user entities.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		r.logger.ErrorContext(ctx, "failed to get user", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *entities.User) error {
	r.logger.InfoContext(ctx, "creating user", "username", user.Username)

	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to create user", "error", err, "username", user.Username)
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}
