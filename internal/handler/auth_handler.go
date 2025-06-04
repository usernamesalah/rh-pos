package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
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
// @Success 200 {object} LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()

	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid request body", "error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
	}

	if err := c.Validate(req); err != nil {
		h.logger.WarnContext(ctx, "validation failed", "error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Validation failed"})
	}

	token, user, err := h.authService.Login(ctx, req.Username, req.Password)
	if err != nil {
		h.logger.WarnContext(ctx, "login failed", "error", err, "username", req.Username)
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid credentials"})
	}

	response := LoginResponse{
		Token:    token,
		Username: user.Username,
		Role:     user.Role,
	}

	return c.JSON(http.StatusOK, response)
}

// GetProfile handles getting user profile
// @Summary Get user profile
// @Description Get current user profile information
// @Tags Authentication
// @Produce json
// @Security bearerAuth
// @Success 200 {object} ProfileResponse
// @Failure 401 {object} ErrorResponse
// @Router /profile [get]
func (h *AuthHandler) GetProfile(c echo.Context) error {
	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Unauthorized"})
	}

	response := ProfileResponse{
		Username: user.Username,
		Role:     user.Role,
	}

	return c.JSON(http.StatusOK, response)
}

// AuthMiddleware validates JWT tokens
func (h *AuthHandler) AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Authorization header required"})
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Bearer token required"})
			}

			user, err := h.authService.ValidateToken(tokenString)
			if err != nil {
				h.logger.WarnContext(c.Request().Context(), "invalid token", "error", err)
				return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid token"})
			}

			// Store user in context
			c.Set("user", user)
			return next(c)
		}
	}
}
