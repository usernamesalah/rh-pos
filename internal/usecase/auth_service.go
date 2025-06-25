package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"github.com/usernamesalah/rh-pos/internal/pkg/hash"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo  interfaces.UserRepository
	jwtSecret string
	logger    *slog.Logger
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo interfaces.UserRepository, jwtSecret string, logger *slog.Logger) interfaces.AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		logger:    logger,
	}
}

// Login authenticates a user and returns a JWT token
func (s *authService) Login(ctx context.Context, username, password string) (string, *entities.User, error) {
	s.logger.InfoContext(ctx, "attempting login", "username", username)

	// Get user by username
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		s.logger.WarnContext(ctx, "login failed: user not found", "username", username)
		return "", nil, fmt.Errorf("invalid credentials")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		s.logger.WarnContext(ctx, "login failed: invalid password", "username", username)
		return "", nil, fmt.Errorf("invalid credentials")
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	// Add tenant_id to claims if it exists
	if user.TenantID != nil {
		// Hash the tenant_id before adding to claims
		hashedTenantID := hash.HashID(*user.TenantID)
		token.Claims.(jwt.MapClaims)["tenant_id"] = hashedTenantID
	}

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to generate token", "error", err, "username", username)
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	s.logger.InfoContext(ctx, "login successful", "username", username)
	return tokenString, user, nil
}

// ValidateToken validates a JWT token and returns the user
func (s *authService) ValidateToken(tokenString string) (*entities.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		user := &entities.User{
			ID:       uint(claims["user_id"].(float64)),
			Username: claims["username"].(string),
			Role:     claims["role"].(string),
		}
		if tenantID, ok := claims["tenant_id"].(string); ok {
			// Decode the hashed tenant ID
			decodedTenantID, err := hash.DecodeHashID(tenantID)
			if err == nil {
				user.TenantID = &decodedTenantID
			}
		}
		return user, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// HashPassword hashes a password using bcrypt
func (s *authService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// GetUserByID retrieves a user by their ID
func (s *authService) GetUserByID(ctx context.Context, id uint) (*entities.User, error) {
	s.logger.InfoContext(ctx, "getting user by ID", "id", id)
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user by ID", "error", err, "id", id)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// CreateUser creates a new user
func (s *authService) CreateUser(ctx context.Context, user *entities.User) error {
	s.logger.InfoContext(ctx, "creating user", "username", user.Username)

	// Hash password
	hashedPassword, err := s.HashPassword(user.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = hashedPassword

	// Create user
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.ErrorContext(ctx, "failed to create user", "error", err, "username", user.Username)
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// UpdatePassword updates a user's password
func (s *authService) UpdatePassword(ctx context.Context, userID uint, currentPassword, newPassword string) error {
	s.logger.InfoContext(ctx, "updating password", "user_id", userID)

	// Get user by ID
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get user for password update", "error", err, "user_id", userID)
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		s.logger.WarnContext(ctx, "password update failed: invalid current password", "user_id", userID)
		return fmt.Errorf("invalid current password")
	}

	// Hash new password
	hashedNewPassword, err := s.HashPassword(newPassword)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to hash new password", "error", err, "user_id", userID)
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update user password
	user.Password = hashedNewPassword
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.ErrorContext(ctx, "failed to update user password", "error", err, "user_id", userID)
		return fmt.Errorf("failed to update password: %w", err)
	}

	s.logger.InfoContext(ctx, "password updated successfully", "user_id", userID)
	return nil
}
