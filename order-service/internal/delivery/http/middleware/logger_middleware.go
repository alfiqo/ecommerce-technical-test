package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// Logger creates a middleware that logs requests and responses
func Logger(log *logrus.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Start timer
		start := time.Now()
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("%d", time.Now().UnixNano())
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