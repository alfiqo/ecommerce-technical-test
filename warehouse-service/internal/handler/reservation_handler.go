package handler

import (
	"errors"
	"strconv"
	"warehouse-service/internal/context"
	"warehouse-service/internal/delivery/http/response"
	appErrors "warehouse-service/internal/errors"
	"warehouse-service/internal/model"
	"warehouse-service/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type ReservationHandler struct {
	Log     *logrus.Logger
	UseCase usecase.ReservationUseCaseInterface
}

func NewReservationHandler(useCase usecase.ReservationUseCaseInterface, logger *logrus.Logger) *ReservationHandler {
	return &ReservationHandler{
		Log:     logger,
		UseCase: useCase,
	}
}

// ReserveStock godoc
// @Summary Reserve inventory stock
// @Description Reserves stock for a product in a warehouse using database-level locking
// @Tags Inventory
// @Accept json
// @Produce json
// @Param reservation body model.ReserveStockRequest true "Reservation details"
// @Success 200 {object} model.ReservationResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security ApiKeyAuth
// @Router /inventory/reserve [post]
func (h *ReservationHandler) ReserveStock(ctx *fiber.Ctx) error {
	// Get request ID for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)

	// Parse request body
	request := new(model.ReserveStockRequest)
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

	// Call the use case to reserve stock
	reservationResponse, err := h.UseCase.ReserveStock(timeoutCtx, request)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id":   requestID,
			"warehouse_id": request.WarehouseID,
			"product_id":   request.ProductID,
			"quantity":     request.Quantity,
			"error":        err.Error(),
		}).Warn("Failed to reserve stock")

		// Handle specific error types
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			return response.JSONError(ctx, appErr, h.Log)
		}

		if err == fiber.ErrBadRequest {
			return response.JSONError(ctx, appErrors.ErrInvalidInput, h.Log)
		}

		if err == fiber.ErrNotFound || errors.Is(err, appErrors.ErrResourceNotFound) {
			return response.JSONError(ctx, appErrors.ErrResourceNotFound, h.Log)
		}

		return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInternalServer, err), h.Log)
	}

	return response.JSONSuccess(ctx, reservationResponse)
}

// CancelReservation godoc
// @Summary Cancel a stock reservation
// @Description Cancels a previously made stock reservation
// @Tags Inventory
// @Accept json
// @Produce json
// @Param cancelation body model.CancelReservationRequest true "Cancellation details"
// @Success 200 {object} response.SuccessMessageResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security ApiKeyAuth
// @Router /inventory/reserve/cancel [post]
func (h *ReservationHandler) CancelReservation(ctx *fiber.Ctx) error {
	// Get request ID for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)

	// Parse request body
	request := new(model.CancelReservationRequest)
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

	// Call the use case to cancel reservation
	err := h.UseCase.CancelReservation(timeoutCtx, request)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id":   requestID,
			"warehouse_id": request.WarehouseID,
			"product_id":   request.ProductID,
			"quantity":     request.Quantity,
			"reference":    request.Reference,
			"error":        err.Error(),
		}).Warn("Failed to cancel reservation")

		// Handle specific error types
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			return response.JSONError(ctx, appErr, h.Log)
		}

		if err == fiber.ErrBadRequest {
			return response.JSONError(ctx, appErrors.ErrInvalidInput, h.Log)
		}

		if err == fiber.ErrNotFound || errors.Is(err, appErrors.ErrResourceNotFound) {
			return response.JSONError(ctx, appErrors.ErrResourceNotFound, h.Log)
		}

		return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInternalServer, err), h.Log)
	}

	return response.JSONSuccess(ctx, map[string]string{"message": "Reservation cancelled successfully"})
}

