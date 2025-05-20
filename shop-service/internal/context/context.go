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

// WithTimeout returns a context that times out after the given duration
func WithTimeout(ctx context.Context, duration time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, duration)
}

// WithDefaultTimeout returns a context with the default timeout of 5 seconds
func WithDefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return WithTimeout(ctx, 5*time.Second)
}