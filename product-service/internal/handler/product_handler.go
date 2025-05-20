package handler

import (
	"product-service/internal/context"
	"product-service/internal/delivery/http/response"
	"product-service/internal/errors"
	"product-service/internal/model"
	"product-service/internal/usecase"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type ProductHandler struct {
	Log     *logrus.Logger
	UseCase usecase.ProductUseCaseInterface
}

func NewProductHandler(useCase usecase.ProductUseCaseInterface, logger *logrus.Logger) *ProductHandler {
	return &ProductHandler{
		Log:     logger,
		UseCase: useCase,
	}
}

// GetProducts godoc
// @Summary Get a list of products
// @Description Get a list of products with pagination
// @Tags products
// @Accept json
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} model.ProductListResponseWrapper
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /products [get]
func (h *ProductHandler) GetProducts(ctx *fiber.Ctx) error {
	// Get request ID for logging context
	requestID := ctx.Get("X-Request-ID")

	// Parse query parameters for pagination
	limitStr := ctx.Query("limit", "10")
	offsetStr := ctx.Query("offset", "0")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"param":      "limit",
			"value":      limitStr,
			"error":      err.Error(),
		}).Warn("Invalid limit parameter")
		
		// Default to 10 if invalid
		limit = 10
	}
	
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"param":      "offset",
			"value":      offsetStr,
			"error":      err.Error(),
		}).Warn("Invalid offset parameter")
		
		// Default to 0 if invalid
		offset = 0
	}

	// Create context with request ID and timeout
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)
	ctxWithTimeout, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()
	
	// Get products from usecase
	products, err := h.UseCase.GetProducts(ctxWithTimeout, limit, offset)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"limit":      limit,
			"offset":     offset,
			"error":      err.Error(),
		}).Warn("Failed to get products")
		
		return response.HandleError(ctx, err, h.Log)
	}
	
	return response.JSONSuccess(ctx, products)
}

// GetProductByID godoc
// @Summary Get a single product by ID
// @Description Get a single product by ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} model.ProductResponseWrapper
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /products/{id} [get]
func (h *ProductHandler) GetProductByID(ctx *fiber.Ctx) error {
	// Get request ID for logging context
	requestID := ctx.Get("X-Request-ID")
	
	// Get ID from URL parameter
	id := ctx.Params("id")
	if id == "" {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      "missing product ID",
		}).Warn("Invalid request: missing product ID")
		
		return response.JSONError(ctx, errors.ErrInvalidProductID, h.Log)
	}
	
	// Create context with request ID and timeout
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)
	ctxWithTimeout, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()
	
	// Get product from usecase
	product, err := h.UseCase.GetProductByID(ctxWithTimeout, id)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"product_id": id,
			"error":      err.Error(),
		}).Warn("Failed to get product by ID")
		
		// Map common errors to application errors
		if err == fiber.ErrNotFound {
			return response.JSONError(ctx, errors.ErrProductNotFound, h.Log)
		}
		
		return response.HandleError(ctx, err, h.Log)
	}
	
	return response.JSONSuccess(ctx, product)
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product
// @Tags products
// @Accept json
// @Produce json
// @Param product body model.CreateProductRequest true "Product data"
// @Success 200 {object} model.ProductResponseWrapper
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /products [post]
func (h *ProductHandler) CreateProduct(ctx *fiber.Ctx) error {
	// Get request ID for logging context
	requestID := ctx.Get("X-Request-ID")
	
	// Parse request body
	request := new(model.CreateProductRequest)
	if err := ctx.BodyParser(request); err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to parse request body")
		
		return response.JSONError(ctx, errors.WithError(errors.ErrInvalidInput, err), h.Log)
	}
	
	// Create context with request ID and timeout
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)
	ctxWithTimeout, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()
	
	// Create product using usecase
	product, err := h.UseCase.CreateProduct(ctxWithTimeout, request)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"name":       request.Name,
			"sku":        request.SKU,
			"error":      err.Error(),
		}).Warn("Failed to create product")
		
		return response.HandleError(ctx, err, h.Log)
	}
	
	// Return created product with 201 status
	return response.JSONSuccess(ctx, product)
}

// UpdateProduct godoc
// @Summary Update an existing product
// @Description Update an existing product
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body model.UpdateProductRequest true "Product data"
// @Success 200 {object} model.ProductResponseWrapper
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /products/{id} [put]
func (h *ProductHandler) UpdateProduct(ctx *fiber.Ctx) error {
	// Get request ID for logging context
	requestID := ctx.Get("X-Request-ID")
	
	// Get ID from URL parameter
	id := ctx.Params("id")
	if id == "" {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      "missing product ID",
		}).Warn("Invalid request: missing product ID")
		
		return response.JSONError(ctx, errors.ErrInvalidProductID, h.Log)
	}
	
	// Parse request body
	request := new(model.UpdateProductRequest)
	if err := ctx.BodyParser(request); err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"product_id": id,
			"error":      err.Error(),
		}).Warn("Failed to parse request body")
		
		return response.JSONError(ctx, errors.WithError(errors.ErrInvalidInput, err), h.Log)
	}
	
	// Create context with request ID and timeout
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)
	ctxWithTimeout, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()
	
	// Update product using usecase
	product, err := h.UseCase.UpdateProduct(ctxWithTimeout, id, request)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"product_id": id,
			"error":      err.Error(),
		}).Warn("Failed to update product")
		
		return response.HandleError(ctx, err, h.Log)
	}
	
	return response.JSONSuccess(ctx, product)
}

