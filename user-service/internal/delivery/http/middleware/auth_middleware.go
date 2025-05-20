package middleware

import (
	"user-service/internal/context"
	"user-service/internal/delivery/http/response"
	appErrors "user-service/internal/errors"
	"user-service/internal/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AuthMiddleware struct {
	DB             *gorm.DB
	UserRepository repository.UserRepositoryInterface
	Log            *logrus.Logger
}

func NewAuthMiddleware(db *gorm.DB, userRepository repository.UserRepositoryInterface) *AuthMiddleware {
	return &AuthMiddleware{
		DB:             db,
		UserRepository: userRepository,
		Log:            logrus.New(),
	}
}

// SetLogger sets the logger instance for the middleware
func (m *AuthMiddleware) SetLogger(log *logrus.Logger) {
	m.Log = log
}

// RequireAuth middleware to validate token from Authorization header
func (m *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		
		// Get token from Authorization header
		authHeader := c.Get("Authorization")
		
		// Check if token exists
		if authHeader == "" {
			m.Log.WithFields(logrus.Fields{
				"request_id": requestID,
				"path":       c.Path(),
				"method":     c.Method(),
			}).Warn("Missing authorization header")
			
			return response.JSONError(c, 
				appErrors.WithMessage(appErrors.ErrUnauthorized, "Missing authorization header"), 
				m.Log)
		}
		
		// Extract token value - handle "Bearer " prefix if present
		token := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
		
		// Create a context with timeout for the token validation
		userCtx := context.WithRequestID(c.UserContext(), requestID)
		timeoutCtx, cancel := context.WithDefaultTimeout(userCtx)
		defer cancel()
		
		// Use a repository with context
		c.SetUserContext(timeoutCtx)
		
		// Validate token
		user, err := m.UserRepository.FindByToken(m.DB, token)
		if err != nil {
			m.Log.WithFields(logrus.Fields{
				"request_id": requestID,
				"token":      token[:8] + "...", // Only log part of the token for security
				"error":      err.Error(),
				"path":       c.Path(),
			}).Warn("Invalid token")
			
			return response.JSONError(c, 
				appErrors.WithError(appErrors.ErrUnauthorized, err), 
				m.Log)
		}
		
		if user == nil {
			m.Log.WithFields(logrus.Fields{
				"request_id": requestID,
				"token":      token[:8] + "...",
				"path":       c.Path(),
			}).Warn("User not found for token")
			
			return response.JSONError(c, 
				appErrors.WithMessage(appErrors.ErrUnauthorized, "Invalid token"), 
				m.Log)
		}
		
		// Set user ID in locals to be used in handlers
		c.Locals("userId", user.ID)
		
		// Also set in the context
		c.SetUserContext(context.WithUserID(c.UserContext(), user.ID.String()))
		
		// Log successful authentication
		m.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    user.ID.String(),
			"path":       c.Path(),
		}).Info("User authenticated successfully")
		
		// Call next handler
		return c.Next()
	}
}