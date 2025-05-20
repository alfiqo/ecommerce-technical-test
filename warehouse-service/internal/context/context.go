package context

import (
	"context"
	"time"
)

// Keys for context values
type contextKey string

const (
	// RequestIDKey is the key for request ID in context
	RequestIDKey contextKey = "request_id"
	
	// UserIDKey is the key for user ID in context
	UserIDKey contextKey = "user_id"
	
	// TimeoutKey is the key for operation timeout
	TimeoutKey contextKey = "timeout"
)

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID retrieves request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetUserID retrieves user ID from context
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}

// WithTimeout returns a context that times out after the given duration
func WithTimeout(ctx context.Context, duration time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, duration)
}

// WithDefaultTimeout returns a context with the default timeout of 30 seconds
func WithDefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return WithTimeout(ctx, 30*time.Second)
}