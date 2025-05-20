package handler

import (
	"errors"
	"order-service/internal/context"
	"order-service/internal/delivery/http/response"
	appErrors "order-service/internal/errors"
	"order-service/internal/model"
	"order-service/internal/usecase"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type ReservationHandler struct {
	Log               *logrus.Logger
	ReservationUseCase usecase.ReservationUseCaseInterface
}

func NewReservationHandler(reservationUseCase usecase.ReservationUseCaseInterface, logger *logrus.Logger) *ReservationHandler {
	return &ReservationHandler{
		Log:               logger,
		ReservationUseCase: reservationUseCase,
	}
}

// CreateReservation godoc
// @Summary Create a stock reservation
// @Description Create a new stock reservation for an order
// @Tags Reservations
// @Accept json
// @Produce json
// @Param reservation body model.ReservationRequest true "Reservation details"
// @Success 201 {object} model.ReservationResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security ApiKeyAuth
// @Router /reservations [post]
func (h *ReservationHandler) CreateReservation(ctx *fiber.Ctx) error {
	// Get request ID from context for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)
	
	request := new(model.ReservationRequest)
	if err := ctx.BodyParser(request); err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to parse request body")
		return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInvalidInput, err), h.Log)
	}

	// Add timeout to context
	timeoutCtx, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()

	reservationResponse, err := h.ReservationUseCase.CreateReservation(timeoutCtx, request)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id":  requestID,
			"order_id":    request.OrderID,
			"product_id":  request.ProductID,
			"error":       err.Error(),
		}).Warn("Failed to create reservation")
		
		// Handle specific error types
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			return response.JSONError(ctx, appErr, h.Log)
		}

		// Convert Fiber errors to application errors
		if err == fiber.ErrNotFound {
			return response.JSONError(ctx, appErrors.ErrOrderNotFound, h.Log)
		} else if err == fiber.ErrBadRequest {
			return response.JSONError(ctx, appErrors.ErrInvalidInput, h.Log)
		} else {
			return response.JSONError(ctx, appErrors.WithError(appErrors.ErrReservationFailed, err), h.Log)
		}
	}

	return response.JSONCreated(ctx, reservationResponse)
}

// GetOrderReservations godoc
// @Summary Get reservations for an order
// @Description Returns all stock reservations for the specified order ID
// @Tags Reservations
// @Produce json
// @Param order_id path int true "Order ID"
// @Success 200 {array} model.ReservationResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security ApiKeyAuth
// @Router /orders/{order_id}/reservations [get]
func (h *ReservationHandler) GetOrderReservations(ctx *fiber.Ctx) error {
	// Get request ID from context for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)
	
	// Parse order ID from URL
	orderIDStr := ctx.Params("order_id")
	if orderIDStr == "" {
		return response.JSONError(ctx, appErrors.WithMessage(appErrors.ErrInvalidInput, "order id is required"), h.Log)
	}

	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"order_id":   orderIDStr,
			"error":      err.Error(),
		}).Warn("Invalid order ID format")
		return response.JSONError(ctx, appErrors.WithMessage(appErrors.ErrInvalidInput, "invalid order id"), h.Log)
	}

	// Add timeout to context
	timeoutCtx, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()

	reservations, err := h.ReservationUseCase.GetReservationsByOrderID(timeoutCtx, uint(orderID))
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"order_id":   orderID,
			"error":      err.Error(),
		}).Warn("Failed to get order reservations")
		
		// Handle specific error types
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			return response.JSONError(ctx, appErr, h.Log)
		}

		return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInternalServer, err), h.Log)
	}

	return response.JSONSuccess(ctx, reservations)
}

// DeactivateReservation godoc
// @Summary Deactivate a reservation
// @Description Deactivate a stock reservation
// @Tags Reservations
// @Produce json
// @Param id path int true "Reservation ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security ApiKeyAuth
// @Router /reservations/{id}/deactivate [post]
func (h *ReservationHandler) DeactivateReservation(ctx *fiber.Ctx) error {
	// Get request ID from context for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)
	
	// Parse reservation ID from URL
	reservationIDStr := ctx.Params("id")
	if reservationIDStr == "" {
		return response.JSONError(ctx, appErrors.WithMessage(appErrors.ErrInvalidInput, "reservation id is required"), h.Log)
	}

	reservationID, err := strconv.ParseUint(reservationIDStr, 10, 32)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id":     requestID,
			"reservation_id": reservationIDStr,
			"error":          err.Error(),
		}).Warn("Invalid reservation ID format")
		return response.JSONError(ctx, appErrors.WithMessage(appErrors.ErrInvalidInput, "invalid reservation id"), h.Log)
	}

	// Add timeout to context
	timeoutCtx, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()

	err = h.ReservationUseCase.DeactivateReservation(timeoutCtx, uint(reservationID))
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id":     requestID,
			"reservation_id": reservationID,
			"error":          err.Error(),
		}).Warn("Failed to deactivate reservation")
		
		// Handle specific error types
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			return response.JSONError(ctx, appErr, h.Log)
		}

		return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInternalServer, err), h.Log)
	}

	return response.JSONSuccess(ctx, map[string]interface{}{
		"message": "Reservation deactivated successfully",
	})
}

// CleanupExpiredReservations godoc
// @Summary Cleanup expired reservations
// @Description Deactivate all expired stock reservations
// @Tags Reservations
// @Produce json
// @Success 200 {object} response.SuccessResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security ApiKeyAuth
// @Router /reservations/cleanup [post]
func (h *ReservationHandler) CleanupExpiredReservations(ctx *fiber.Ctx) error {
	// Get request ID from context for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)
	
	// Add timeout to context
	timeoutCtx, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()

	err := h.ReservationUseCase.CleanupExpiredReservations(timeoutCtx)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to cleanup expired reservations")
		
		// Handle specific error types
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			return response.JSONError(ctx, appErr, h.Log)
		}

		return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInternalServer, err), h.Log)
	}

	return response.JSONSuccess(ctx, map[string]interface{}{
		"message": "Expired reservations cleaned up successfully",
	})
}