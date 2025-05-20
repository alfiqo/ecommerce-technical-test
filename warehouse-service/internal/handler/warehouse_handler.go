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

type WarehouseHandler struct {
	Log     *logrus.Logger
	UseCase usecase.WarehouseUseCaseInterface
}

func NewWarehouseHandler(useCase usecase.WarehouseUseCaseInterface, logger *logrus.Logger) *WarehouseHandler {
	return &WarehouseHandler{
		Log:     logger,
		UseCase: useCase,
	}
}

// GetWarehouse godoc
// @Summary Get warehouse by ID
// @Description Returns warehouse details for the specified ID with statistics
// @Tags Warehouses
// @Produce json
// @Param id path string true "Warehouse ID"
// @Success 200 {object} model.WarehouseResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security ApiKeyAuth
// @Router /warehouses/{id} [get]
func (c *WarehouseHandler) GetWarehouse(ctx *fiber.Ctx) error {
	// Get request ID for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)

	// Get the warehouse ID from the URL
	idParam := ctx.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"id":         idParam,
			"error":      err.Error(),
		}).Warn("Invalid warehouse ID format")
		return response.JSONError(ctx, appErrors.ErrInvalidInput, c.Log)
	}

	// Add timeout to context
	timeoutCtx, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()

	// Call the use case to get the warehouse
	warehouseResponse, err := c.UseCase.GetWarehouse(timeoutCtx, uint(id))
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"id":         id,
			"error":      err.Error(),
		}).Warn("Failed to get warehouse")

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

	return response.JSONSuccess(ctx, warehouseResponse)
}

// CreateWarehouse godoc
// @Summary Create a new warehouse
// @Description Creates a new warehouse
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param warehouse body model.CreateWarehouseRequest true "Warehouse creation details"
// @Success 201 {object} model.WarehouseResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security ApiKeyAuth
// @Router /warehouses [post]
func (c *WarehouseHandler) CreateWarehouse(ctx *fiber.Ctx) error {
	// Get request ID for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)

	// Parse request body
	request := new(model.CreateWarehouseRequest)
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

	// Call the use case to create the warehouse
	warehouseResponse, err := c.UseCase.CreateWarehouse(timeoutCtx, request)
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"name":       request.Name,
			"error":      err.Error(),
		}).Warn("Failed to create warehouse")

		// Handle specific error types
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			return response.JSONError(ctx, appErr, c.Log)
		}

		if err == fiber.ErrBadRequest {
			return response.JSONError(ctx, appErrors.ErrInvalidInput, c.Log)
		}

		return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInternalServer, err), c.Log)
	}

	return response.JSONSuccess(ctx, warehouseResponse)
}

// UpdateWarehouse godoc
// @Summary Update an existing warehouse
// @Description Updates an existing warehouse
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param id path string true "Warehouse ID"
// @Param warehouse body model.UpdateWarehouseRequest true "Warehouse update details"
// @Success 200 {object} model.WarehouseResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security ApiKeyAuth
// @Router /warehouses/{id} [put]
func (c *WarehouseHandler) UpdateWarehouse(ctx *fiber.Ctx) error {
	// Get request ID for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)

	// Get the warehouse ID from the URL
	idParam := ctx.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"id":         idParam,
			"error":      err.Error(),
		}).Warn("Invalid warehouse ID format")
		return response.JSONError(ctx, appErrors.ErrInvalidInput, c.Log)
	}

	// Parse request body
	request := new(model.UpdateWarehouseRequest)
	if err := ctx.BodyParser(request); err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to parse request body")
		return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInvalidInput, err), c.Log)
	}

	// Ensure ID in the URL matches ID in the body
	request.ID = uint(id)

	// Add timeout to context
	timeoutCtx, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()

	// Call the use case to update the warehouse
	warehouseResponse, err := c.UseCase.UpdateWarehouse(timeoutCtx, request)
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"id":         id,
			"error":      err.Error(),
		}).Warn("Failed to update warehouse")

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

	return response.JSONSuccess(ctx, warehouseResponse)
}

// DeleteWarehouse godoc
// @Summary Delete a warehouse
// @Description Deletes an existing warehouse by ID
// @Tags Warehouses
// @Produce json
// @Param id path string true "Warehouse ID"
// @Success 200 {object} response.SuccessMessageResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security ApiKeyAuth
// @Router /warehouses/{id} [delete]
func (c *WarehouseHandler) DeleteWarehouse(ctx *fiber.Ctx) error {
	// Get request ID for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)

	// Get the warehouse ID from the URL
	idParam := ctx.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"id":         idParam,
			"error":      err.Error(),
		}).Warn("Invalid warehouse ID format")
		return response.JSONError(ctx, appErrors.ErrInvalidInput, c.Log)
	}

	// Add timeout to context
	timeoutCtx, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()

	// Call the use case to delete the warehouse
	err = c.UseCase.DeleteWarehouse(timeoutCtx, uint(id))
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"id":         id,
			"error":      err.Error(),
		}).Warn("Failed to delete warehouse")

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

	return response.JSONSuccess(ctx, map[string]string{"message": "Warehouse deleted successfully"})
}

// ListWarehouses godoc
// @Summary List warehouses
// @Description Returns a paginated list of warehouses
// @Tags Warehouses
// @Produce json
// @Param page query int false "Page number (defaults to 1)"
// @Param limit query int false "Items per page (defaults to 20, max 100)"
// @Success 200 {object} model.WarehouseListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security ApiKeyAuth
// @Router /warehouses [get]
func (c *WarehouseHandler) ListWarehouses(ctx *fiber.Ctx) error {
	// Get request ID for tracking
	requestID := ctx.Get("X-Request-ID")
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)

	// Parse pagination parameters
	request := &model.ListWarehouseRequest{
		Page:  1,
		Limit: 20,
	}

	// Parse page parameter
	if pageStr := ctx.Query("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			c.Log.WithFields(logrus.Fields{
				"request_id": requestID,
				"page":       pageStr,
				"error":      "Invalid page parameter",
			}).Warn("Invalid page parameter")
			return response.JSONError(ctx, appErrors.WithMessage(appErrors.ErrInvalidInput, "Invalid page parameter"), c.Log)
		}
		request.Page = page
	}

	// Parse limit parameter
	if limitStr := ctx.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			c.Log.WithFields(logrus.Fields{
				"request_id": requestID,
				"limit":      limitStr,
				"error":      "Invalid limit parameter",
			}).Warn("Invalid limit parameter")
			return response.JSONError(ctx, appErrors.WithMessage(appErrors.ErrInvalidInput, "Invalid limit parameter (1-100)"), c.Log)
		}
		request.Limit = limit
	}

	// Add timeout to context
	timeoutCtx, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()

	// Call the use case to list warehouses
	warehouseResponse, err := c.UseCase.ListWarehouses(timeoutCtx, request)
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"page":       request.Page,
			"limit":      request.Limit,
			"error":      err.Error(),
		}).Warn("Failed to list warehouses")

		// Handle specific error types
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			return response.JSONError(ctx, appErr, c.Log)
		}

		return response.JSONError(ctx, appErrors.WithError(appErrors.ErrInternalServer, err), c.Log)
	}

	return response.JSONSuccess(ctx, warehouseResponse)
}