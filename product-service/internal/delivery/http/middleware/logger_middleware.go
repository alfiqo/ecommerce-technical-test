package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Logger creates a middleware that logs requests and responses
func Logger(log *logrus.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Start timer
		start := time.Now()
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			c.Set("X-Request-ID", requestID)
		}

		// Log request
		log.WithFields(logrus.Fields{
			"request_id": requestID,
			"method":     c.Method(),
			"path":       c.Path(),
			"ip":         c.IP(),
			"user_agent": c.Get("User-Agent"),
		}).Info("Incoming request")

		// Process request
		err := c.Next()

		// Get latency
		latency := time.Since(start)

		// Log response
		statusCode := c.Response().StatusCode()
		fields := logrus.Fields{
			"request_id":  requestID,
			"status_code": statusCode,
			"latency_ms":  latency.Milliseconds(),
		}

		// Customize log level based on status code
		switch {
		case statusCode >= 500:
			log.WithFields(fields).Error("Server error response")
		case statusCode >= 400:
			log.WithFields(fields).Warn("Client error response")
		case statusCode >= 300:
			log.WithFields(fields).Info("Redirection response")
		default:
			log.WithFields(fields).Info("Success response")
		}

		return err
	}
}

// RequestID adds a request ID to all requests
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if request already has an ID
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			// Generate a new request ID
			requestID = uuid.New().String()
			c.Set("X-Request-ID", requestID)
		}

		// Add request ID to context for logging
		c.Locals("requestID", requestID)

		// Continue with the request
		return c.Next()
	}
}

// ErrorHandler creates a middleware that handles errors
func ErrorHandler(log *logrus.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Get request ID for logging context
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("%d", time.Now().UnixNano())
		}

		// Log the error
		log.WithFields(logrus.Fields{
			"request_id":   requestID,
			"method":       c.Method(),
			"path":         c.Path(),
			"error":        err.Error(),
			"error_type":   fmt.Sprintf("%T", err),
			"status_code":  fiber.StatusInternalServerError,
		}).Error("Request error")

		// Default to internal server error
		code := fiber.StatusInternalServerError
		message := "Internal Server Error"

		// Check if it's a Fiber error
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
			message = e.Message
		}

		// Return JSON error response
		return c.Status(code).JSON(fiber.Map{
			"success": false,
			"error": fiber.Map{
				"code":    code,
				"message": message,
			},
		})
	}
}