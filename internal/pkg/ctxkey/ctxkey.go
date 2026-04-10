package ctxkey

import "context"

type contextKey string

// TenantID is the context key for tenant_id.
const TenantID contextKey = "tenant_id"

// UserID is the context key for user_id.
const UserID contextKey = "user_id"

// WithTenantID returns a new context with the given tenant ID stored under TenantID.
func WithTenantID(ctx context.Context, tenantID uint) context.Context {
	return context.WithValue(ctx, TenantID, tenantID)
}

// TenantIDFromContext extracts the tenant_id from context.
// Returns (0, false) if the key is absent or has the wrong type.
func TenantIDFromContext(ctx context.Context) (uint, bool) {
	v, ok := ctx.Value(TenantID).(uint)
	return v, ok
}

// WithUserID returns a new context with the given user ID stored under UserID.
func WithUserID(ctx context.Context, userID uint) context.Context {
	return context.WithValue(ctx, UserID, userID)
}

// UserIDFromContext extracts the user_id from context.
// Returns (0, false) if the key is absent or has the wrong type.
func UserIDFromContext(ctx context.Context) (uint, bool) {
	v, ok := ctx.Value(UserID).(uint)
	return v, ok
}
