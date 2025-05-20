package middleware

import (
	"shop-service/internal/context"
	"shop-service/internal/delivery/http/response"
	appErrors "shop-service/internal/errors"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type AuthMiddleware struct {
	Log *logrus.Logger
}

func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{
		Log: logrus.New(),
	}
}

// SetLogger sets the logger instance for the middleware
func (m *AuthMiddleware) SetLogger(log *logrus.Logger) {
	m.Log = log
}

// Basic auth middleware for API key validation
func (m *AuthMiddleware) RequireAPIKey(apiKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		
		// Get API key from header
		headerAPIKey := c.Get("X-API-Key")
		
		// Check if API key exists and matches
		if headerAPIKey == "" || headerAPIKey != apiKey {
			m.Log.WithFields(logrus.Fields{
				"request_id": requestID,
				"path":       c.Path(),
				"method":     c.Method(),
			}).Warn("Invalid or missing API key")
			
			return response.JSONError(c, 
				appErrors.WithMessage(appErrors.ErrUnauthorized, "Invalid or missing API key"), 
				m.Log)
		}
		
		// Create a context with request ID
		userCtx := context.WithRequestID(c.UserContext(), requestID)
		timeoutCtx, cancel := context.WithDefaultTimeout(userCtx)
		defer cancel()
		
		// Use a context with timeout
		c.SetUserContext(timeoutCtx)
		
		// Log successful authentication
		m.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"path":       c.Path(),
		}).Info("API key authentication successful")
		
		// Call next handler
		return c.Next()
	}
}