// CommitReservation godoc
// @Summary Commit a stock reservation
// @Description Commits a previously made stock reservation, reducing actual stock
// @Tags Inventory
// @Accept json
// @Produce json
// @Param commit body model.CommitReservationRequest true "Commit details"
// @Success 200 {object} response.SuccessMessageResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security ApiKeyAuth
// @Router /inventory/reserve/commit [post]
func (h *ReservationHandler) CommitReservation(ctx *fiber.Ctx) error {
	// Get request ID for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)

	// Parse request body
	request := new(model.CommitReservationRequest)
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

	// Call the use case to commit reservation
	err := h.UseCase.CommitReservation(timeoutCtx, request)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id":   requestID,
			"warehouse_id": request.WarehouseID,
			"product_id":   request.ProductID,
			"quantity":     request.Quantity,
			"reference":    request.Reference,
			"error":        err.Error(),
		}).Warn("Failed to commit reservation")

		// Handle specific error types
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			return response.JSONError(ctx, appErr, h.Log)
		}

		if err == fiber.ErrBadRequest {
			return response.JSONError(ctx, appErrors.ErrInvalidInput, h.Log)
		}

		if err == fiber.ErrNotFound || errors.Is(err, appErrors.ErrResourceNotFound) {
			return response.JSONError(ctx, appErrors.ErrResourceNotFound, h.Log)
		}

		return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInternalServer, err), h.Log)
	}

	return response.JSONSuccess(ctx, map[string]string{"message": "Reservation committed successfully"})
}

// GetReservationHistory godoc
// @Summary Get reservation history
// @Description Returns the reservation history for a product in a warehouse
// @Tags Inventory
// @Produce json
// @Param warehouse_id path int true "Warehouse ID"
// @Param product_id path int true "Product ID"
// @Param page query int false "Page number (defaults to 1)"
// @Param limit query int false "Items per page (defaults to 20, max 100)"
// @Success 200 {object} model.ReservationHistoryResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security ApiKeyAuth
// @Router /inventory/warehouses/{warehouse_id}/products/{product_id}/reservations [get]
func (h *ReservationHandler) GetReservationHistory(ctx *fiber.Ctx) error {
	// Get request ID for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)

	// Get path parameters
	warehouseIDParam := ctx.Params("warehouse_id")
	warehouseID, err := strconv.ParseUint(warehouseIDParam, 10, 32)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id":   requestID,
			"warehouse_id": warehouseIDParam,
			"error":        err.Error(),
		}).Warn("Invalid warehouse ID format")
		return response.JSONError(ctx, appErrors.ErrInvalidInput, h.Log)
	}

	productIDParam := ctx.Params("product_id")
	productID, err := strconv.ParseUint(productIDParam, 10, 32)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"product_id": productIDParam,
			"error":      err.Error(),
		}).Warn("Invalid product ID format")
		return response.JSONError(ctx, appErrors.ErrInvalidInput, h.Log)
	}

	// Parse pagination parameters
	page := 1
	limit := 20

	// Parse page parameter
	if pageStr := ctx.Query("page"); pageStr != "" {
		pageNum, err := strconv.Atoi(pageStr)
		if err != nil || pageNum < 1 {
			h.Log.WithFields(logrus.Fields{
				"request_id": requestID,
				"page":       pageStr,
				"error":      "Invalid page parameter",
			}).Warn("Invalid page parameter")
			return response.JSONError(ctx, appErrors.WithMessage(appErrors.ErrInvalidInput, "Invalid page parameter"), h.Log)
		}
		page = pageNum
	}

	// Parse limit parameter
	if limitStr := ctx.Query("limit"); limitStr != "" {
		limitNum, err := strconv.Atoi(limitStr)
		if err != nil || limitNum < 1 || limitNum > 100 {
			h.Log.WithFields(logrus.Fields{
				"request_id": requestID,
				"limit":      limitStr,
				"error":      "Invalid limit parameter",
			}).Warn("Invalid limit parameter")
			return response.JSONError(ctx, appErrors.WithMessage(appErrors.ErrInvalidInput, "Invalid limit parameter (1-100)"), h.Log)
		}
		limit = limitNum
	}

	// Add timeout to context
	timeoutCtx, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()

	// Call the use case to get reservation history
	history, err := h.UseCase.GetReservationHistory(timeoutCtx, uint(warehouseID), uint(productID), page, limit)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id":   requestID,
			"warehouse_id": warehouseID,
			"product_id":   productID,
			"page":         page,
			"limit":        limit,
			"error":        err.Error(),
		}).Warn("Failed to get reservation history")

		// Handle specific error types
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			return response.JSONError(ctx, appErr, h.Log)
		}

		return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInternalServer, err), h.Log)
	}

	return response.JSONSuccess(ctx, history)
}