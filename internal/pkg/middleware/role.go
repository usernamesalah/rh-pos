package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/usernamesalah/rh-pos/internal/domain/entities"
)

// RequireRole returns an Echo middleware that rejects requests from users
// whose role is not in the allowed list.
func RequireRole(roles ...string) echo.MiddlewareFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, ok := c.Get("role").(string)
			if !ok || role == "" {
				return echo.NewHTTPError(http.StatusForbidden, "access denied")
			}
			if _, permitted := allowed[role]; !permitted {
				return echo.NewHTTPError(http.StatusForbidden, "access denied")
			}
			return next(c)
		}
	}
}

// AdminOnly is a convenience middleware that allows only admin role.
var AdminOnly = RequireRole(entities.RoleAdmin)
