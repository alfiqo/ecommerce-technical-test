package errors

import (
	"net/http"
)

// Order-specific error types
var (
	ErrOrderNotFound = NewAppError(
		"ORDER_NOT_FOUND",
		"Order not found",
		http.StatusNotFound,
		nil,
	)

	ErrInvalidOrderStatus = NewAppError(
		"INVALID_ORDER_STATUS",
		"Invalid order status",
		http.StatusBadRequest,
		nil,
	)

	ErrOrderAlreadyPaid = NewAppError(
		"ORDER_ALREADY_PAID",
		"Order has already been paid",
		http.StatusBadRequest,
		nil,
	)

	ErrOrderCancelled = NewAppError(
		"ORDER_CANCELLED",
		"Order has been cancelled",
		http.StatusBadRequest,
		nil,
	)

	ErrPaymentFailed = NewAppError(
		"PAYMENT_FAILED",
		"Failed to process payment",
		http.StatusInternalServerError,
		nil,
	)

	ErrInsufficientStock = NewAppError(
		"INSUFFICIENT_STOCK",
		"Insufficient stock to fulfill order",
		http.StatusBadRequest,
		nil,
	)

	ErrProductNotFound = NewAppError(
		"PRODUCT_NOT_FOUND",
		"Product not found",
		http.StatusBadRequest,
		nil,
	)

	ErrReservationFailed = NewAppError(
		"RESERVATION_FAILED",
		"Failed to reserve stock",
		http.StatusInternalServerError,
		nil,
	)

	ErrReservationNotFound = NewAppError(
		"RESERVATION_NOT_FOUND",
		"Reservation not found",
		http.StatusNotFound,
		nil,
	)

	ErrReservationExpired = NewAppError(
		"RESERVATION_EXPIRED",
		"Reservation has expired",
		http.StatusBadRequest,
		nil,
	)
)