package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/usernamesalah/rh-pos/internal/config"
)

// AdminAuth is a middleware that checks for Basic Auth credentials
func AdminAuth(cfg *config.Config) echo.MiddlewareFunc {
	return middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		return username == cfg.Admin.Username && password == cfg.Admin.Password, nil
	})
}
