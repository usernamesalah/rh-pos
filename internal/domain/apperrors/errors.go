package apperrors

import "errors"

// Sentinel errors for use with errors.Is across service/handler boundaries.
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidPassword   = errors.New("invalid current password")
	ErrTenantNotInContext = errors.New("tenant_id not found in context")
)
