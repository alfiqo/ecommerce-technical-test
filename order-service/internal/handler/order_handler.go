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

type OrderHandler struct {
	Log         *logrus.Logger
	OrderUseCase usecase.OrderUseCaseInterface
}

func NewOrderHandler(orderUseCase usecase.OrderUseCaseInterface, logger *logrus.Logger) *OrderHandler {
	return &OrderHandler{
		Log:         logger,
		OrderUseCase: orderUseCase,
	}
}

// CreateOrder godoc
// @Summary Create a new order
// @Description Create a new order with items
// @Tags Orders
// @Accept json
// @Produce json
// @Param order body model.CreateOrderRequest true "Order creation request"
// @Success 201 {object} model.OrderResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security ApiKeyAuth
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(ctx *fiber.Ctx) error {
	// Get request ID from context for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)
	
	request := new(model.CreateOrderRequest)
	if err := ctx.BodyParser(request); err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to parse request body")
		return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInvalidInput, err), h.Log)
	}

	// Get the authenticated user ID from context
	userId := ctx.Locals("userId")
	if userId != nil {
		// Set the user ID if provided through auth
		request.UserID = userId.(string)
	}

	// Add timeout to context
	timeoutCtx, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()

	orderResponse, err := h.OrderUseCase.CreateOrder(timeoutCtx, request)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    request.UserID,
			"error":      err.Error(),
		}).Warn("Failed to create order")
		
		// Handle specific error types
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			return response.JSONError(ctx, appErr, h.Log)
		}

		// Convert Fiber errors to application errors
		if err == fiber.ErrBadRequest {
			return response.JSONError(ctx, appErrors.ErrInvalidInput, h.Log)
		} else {
			return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInternalServer, err), h.Log)
		}
	}

	return response.JSONCreated(ctx, orderResponse)
}

// GetOrder godoc
// @Summary Get order by ID
// @Description Returns order details for the specified ID
// @Tags Orders
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} model.OrderResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security ApiKeyAuth
// @Router /orders/{id} [get]
func (h *OrderHandler) GetOrder(ctx *fiber.Ctx) error {
	// Get request ID from context for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)
	
	// Parse order ID from URL
	orderIDStr := ctx.Params("id")
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

	orderResponse, err := h.OrderUseCase.GetOrderByID(timeoutCtx, uint(orderID))
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"order_id":   orderID,
			"error":      err.Error(),
		}).Warn("Failed to get order")
		
		// Handle specific error types
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			return response.JSONError(ctx, appErr, h.Log)
		}

		// Convert Fiber errors to application errors
		if err == fiber.ErrNotFound {
			return response.JSONError(ctx, appErrors.ErrOrderNotFound, h.Log)
		} else {
			return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInternalServer, err), h.Log)
		}
	}

	return response.JSONSuccess(ctx, orderResponse)
}

// GetUserOrders godoc
// @Summary Get orders for a user
// @Description Returns paginated list of orders for the specified user ID
// @Tags Orders
// @Produce json
// @Param user_id query string false "User ID (defaults to authenticated user)"
// @Param page query int false "Page number (defaults to 1)"
// @Param limit query int false "Items per page (defaults to 10, max 100)"
// @Success 200 {array} model.OrderResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security ApiKeyAuth
// @Router /orders [get]
func (h *OrderHandler) GetUserOrders(ctx *fiber.Ctx) error {
	// Get request ID from context for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)
	
	// Get the authenticated user ID from context
	authUserID := ctx.Locals("userId").(string)
	
	// Parse query parameters
	userID := ctx.Query("user_id", authUserID) // Default to authenticated user
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	
	// Validate page and limit
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Add timeout to context
	timeoutCtx, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()

	orders, total, err := h.OrderUseCase.GetOrdersByUserID(timeoutCtx, userID, page, limit)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"user_id":    userID,
			"error":      err.Error(),
		}).Warn("Failed to get user orders")
		
		// Handle specific error types
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			return response.JSONError(ctx, appErr, h.Log)
		}

		return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInternalServer, err), h.Log)
	}

	// Create pagination metadata
	meta := map[string]interface{}{
		"total":     total,
		"page":      page,
		"limit":     limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	}

	// Create response with data and metadata
	result := map[string]interface{}{
		"data": orders,
		"meta": meta,
	}

	return response.JSONSuccess(ctx, result)
}

// UpdateOrderStatus godoc
// @Summary Update order status
// @Description Update the status of an order
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param status body model.UpdateOrderStatusRequest true "New order status"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security ApiKeyAuth
// @Router /orders/{id}/status [patch]
func (h *OrderHandler) UpdateOrderStatus(ctx *fiber.Ctx) error {
	// Get request ID from context for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)
	
	// Parse order ID from URL
	orderIDStr := ctx.Params("id")
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

	// Parse request body
	request := new(model.UpdateOrderStatusRequest)
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

	err = h.OrderUseCase.UpdateOrderStatus(timeoutCtx, uint(orderID), request.Status)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"order_id":   orderID,
			"status":     request.Status,
			"error":      err.Error(),
		}).Warn("Failed to update order status")
		
		// Handle specific error types
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			return response.JSONError(ctx, appErr, h.Log)
		}

		// Convert Fiber errors to application errors
		if err == fiber.ErrNotFound {
			return response.JSONError(ctx, appErrors.ErrOrderNotFound, h.Log)
		} else if err == fiber.ErrBadRequest {
			return response.JSONError(ctx, appErrors.ErrInvalidOrderStatus, h.Log)
		} else {
			return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInternalServer, err), h.Log)
		}
	}

	return response.JSONSuccess(ctx, map[string]interface{}{
		"message": "Order status updated successfully",
	})
}

// ProcessPayment godoc
// @Summary Process payment for an order
// @Description Process payment for a pending order
// @Tags Orders
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security ApiKeyAuth
// @Router /orders/{id}/payment [post]
func (h *OrderHandler) ProcessPayment(ctx *fiber.Ctx) error {
	// Get request ID from context for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)
	
	// Parse order ID from URL
	orderIDStr := ctx.Params("id")
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

	err = h.OrderUseCase.ProcessPayment(timeoutCtx, uint(orderID))
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"order_id":   orderID,
			"error":      err.Error(),
		}).Warn("Failed to process payment")
		
		// Handle specific error types
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			return response.JSONError(ctx, appErr, h.Log)
		}

		// Convert Fiber errors to application errors
		if err == fiber.ErrNotFound {
			return response.JSONError(ctx, appErrors.ErrOrderNotFound, h.Log)
		} else if err == fiber.ErrBadRequest {
			return response.JSONError(ctx, appErrors.ErrOrderAlreadyPaid, h.Log)
		} else {
			return response.JSONError(ctx, appErrors.WithError(appErrors.ErrPaymentFailed, err), h.Log)
		}
	}

	return response.JSONSuccess(ctx, map[string]interface{}{
		"message": "Payment processed successfully",
	})
}