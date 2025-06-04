package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/usernamesalah/rh-pos/internal/domain/entities"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// GetUserFromContext retrieves the user from the echo context
func GetUserFromContext(c echo.Context) *entities.User {
	user, ok := c.Get("user").(*entities.User)
	if !ok {
		return nil
	}
	return user
}
