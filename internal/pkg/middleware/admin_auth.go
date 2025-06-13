package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/usernamesalah/rh-pos/internal/config"
)

// AdminAuth is a middleware that checks for Basic Auth credentials and sets tenant_id
func AdminAuth(cfg *config.Config) echo.MiddlewareFunc {
	return middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		// Check admin credentials
		if username != cfg.Admin.Username || password != cfg.Admin.Password {
			return false, nil
		}

		// Set tenant_id to 0 for admin operations (super admin)
		c.Set("tenant_id", uint(0))
		return true, nil
	})
}
