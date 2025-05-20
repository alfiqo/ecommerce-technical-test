package middleware

import (
	"warehouse-service/internal/context"
	"warehouse-service/internal/delivery/http/response"
	appErrors "warehouse-service/internal/errors"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Simplified AuthMiddleware to check for admin token
// In a real application, this would be replaced with JWT or a proper authentication system
type AuthMiddleware struct {
	DB          *gorm.DB
	Log         *logrus.Logger
	AdminTokens []string
}

func NewAuthMiddleware(db *gorm.DB) *AuthMiddleware {
	// Fixed admin tokens for demo purposes
	// In a real application, these would be stored securely, not hardcoded
	return &AuthMiddleware{
		DB:          db,
		Log:         logrus.New(),
		AdminTokens: []string{"admin_token_here"}, // Use value from sequence diagram
	}
}

// SetLogger sets the logger instance for the middleware
func (m *AuthMiddleware) SetLogger(log *logrus.Logger) {
	m.Log = log
}

// RequireAuth middleware to validate token from Authorization header
// Requires admin access
func (m *AuthMiddleware) RequireAuth() fiber.Handler {
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
		if apiKey != "warehouse-service-api-key" {
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
