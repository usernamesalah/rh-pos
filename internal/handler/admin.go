package handler

import (
	"net/http"
	"strconv"

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

	// Create tenant
	if err := h.tenantService.CreateTenant(c.Request().Context(), &tenant); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, tenant)
}

// GetTenant handles getting tenant details
func (h *AdminHandler) GetTenant(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid tenant ID"})
	}

	tenant, err := h.tenantService.GetTenant(c.Request().Context(), uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, tenant)
}

// UpdateTenant handles tenant updates
func (h *AdminHandler) UpdateTenant(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid tenant ID"})
	}

	var tenant entities.Tenant
	if err := c.Bind(&tenant); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	tenant.ID = uint(id)

	// Update tenant
	if err := h.tenantService.UpdateTenant(c.Request().Context(), &tenant); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, tenant)
}

// ListTenants handles listing all tenants
func (h *AdminHandler) ListTenants(c echo.Context) error {
	tenants, err := h.tenantService.ListTenants(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, tenants)
}

// CreateUser handles user creation by admin
func (h *AdminHandler) CreateUser(c echo.Context) error {
	var user entities.User
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Validate tenant_id is provided in request
	if user.TenantID == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Tenant ID is required for user creation"})
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
