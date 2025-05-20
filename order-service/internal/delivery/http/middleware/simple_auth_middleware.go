package middleware

import (
	"order-service/internal/context"
	"order-service/internal/delivery/http/response"
	appErrors "order-service/internal/errors"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// SimpleAuthMiddleware provides basic authentication functionality
type SimpleAuthMiddleware struct {
	Log *logrus.Logger
}

// NewSimpleAuthMiddleware creates a new authentication middleware
func NewSimpleAuthMiddleware(logger *logrus.Logger) *SimpleAuthMiddleware {
	return &SimpleAuthMiddleware{
		Log: logger,
	}
}

// RequireAuth middleware to validate API key from X-API-Key header
func (m *SimpleAuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		
		// Get API key from header
		apiKey := c.Get("X-API-Key")
		
		// Check if API key exists
		if apiKey == "" {
			m.Log.WithFields(logrus.Fields{
				"request_id": requestID,
				"path":       c.Path(),
				"method":     c.Method(),
			}).Warn("Missing API key")
			
			return response.JSONError(c, 
				appErrors.WithMessage(appErrors.ErrUnauthorized, "Missing API key"), 
				m.Log)
		}

		// Simple API key validation - in a real app, this would check against a database
		// For this example, we're using a hardcoded value
		if apiKey != "order-service-api-key" {
			m.Log.WithFields(logrus.Fields{
				"request_id": requestID,
				"path":       c.Path(),
				"method":     c.Method(),
			}).Warn("Invalid API key")
			
			return response.JSONError(c, 
				appErrors.WithMessage(appErrors.ErrUnauthorized, "Invalid API key"), 
				m.Log)
		}
		
		// Set a dummy user ID in locals - this would normally come from your auth service
		c.Locals("userId", "service-account")
		
		// Also set in the context
		c.SetUserContext(context.WithUserID(c.UserContext(), "service-account"))
		
		// Log successful authentication
		m.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"path":       c.Path(),
		}).Info("API key authentication successful")
		
		// Call next handler
		return c.Next()
	}
}