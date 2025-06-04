package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"gorm.io/gorm"
)

type AuthHandler struct {
	authService interfaces.AuthService
	logger      *slog.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService interfaces.AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the login response payload
type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// ProfileResponse represents the profile response payload
type ProfileResponse struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

// Login handles user authentication
// @Summary Login to the system
// @Description Authenticate user with username and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} Response{data=HashIDResponse}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Router /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	token, user, err := h.authService.Login(c.Request().Context(), req.Username, req.Password)
	if err != nil {
		return ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials")
	}

	response := WithHashID(
		user.ID,
		user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		map[string]interface{}{
			"token":    token,
			"username": user.Username,
			"role":     user.Role,
		},
	)

	return SuccessResponse(c, http.StatusOK, "Login successful", response)
}

// GetProfile handles getting user profile
// @Summary Get user profile
// @Description Get current user profile information
// @Tags Authentication
// @Produce json
// @Security bearerAuth
// @Success 200 {object} Response{data=HashIDResponse}
// @Failure 401 {object} Response
// @Router /api/profile [get]
func (h *AuthHandler) GetProfile(c echo.Context) error {
	userID := c.Get("user_id").(uint)
	user, err := h.authService.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrorResponse(c, http.StatusNotFound, "User not found")
		}
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to get profile")
	}

	response := WithHashID(
		user.ID,
		user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		user,
	)

	return SuccessResponse(c, http.StatusOK, "Profile retrieved successfully", response)
}

// AuthMiddleware validates JWT tokens
func (h *AuthHandler) AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return ErrorResponse(c, http.StatusUnauthorized, "Authorization header required")
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return ErrorResponse(c, http.StatusUnauthorized, "Bearer token required")
			}

			user, err := h.authService.ValidateToken(tokenString)
			if err != nil {
				h.logger.WarnContext(c.Request().Context(), "invalid token", "error", err)
				return ErrorResponse(c, http.StatusUnauthorized, "Invalid token")
			}

			// Store user in context
			c.Set("user", user)
			return next(c)
		}
	}
}
