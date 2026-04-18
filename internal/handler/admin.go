package handler

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/domain/interfaces"
	"github.com/usernamesalah/rh-pos/internal/pkg/ctxkey"
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
// @Summary Create a new tenant
// @Description Create a new tenant
// @Tags Admin
// @Accept json
// @Produce json
// @Security basicAuth
// @Param request body entities.Tenant true "Tenant data"
// @Success 201 {object} entities.Tenant
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/tenants [post]
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
// @Summary Get a tenant by ID
// @Description Get tenant details by ID
// @Tags Admin
// @Accept json
// @Produce json
// @Security basicAuth
// @Param id path int true "Tenant ID"
// @Success 200 {object} entities.Tenant
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /admin/tenants/{id} [get]
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

	ctx := context.WithValue(c.Request().Context(), ctxkey.TenantID, tenant.ID)
	logoURL, _ := h.tenantService.GetTenantLogoURL(ctx, tenant)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":              tenant.ID,
		"name":            tenant.Name,
		"about":           tenant.About,
		"address":         tenant.Address,
		"phone_number":   tenant.PhoneNumber,
		"logo_url":        logoURL,
		"terms_of_service": tenant.TermsOfService,
		"created_at":      tenant.CreatedAt,
		"updated_at":      tenant.UpdatedAt,
	})
}

// UpdateTenant handles tenant updates
// @Summary Update a tenant
// @Description Update an existing tenant by ID
// @Tags Admin
// @Accept json
// @Produce json
// @Security basicAuth
// @Param id path int true "Tenant ID"
// @Param request body entities.Tenant true "Tenant data"
// @Success 200 {object} entities.Tenant
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/tenants/{id} [put]
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
// @Summary List all tenants
// @Description List all tenants
// @Tags Admin
// @Accept json
// @Produce json
// @Security basicAuth
// @Success 200 {array} entities.Tenant
// @Failure 500 {object} map[string]string
// @Router /admin/tenants [get]
func (h *AdminHandler) ListTenants(c echo.Context) error {
	tenants, err := h.tenantService.ListTenants(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	result := make([]map[string]interface{}, 0, len(tenants))
	for _, tenant := range tenants {
		ctx := context.WithValue(c.Request().Context(), ctxkey.TenantID, tenant.ID)
		logoURL, _ := h.tenantService.GetTenantLogoURL(ctx, tenant)
		result = append(result, map[string]interface{}{
			"id":                tenant.ID,
			"name":              tenant.Name,
			"about":             tenant.About,
			"address":           tenant.Address,
			"phone_number":     tenant.PhoneNumber,
			"logo_url":          logoURL,
			"terms_of_service":  tenant.TermsOfService,
			"created_at":        tenant.CreatedAt,
			"updated_at":        tenant.UpdatedAt,
		})
	}

	return c.JSON(http.StatusOK, result)
}

type adminCreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
	TenantID *uint  `json:"tenant_id"`
}

// CreateUser handles user creation by admin
// @Summary Create a new user
// @Description Create a new user assigned to a tenant
// @Tags Admin
// @Accept json
// @Produce json
// @Security basicAuth
// @Param request body adminCreateUserRequest true "User data"
// @Success 201 {object} entities.User
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/users [post]
func (h *AdminHandler) CreateUser(c echo.Context) error {
	var req adminCreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.TenantID == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Tenant ID is required for user creation"})
	}

	user := entities.User{
		Username: req.Username,
		Password: req.Password,
		Role:     req.Role,
		TenantID: req.TenantID,
	}

	if err := h.userService.CreateUser(c.Request().Context(), &user); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, user)
}

// UpdateTenantLogo handles uploading a logo for a tenant
// @Summary Upload tenant logo
// @Description Upload a logo image for a tenant
// @Tags Admin
// @Accept multipart/form-data
// @Produce json
// @Security basicAuth
// @Param id path int true "Tenant ID"
// @Param file formData file true "Logo file"
// @Success 200 {object} entities.Tenant
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/tenants/{id}/logo [post]
func (h *AdminHandler) UpdateTenantLogo(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid tenant ID"})
	}

	if err := c.Request().ParseMultipartForm(32 << 20); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Failed to parse form data"})
	}

	form := c.Request().MultipartForm
	files, ok := form.File["file"]
	if !ok || len(files) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Logo file is required"})
	}

	file := files[0]
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to process uploaded file"})
	}
	defer src.Close()

	fileData, err := io.ReadAll(src)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to read uploaded file"})
	}

	ctx := context.WithValue(c.Request().Context(), ctxkey.TenantID, uint(id))
	tenant, err := h.tenantService.UploadTenantLogo(ctx, uint(id), fileData, file.Header.Get("Content-Type"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	logoURL, _ := h.tenantService.GetTenantLogoURL(c.Request().Context(), tenant)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":              tenant.ID,
		"name":            tenant.Name,
		"about":           tenant.About,
		"address":         tenant.Address,
		"phone_number":   tenant.PhoneNumber,
		"logo_url":        logoURL,
		"terms_of_service": tenant.TermsOfService,
		"created_at":      tenant.CreatedAt,
		"updated_at":      tenant.UpdatedAt,
	})
}
