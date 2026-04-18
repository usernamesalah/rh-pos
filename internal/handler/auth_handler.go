package handler

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/usernamesalah/rh-pos/internal/domain/apperrors"
	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"github.com/usernamesalah/rh-pos/internal/pkg/ctxkey"
	"github.com/usernamesalah/rh-pos/internal/pkg/hash"
	"gorm.io/gorm"
)

type AuthHandler struct {
	authService   interfaces.AuthService
	tenantService interfaces.TenantService
	logger        *slog.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService interfaces.AuthService, tenantService interfaces.TenantService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		authService:   authService,
		tenantService: tenantService,
		logger:        logger,
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

// UpdatePasswordRequest represents the update password request payload
type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
}

type UpdateMyTenantRequest struct {
	Name           string `json:"name"`
	About          string `json:"about"`
	Address        string `json:"address"`
	PhoneNumber    string `json:"phone_number"`
	TermsOfService string `json:"terms_of_service"`
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
	userIDRaw, ok := c.Get("user_id").(uint)
	if !ok {
		return ErrorResponse(c, http.StatusUnauthorized, "Invalid token claims")
	}
	user, err := h.authService.GetUserByID(c.Request().Context(), userIDRaw)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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

// GetMyTenant handles getting current user's tenant information
// @Summary Get user's tenant information
// @Description Get current user's tenant details from JWT token
// @Tags Authentication
// @Produce json
// @Security bearerAuth
// @Success 200 {object} Response{data=HashIDResponse}
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Router /api/my-tenant [get]
func (h *AuthHandler) GetMyTenant(c echo.Context) error {
	// Get tenant_id from context (set by JWT middleware)
	tenantID, ok := c.Get("tenant_id").(uint)
	if !ok {
		return ErrorResponse(c, http.StatusUnauthorized, "Tenant information not available")
	}

	tenant, err := h.tenantService.GetTenant(c.Request().Context(), tenantID)
	if err != nil {
		h.logger.ErrorContext(c.Request().Context(), "failed to get tenant", "error", err, "tenant_id", tenantID)
		return ErrorResponse(c, http.StatusNotFound, "Tenant not found")
	}

	logoURL, _ := h.tenantService.GetTenantLogoURL(c.Request().Context(), tenant)

	response := WithHashID(
		tenant.ID,
		tenant.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		tenant.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		map[string]interface{}{
			"name":            tenant.Name,
			"about":           tenant.About,
			"address":         tenant.Address,
			"phone_number":   tenant.PhoneNumber,
			"logo_url":        logoURL,
			"terms_of_service": tenant.TermsOfService,
		},
	)

	return SuccessResponse(c, http.StatusOK, "Tenant information retrieved successfully", response)
}

// UpdateMyTenant handles updating current user's tenant information
// @Summary Update user's tenant information
// @Description Update current tenant details for tenant admin
// @Tags Authentication
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param request body UpdateMyTenantRequest true "Tenant update request"
// @Success 200 {object} Response{data=HashIDResponse}
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 404 {object} Response
// @Router /api/my-tenant [put]
func (h *AuthHandler) UpdateMyTenant(c echo.Context) error {
	tenantID, ok := c.Get("tenant_id").(uint)
	if !ok {
		return ErrorResponse(c, http.StatusUnauthorized, "Tenant information not available")
	}

	tenant, err := h.tenantService.GetTenant(c.Request().Context(), tenantID)
	if err != nil {
		h.logger.ErrorContext(c.Request().Context(), "failed to get tenant before update", "error", err, "tenant_id", tenantID)
		return ErrorResponse(c, http.StatusNotFound, "Tenant not found")
	}

	var req UpdateMyTenantRequest
	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	tenant.Name = req.Name
	tenant.About = req.About
	tenant.Address = req.Address
	tenant.PhoneNumber = req.PhoneNumber
	tenant.TermsOfService = req.TermsOfService

	if err := h.tenantService.UpdateTenant(c.Request().Context(), tenant); err != nil {
		h.logger.ErrorContext(c.Request().Context(), "failed to update tenant", "error", err, "tenant_id", tenantID)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to update tenant")
	}

	logoURL, _ := h.tenantService.GetTenantLogoURL(c.Request().Context(), tenant)

	response := WithHashID(
		tenant.ID,
		tenant.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		tenant.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		map[string]interface{}{
			"name":            tenant.Name,
			"about":           tenant.About,
			"address":         tenant.Address,
			"phone_number":   tenant.PhoneNumber,
			"logo_url":        logoURL,
			"terms_of_service": tenant.TermsOfService,
		},
	)

	return SuccessResponse(c, http.StatusOK, "Tenant updated successfully", response)
}

// UploadMyTenantLogo handles uploading a logo for the current tenant
// @Summary Upload tenant logo
// @Description Upload a logo image for the current tenant (admin only)
// @Tags Authentication
// @Accept multipart/form-data
// @Produce json
// @Security bearerAuth
// @Param file formData file true "Logo file"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Failure 500 {object} Response
// @Router /api/my-tenant/logo [post]
func (h *AuthHandler) UploadMyTenantLogo(c echo.Context) error {
	tenantID, ok := c.Get("tenant_id").(uint)
	if !ok {
		return ErrorResponse(c, http.StatusUnauthorized, "Tenant information not available")
	}

	if err := c.Request().ParseMultipartForm(32 << 20); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Failed to parse form data")
	}

	form := c.Request().MultipartForm
	files, ok := form.File["file"]
	if !ok || len(files) == 0 {
		return ErrorResponse(c, http.StatusBadRequest, "Logo file is required")
	}

	file := files[0]
	src, err := file.Open()
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to process uploaded file")
	}
	defer src.Close()

	fileData, err := io.ReadAll(src)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to read uploaded file")
	}

	ctx := context.WithValue(c.Request().Context(), ctxkey.TenantID, tenantID)
	tenant, err := h.tenantService.UploadTenantLogo(ctx, tenantID, fileData, file.Header.Get("Content-Type"))
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to upload logo")
	}

	logoURL, _ := h.tenantService.GetTenantLogoURL(c.Request().Context(), tenant)

	response := map[string]interface{}{
		"name":            tenant.Name,
		"about":           tenant.About,
		"address":         tenant.Address,
		"phone_number":   tenant.PhoneNumber,
		"logo_url":        logoURL,
		"terms_of_service": tenant.TermsOfService,
	}

	return SuccessResponse(c, http.StatusOK, "Logo uploaded successfully", response)
}