// DeleteProduct godoc
// @Summary Delete a product
// @Description Delete a product
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 204 "No Content"
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /products/{id} [delete]
func (h *ProductHandler) DeleteProduct(ctx *fiber.Ctx) error {
	// Get request ID for logging context
	requestID := ctx.Get("X-Request-ID")
	
	// Get ID from URL parameter
	id := ctx.Params("id")
	if id == "" {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      "missing product ID",
		}).Warn("Invalid request: missing product ID")
		
		return response.JSONError(ctx, errors.ErrInvalidProductID, h.Log)
	}
	
	// Create context with request ID and timeout
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)
	ctxWithTimeout, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()
	
	// Delete product using usecase
	err := h.UseCase.DeleteProduct(ctxWithTimeout, id)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"product_id": id,
			"error":      err.Error(),
		}).Warn("Failed to delete product")
		
		return response.HandleError(ctx, err, h.Log)
	}
	
	// Return success with 204 status (No Content)
	return ctx.Status(fiber.StatusNoContent).Send(nil)
}

// SearchProducts godoc
// @Summary Search for products
// @Description Search for products based on a query string
// @Tags products
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} model.ProductListResponseWrapper
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /products/search [get]
func (h *ProductHandler) SearchProducts(ctx *fiber.Ctx) error {
	// Get request ID for logging context
	requestID := ctx.Get("X-Request-ID")
	
	// Get query parameters
	query := ctx.Query("q")
	if query == "" {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      "missing search query",
		}).Warn("Invalid request: missing search query")
		
		return response.JSONError(ctx, errors.WithMessage(errors.ErrInvalidInput, "Search query is required"), h.Log)
	}
	
	// Parse pagination parameters
	limitStr := ctx.Query("limit", "10")
	offsetStr := ctx.Query("offset", "0")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"param":      "limit",
			"value":      limitStr,
			"error":      err.Error(),
		}).Warn("Invalid limit parameter")
		
		// Default to 10 if invalid
		limit = 10
	}
	
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"param":      "offset",
			"value":      offsetStr,
			"error":      err.Error(),
		}).Warn("Invalid offset parameter")
		
		// Default to 0 if invalid
		offset = 0
	}
	
	// Create context with request ID and timeout
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)
	ctxWithTimeout, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()
	
	// Search products using usecase
	products, err := h.UseCase.SearchProducts(ctxWithTimeout, query, limit, offset)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"query":      query,
			"limit":      limit,
			"offset":     offset,
			"error":      err.Error(),
		}).Warn("Failed to search products")
		
		return response.HandleError(ctx, err, h.Log)
	}
	
	return response.JSONSuccess(ctx, products)
}

// GetProductsByCategory godoc
// @Summary Get products by category
// @Description Get products filtered by category
// @Tags products
// @Accept json
// @Produce json
// @Param category path string true "Category"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} model.ProductListResponseWrapper
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /products/category/{category} [get]
func (h *ProductHandler) GetProductsByCategory(ctx *fiber.Ctx) error {
	// Get request ID for logging context
	requestID := ctx.Get("X-Request-ID")
	
	// Get category from parameter
	category := ctx.Params("category")
	if category == "" {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      "missing category",
		}).Warn("Invalid request: missing category")
		
		return response.JSONError(ctx, errors.WithMessage(errors.ErrInvalidInput, "Category is required"), h.Log)
	}
	
	// Parse pagination parameters
	limitStr := ctx.Query("limit", "10")
	offsetStr := ctx.Query("offset", "0")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"param":      "limit",
			"value":      limitStr,
			"error":      err.Error(),
		}).Warn("Invalid limit parameter")
		
		// Default to 10 if invalid
		limit = 10
	}
	
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"param":      "offset",
			"value":      offsetStr,
			"error":      err.Error(),
		}).Warn("Invalid offset parameter")
		
		// Default to 0 if invalid
		offset = 0
	}
	
	// Create context with request ID and timeout
	userCtx := context.WithRequestID(ctx.UserContext(), requestID)
	ctxWithTimeout, cancel := context.WithDefaultTimeout(userCtx)
	defer cancel()
	
	// Get products by category using usecase
	products, err := h.UseCase.GetProductsByCategory(ctxWithTimeout, category, limit, offset)
	if err != nil {
		h.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"category":   category,
			"limit":      limit,
			"offset":     offset,
			"error":      err.Error(),
		}).Warn("Failed to get products by category")
		
		return response.HandleError(ctx, err, h.Log)
	}
	
	return response.JSONSuccess(ctx, products)
}