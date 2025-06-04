package handler

import (
	"encoding/json"

	"github.com/labstack/echo/v4"
	"github.com/usernamesalah/rh-pos/internal/domain/entities"
	"github.com/usernamesalah/rh-pos/internal/pkg/hash"
)

// Response represents a standard API response
type Response struct {
	Status     string             `json:"status"`
	Message    string             `json:"message,omitempty"`
	Data       interface{}        `json:"data,omitempty"`
	Pagination *PaginatedResponse `json:"pagination,omitempty"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
}

// SuccessResponse returns a success response
func SuccessResponse(c echo.Context, code int, message string, data interface{}) error {
	return c.JSON(code, Response{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// SuccessPaginatedResponse returns a success response with pagination
func SuccessPaginatedResponse(c echo.Context, code int, message string, data interface{}, total int64, page, limit int) error {
	return c.JSON(code, Response{
		Status:  "success",
		Message: message,
		Data:    data,
		Pagination: &PaginatedResponse{
			Total: total,
			Page:  page,
			Limit: limit,
		},
	})
}

// ErrorResponse returns an error response
func ErrorResponse(c echo.Context, code int, message string) error {
	return c.JSON(code, Response{
		Status:  "error",
		Message: message,
	})
}

// HashIDResponse wraps the response data with hashed IDs
type HashIDResponse map[string]interface{}

// WithHashID wraps response data with a hashed ID
func WithHashID[T any](id uint, createdAt, updatedAt string, data T) HashIDResponse {
	response := HashIDResponse{
		"id":         hash.HashID(id),
		"created_at": createdAt,
		"updated_at": updatedAt,
	}

	// If data is a map, merge it with the response
	if dataMap, ok := any(data).(map[string]interface{}); ok {
		for k, v := range dataMap {
			response[k] = v
		}
	} else {
		// If data is a struct, convert it to map
		dataBytes, err := json.Marshal(data)
		if err == nil {
			var dataMap map[string]interface{}
			if err := json.Unmarshal(dataBytes, &dataMap); err == nil {
				for k, v := range dataMap {
					response[k] = v
				}
			}
		}
	}

	return response
}

// WithHashIDs wraps a slice of responses with hashed IDs
func WithHashIDs[T any](items []T, idExtractor func(T) uint, timeExtractor func(T) (string, string)) []HashIDResponse {
	result := make([]HashIDResponse, len(items))
	for i, item := range items {
		id := idExtractor(item)
		createdAt, updatedAt := timeExtractor(item)
		result[i] = WithHashID(id, createdAt, updatedAt, item)
	}
	return result
}

// GetUserFromContext retrieves the user from the echo context
func GetUserFromContext(c echo.Context) *entities.User {
	user, ok := c.Get("user").(*entities.User)
	if !ok {
		return nil
	}
	return user
}