// UpdatePassword handles password update for the current user
// @Summary Update user password
// @Description Update current user's password with current password verification
// @Tags Authentication
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param request body UpdatePasswordRequest true "Password update request"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Router /api/update-password [put]
func (h *AuthHandler) UpdatePassword(c echo.Context) error {
	var req UpdatePasswordRequest
	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Validation failed")
	}

	// Get user ID from context (set by JWT middleware)
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return ErrorResponse(c, http.StatusUnauthorized, "Invalid token claims")
	}

	// Update password
	if err := h.authService.UpdatePassword(c.Request().Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		h.logger.ErrorContext(c.Request().Context(), "failed to update password", "error", err, "user_id", userID)

		if errors.Is(err, apperrors.ErrInvalidPassword) {
			return ErrorResponse(c, http.StatusUnauthorized, "Invalid current password")
		}

		return ErrorResponse(c, http.StatusBadRequest, "Failed to update password")
	}

	return SuccessResponse(c, http.StatusOK, "Password updated successfully", nil)
}

// ListUsers handles listing users with pagination
// @Summary List users
// @Description Get all users for the current tenant with pagination
// @Tags Users
// @Produce json
// @Security bearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Items per page (max 100)"
// @Success 200 {object} Response
// @Failure 401 {object} Response
// @Router /api/users [get]
func (h *AuthHandler) ListUsers(c echo.Context) error {
	ctx := c.Request().Context()

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	users, total, err := h.authService.ListUsers(ctx, page, limit)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to list users", "error", err)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to list users")
	}

	items := make([]HashIDResponse, len(users))
	for i, u := range users {
		items[i] = WithHashID(
			u.ID,
			u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			map[string]interface{}{
				"username": u.Username,
				"role":     u.Role,
			},
		)
	}

	return SuccessPaginatedResponse(
		c,
		http.StatusOK,
		"Users retrieved successfully",
		items,
		total,
		page,
		limit,
	)
}

type CreateUserRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role"`
}

type UpdateUserRequest struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

// CreateUser handles creating a new user
// @Summary Create user
// @Description Create a new user for the current tenant
// @Tags Users
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param request body CreateUserRequest true "User data"
// @Success 201 {object} Response
// @Failure 400 {object} Response
// @Failure 401 {object} Response
// @Router /api/users [post]
func (h *AuthHandler) CreateUser(c echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Validation failed")
	}

	role := req.Role
	if role == "" {
		role = "cashier"
	}

	user := &entities.User{
		Username: req.Username,
		Password: req.Password,
		Role:     role,
	}

	if err := h.authService.CreateUser(c.Request().Context(), user); err != nil {
		h.logger.ErrorContext(c.Request().Context(), "failed to create user", "error", err, "username", req.Username)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to create user")
	}

	response := WithHashID(
		user.ID,
		user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		map[string]interface{}{
			"username": user.Username,
			"role":     user.Role,
		},
	)

	return SuccessResponse(c, http.StatusCreated, "User created successfully", response)
}

// GetUser handles getting a user by ID
// @Summary Get user by ID
// @Description Get a specific user by ID
// @Tags Users
// @Produce json
// @Security bearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /api/users/{id} [get]
func (h *AuthHandler) GetUser(c echo.Context) error {
	ctx := c.Request().Context()

	hashedID := c.Param("id")
	id, err := hash.DecodeHashID(hashedID)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid user ID format", "error", err, "hashed_id", hashedID)
		return ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format")
	}

	user, err := h.authService.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "User not found")
		}
		h.logger.ErrorContext(ctx, "failed to get user", "error", err, "id", id)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to get user")
	}

	response := WithHashID(
		user.ID,
		user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		map[string]interface{}{
			"username": user.Username,
			"role":     user.Role,
		},
	)

	return SuccessResponse(c, http.StatusOK, "User retrieved successfully", response)
}

// UpdateUser handles updating a user
// @Summary Update user
// @Description Update an existing user
// @Tags Users
// @Accept json
// @Produce json
// @Security bearerAuth
// @Param id path string true "User ID"
// @Param request body UpdateUserRequest true "User data"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /api/users/{id} [put]
func (h *AuthHandler) UpdateUser(c echo.Context) error {
	ctx := c.Request().Context()

	hashedID := c.Param("id")
	id, err := hash.DecodeHashID(hashedID)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid user ID format", "error", err, "hashed_id", hashedID)
		return ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format")
	}

	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	user, err := h.authService.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrorResponse(c, http.StatusNotFound, "User not found")
		}
		h.logger.ErrorContext(ctx, "failed to get user", "error", err, "id", id)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to get user")
	}

	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Role != "" {
		user.Role = req.Role
	}

	if err := h.authService.UpdateUser(ctx, user); err != nil {
		h.logger.ErrorContext(ctx, "failed to update user", "error", err, "id", id)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to update user")
	}

	response := WithHashID(
		user.ID,
		user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		map[string]interface{}{
			"username": user.Username,
			"role":     user.Role,
		},
	)

	return SuccessResponse(c, http.StatusOK, "User updated successfully", response)
}

// DeleteUser handles deleting a user
// @Summary Delete user
// @Description Delete a user by ID
// @Tags Users
// @Produce json
// @Security bearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Router /api/users/{id} [delete]
func (h *AuthHandler) DeleteUser(c echo.Context) error {
	ctx := c.Request().Context()

	hashedID := c.Param("id")
	id, err := hash.DecodeHashID(hashedID)
	if err != nil {
		h.logger.WarnContext(ctx, "invalid user ID format", "error", err, "hashed_id", hashedID)
		return ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format")
	}

	if err := h.authService.DeleteUser(ctx, id); err != nil {
		h.logger.ErrorContext(ctx, "failed to delete user", "error", err, "id", id)
		return ErrorResponse(c, http.StatusInternalServerError, "Failed to delete user")
	}

	return SuccessResponse(c, http.StatusOK, "User deleted successfully", nil)
}
