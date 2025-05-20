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

type StockHandler struct {
	Log     *logrus.Logger
	UseCase usecase.StockUseCaseInterface
}

func NewStockHandler(useCase usecase.StockUseCaseInterface, logger *logrus.Logger) *StockHandler {
	return &StockHandler{
		Log:     logger,
		UseCase: useCase,
	}
}

// GetWarehouseStock godoc
// @Summary Get warehouse stock
// @Description Returns a paginated list of stock items in a warehouse
// @Tags Stock
// @Produce json
// @Param warehouseId path string true "Warehouse ID"
// @Param productId query string false "Product ID filter"
// @Param page query int false "Page number (defaults to 1)"
// @Param limit query int false "Items per page (defaults to 20, max 100)"
// @Success 200 {object} model.WarehouseStockListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security ApiKeyAuth
// @Router /warehouses/{warehouseId}/stock [get]
func (c *StockHandler) GetWarehouseStock(ctx *fiber.Ctx) error {
	// Get request ID for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)

	// Get the warehouse ID from the URL
	warehouseIDParam := ctx.Params("warehouseId")
	warehouseID, err := strconv.ParseUint(warehouseIDParam, 10, 32)
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"id":         warehouseIDParam,
			"error":      err.Error(),
		}).Warn("Invalid warehouse ID format")
		return response.JSONError(ctx, appErrors.ErrInvalidInput, c.Log)
	}

	// Parse product ID filter if provided
	var productID uint = 0
	if productIDParam := ctx.Query("productId"); productIDParam != "" {
		productIDUint, err := strconv.ParseUint(productIDParam, 10, 32)
		if err != nil {
			c.Log.WithFields(logrus.Fields{
				"request_id": requestID,
				"productId":  productIDParam,
				"error":      err.Error(),
			}).Warn("Invalid product ID format")
			return response.JSONError(ctx, appErrors.ErrInvalidInput, c.Log)
		}
		productID = uint(productIDUint)
	}

	// Parse pagination parameters
	page := 1
	limit := 20

	// Parse page parameter
	if pageStr := ctx.Query("page"); pageStr != "" {
		pageVal, err := strconv.Atoi(pageStr)
		if err != nil || pageVal < 1 {
			c.Log.WithFields(logrus.Fields{
				"request_id": requestID,
				"page":       pageStr,
				"error":      "Invalid page parameter",
			}).Warn("Invalid page parameter")
			return response.JSONError(ctx, appErrors.WithMessage(appErrors.ErrInvalidInput, "Invalid page parameter"), c.Log)
		}
		page = pageVal
	}

	// Parse limit parameter
	if limitStr := ctx.Query("limit"); limitStr != "" {
		limitVal, err := strconv.Atoi(limitStr)
		if err != nil || limitVal < 1 || limitVal > 100 {
			c.Log.WithFields(logrus.Fields{
				"request_id": requestID,
				"limit":      limitStr,
				"error":      "Invalid limit parameter",
			}).Warn("Invalid limit parameter")
			return response.JSONError(ctx, appErrors.WithMessage(appErrors.ErrInvalidInput, "Invalid limit parameter (1-100)"), c.Log)
		}
		limit = limitVal
	}

	// Add timeout to context
	timeoutCtx, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()

	// Call the use case to get warehouse stock
	stockResponse, err := c.UseCase.GetWarehouseStock(timeoutCtx, uint(warehouseID), productID, page, limit)
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id":  requestID,
			"warehouseId": warehouseID,
			"productId":   productID,
			"page":        page,
			"limit":       limit,
			"error":       err.Error(),
		}).Warn("Failed to get warehouse stock")

		// Handle specific error types
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			return response.JSONError(ctx, appErr, c.Log)
		}

		if err == fiber.ErrNotFound || errors.Is(err, appErrors.ErrResourceNotFound) {
			return response.JSONError(ctx, appErrors.ErrResourceNotFound, c.Log)
		}

		return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInternalServer, err), c.Log)
	}

	return response.JSONSuccess(ctx, stockResponse)
}

// AddStock godoc
// @Summary Add stock to a warehouse
// @Description Adds stock to a warehouse for a specific product
// @Tags Stock
// @Accept json
// @Produce json
// @Param warehouseId path string true "Warehouse ID"
// @Param stock body model.AddStockRequest true "Stock details to add"
// @Success 200 {object} model.StockResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security ApiKeyAuth
// @Router /warehouses/{warehouseId}/stock [post]
func (c *StockHandler) AddStock(ctx *fiber.Ctx) error {
	// Get request ID for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)

	// Get the warehouse ID from the URL
	warehouseIDParam := ctx.Params("warehouseId")
	warehouseID, err := strconv.ParseUint(warehouseIDParam, 10, 32)
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"id":         warehouseIDParam,
			"error":      err.Error(),
		}).Warn("Invalid warehouse ID format")
		return response.JSONError(ctx, appErrors.ErrInvalidInput, c.Log)
	}

	// Parse request body
	request := new(model.AddStockRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to parse request body")
		return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInvalidInput, err), c.Log)
	}

	// Ensure warehouse ID in the URL matches the one in the request
	request.WarehouseID = uint(warehouseID)

	// Add timeout to context
	timeoutCtx, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()

	// Call the use case to add stock
	stockResponse, err := c.UseCase.AddStock(timeoutCtx, request)
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id":  requestID,
			"warehouseId": warehouseID,
			"productId":   request.ProductID,
			"quantity":    request.Quantity,
			"error":       err.Error(),
		}).Warn("Failed to add stock")

		// Handle specific error types
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			return response.JSONError(ctx, appErr, c.Log)
		}

		if err == fiber.ErrBadRequest {
			return response.JSONError(ctx, appErrors.ErrInvalidInput, c.Log)
		}

		if err == fiber.ErrNotFound || errors.Is(err, appErrors.ErrResourceNotFound) {
			return response.JSONError(ctx, appErrors.ErrResourceNotFound, c.Log)
		}

		return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInternalServer, err), c.Log)
	}

	return response.JSONSuccess(ctx, stockResponse)
}

// TransferStock godoc
// @Summary Transfer stock between warehouses
// @Description Transfers stock from one warehouse to another for a specific product
// @Tags Stock
// @Accept json
// @Produce json
// @Param transfer body model.StockTransferRequest true "Stock transfer details"
// @Success 200 {object} model.StockTransferResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security ApiKeyAuth
// @Router /stock/transfer [post]
func (c *StockHandler) TransferStock(ctx *fiber.Ctx) error {
	// Get request ID for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)

	// Parse request body
	request := new(model.StockTransferRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to parse request body")
		return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInvalidInput, err), c.Log)
	}

	// Add timeout to context
	timeoutCtx, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()

	// Call the use case to transfer stock
	transferResponse, err := c.UseCase.TransferStock(timeoutCtx, request)
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id":        requestID,
			"sourceWarehouseId": request.SourceWarehouseID,
			"targetWarehouseId": request.TargetWarehouseID,
			"productId":         request.ProductID,
			"quantity":          request.Quantity,
			"error":             err.Error(),
		}).Warn("Failed to transfer stock")

		// Handle specific error types
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			return response.JSONError(ctx, appErr, c.Log)
		}

		if err == fiber.ErrBadRequest {
			return response.JSONError(ctx, appErrors.ErrInvalidInput, c.Log)
		}

		if err == fiber.ErrNotFound || errors.Is(err, appErrors.ErrResourceNotFound) {
			return response.JSONError(ctx, appErrors.ErrResourceNotFound, c.Log)
		}

		return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInternalServer, err), c.Log)
	}

	return response.JSONSuccess(ctx, transferResponse)
}