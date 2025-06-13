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
	query := r.db.WithContext(ctx).Where("username = ?", username)

	// Add tenant_id filter if it exists in context
	if tenantID, ok := ctx.Value("tenant_id").(uint); ok {
		query = query.Where("tenant_id = ?", tenantID)
	}

	if err := query.First(&user).Error; err != nil {
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

	// Set tenant_id from context
	if tenantID, ok := ctx.Value("tenant_id").(uint); ok {
		user.TenantID = &tenantID
	}

	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to create user", "error", err, "username", user.Username)
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// Delete deletes a user
func (r *userRepository) Delete(ctx context.Context, id uint) error {
	r.logger.InfoContext(ctx, "deleting user", "id", id)
	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, ctx.Value("tenant_id")).Delete(&entities.User{}).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to delete user", "error", err, "id", id)
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// List retrieves all users
func (r *userRepository) List(ctx context.Context) ([]*entities.User, error) {
	r.logger.InfoContext(ctx, "listing users")

	var users []*entities.User
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", ctx.Value("tenant_id")).Find(&users).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to list users", "error", err)
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *entities.User) error {
	r.logger.InfoContext(ctx, "updating user", "id", user.ID)

	// Ensure tenant_id is set from context
	if tenantID, ok := ctx.Value("tenant_id").(uint); ok {
		user.TenantID = &tenantID
	}

	if err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", user.ID, ctx.Value("tenant_id")).Save(user).Error; err != nil {
		r.logger.ErrorContext(ctx, "failed to update user", "error", err, "id", user.ID)
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}
