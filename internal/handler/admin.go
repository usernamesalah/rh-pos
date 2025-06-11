package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
)

type AdminHandler struct {
	tenantService interfaces.TenantService
	userService   interfaces.AuthService
}

func NewAdminHandler(tenantService interfaces.TenantService, userService interfaces.AuthService) *AdminHandler {
	return &AdminHandler{
		tenantService: tenantService,
		userService:   userService,
	}
}

// CreateTenant handles tenant creation
func (h *AdminHandler) CreateTenant(c echo.Context) error {
	var tenant entities.Tenant
	if err := c.Bind(&tenant); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if err := h.tenantService.CreateTenant(c.Request().Context(), &tenant); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, tenant)
}

// CreateUser handles user creation by admin
func (h *AdminHandler) CreateUser(c echo.Context) error {
	var user entities.User
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Hash the password before creating the user
	hashedPassword, err := h.userService.HashPassword(user.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to hash password"})
	}
	user.Password = hashedPassword

	// Create the user
	if err := h.userService.CreateUser(c.Request().Context(), &user); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, user)
}
