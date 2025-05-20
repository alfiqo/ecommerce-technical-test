package response

import (
	"errors"
	appErrors "warehouse-service/internal/errors"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// Response is a standardized API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo provides detailed error information
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// JSONSuccess sends a successful JSON response
func JSONSuccess(c *fiber.Ctx, data interface{}) error {
	response := Response{
		Success: true,
		Data:    data,
	}
	
	return c.Status(fiber.StatusOK).JSON(response)
}

// JSONError sends an error JSON response
func JSONError(c *fiber.Ctx, err error, logger *logrus.Logger) error {
	// Default to internal server error
	statusCode := fiber.StatusInternalServerError
	errorCode := "INTERNAL_SERVER_ERROR"
	message := "An unexpected error occurred"

	// Extract details from AppError if possible
	var appErr *appErrors.AppError
	if errors.As(err, &appErr) {
		statusCode = appErr.StatusCode
		errorCode = appErr.Code
		message = appErr.Message
		
		// Log the error with context
		fields := logrus.Fields{
			"error_code":   appErr.Code,
			"status_code":  appErr.StatusCode,
			"request_path": c.Path(),
			"method":       c.Method(),
		}

		if appErr.Err != nil {
			fields["original_error"] = appErr.Err.Error()
			logger.WithFields(fields).Error(appErr.Message)
		} else {
			logger.WithFields(fields).Error(appErr.Message)
		}
	} else {
		// For standard errors
		logger.WithFields(logrus.Fields{
			"request_path": c.Path(),
			"method":       c.Method(),
		}).Error(err.Error())
	}

	response := Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    errorCode,
			Message: message,
		},
	}

	return c.Status(statusCode).JSON(response)
}

// HandleError provides a centralized error handler
func HandleError(c *fiber.Ctx, err error, logger *logrus.Logger) error {
	// Fiber errors are handled specially
	if fiberErr, ok := err.(*fiber.Error); ok {
		switch fiberErr.Code {
		case fiber.StatusBadRequest:
			err = appErrors.ErrInvalidInput
		case fiber.StatusUnauthorized:
			err = appErrors.ErrUnauthorized
		case fiber.StatusNotFound:
			err = appErrors.ErrResourceNotFound
		default:
			err = appErrors.WithError(appErrors.ErrInternalServer, fiberErr)
		}
	}
	
	return JSONError(c, err, logger)
}