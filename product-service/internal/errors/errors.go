package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError represents application-specific errors
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"` // HTTP status code
	Err        error  `json:"-"` // Original error
}

// Error returns the error message
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is compares error types
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// NewAppError creates a new AppError
func NewAppError(code string, message string, statusCode int, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// Common error types
var (
	ErrInvalidInput = NewAppError(
		"INVALID_INPUT",
		"Invalid input data",
		http.StatusBadRequest,
		nil,
	)

	ErrResourceNotFound = NewAppError(
		"RESOURCE_NOT_FOUND",
		"Resource not found",
		http.StatusNotFound,
		nil,
	)

	ErrProductNotFound = NewAppError(
		"PRODUCT_NOT_FOUND",
		"Product not found",
		http.StatusNotFound,
		nil,
	)

	ErrDuplicateSKU = NewAppError(
		"DUPLICATE_SKU",
		"Product with this SKU already exists",
		http.StatusConflict,
		nil,
	)

	ErrInternalServer = NewAppError(
		"INTERNAL_SERVER_ERROR",
		"Internal server error",
		http.StatusInternalServerError,
		nil,
	)

	ErrTimeout = NewAppError(
		"TIMEOUT",
		"Operation timed out",
		http.StatusRequestTimeout,
		nil,
	)

	ErrInvalidProductID = NewAppError(
		"INVALID_PRODUCT_ID",
		"Invalid product ID format",
		http.StatusBadRequest,
		nil,
	)
)

// WithError wraps the original error with AppError
func WithError(appErr *AppError, err error) *AppError {
	return &AppError{
		Code:       appErr.Code,
		Message:    appErr.Message,
		StatusCode: appErr.StatusCode,
		Err:        err,
	}
}

// WithMessage creates a new error with a custom message
func WithMessage(appErr *AppError, message string) *AppError {
	return &AppError{
		Code:       appErr.Code,
		Message:    message,
		StatusCode: appErr.StatusCode,
		Err:        appErr.Err,
	}
}

// As is a helper for errors.As
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Is is a helper for errors.Is
func Is(err, target error) bool {
	return errors.Is(err, target)
}