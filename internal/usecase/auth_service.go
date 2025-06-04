package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
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
