package middleware

import (
	"crypto/subtle"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/usernamesalah/rh-pos/internal/config"
)

// AdminAuth is a middleware that checks for Basic Auth credentials.
// Uses constant-time comparison to prevent timing attacks.
func AdminAuth(cfg *config.Config) echo.MiddlewareFunc {
	return middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(cfg.Admin.Username))
		passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(cfg.Admin.Password))
		if usernameMatch != 1 || passwordMatch != 1 {
			return false, nil
		}
		return true, nil
	})
}
