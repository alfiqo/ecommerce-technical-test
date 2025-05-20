package handler

import (
	"strconv"
	"shop-service/internal/delivery/http/response"
	"shop-service/internal/model"
	"shop-service/internal/model/converter"
	"shop-service/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// ShopHandler handles HTTP requests for shop operations
type ShopHandler struct {
	ShopUsecase usecase.ShopUsecaseInterface
	Log         *logrus.Logger
}

// NewShopHandler creates a new shop handler instance
func NewShopHandler(shopUsecase usecase.ShopUsecaseInterface, log *logrus.Logger) *ShopHandler {
	return &ShopHandler{
		ShopUsecase: shopUsecase,
		Log:         log,
	}
}

// ListShops handles GET /shops to retrieve a paginated list of shops
// @Summary List shops
// @Description Get a paginated list of shops
// @Tags shops
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Param search query string false "Search term"
// @Param include_inactive query bool false "Include inactive shops"
// @Success 200 {object} response.Response{data=model.ShopListResponse}
// @Failure 400 {object} response.Response{error=response.ErrorInfo}
// @Failure 500 {object} response.Response{error=response.ErrorInfo}
// @Router /shops [get]
func (h *ShopHandler) ListShops(c *fiber.Ctx) error {
	// Extract query parameters
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	
	pageSize, err := strconv.Atoi(c.Query("page_size", "10"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	
	searchTerm := c.Query("search", "")
	
	includeInactive := false
	if c.Query("include_inactive") == "true" {
		includeInactive = true
	}
	
	// Get shops from use case
	shops, totalCount, err := h.ShopUsecase.ListShops(page, pageSize, searchTerm, includeInactive)
	if err != nil {
		h.Log.WithError(err).Error("Failed to list shops")
		return response.JSONError(c, err, h.Log)
	}
	
	// Convert to response model
	shopListResponse := converter.ToShopListResponse(shops, totalCount, page, pageSize)
	
	// Return JSON response
	return response.JSONSuccess(c, shopListResponse)
}

// GetShopByID handles GET /shops/:id to retrieve a single shop by ID with its warehouse references
// @Summary Get shop by ID
// @Description Get detailed shop information by ID including warehouse references
// @Tags shops
// @Accept json
// @Produce json
// @Param id path int true "Shop ID"
// @Success 200 {object} response.Response{data=model.ShopDetailResponse}
// @Failure 400 {object} response.Response{error=response.ErrorInfo}
// @Failure 404 {object} response.Response{error=response.ErrorInfo}
// @Failure 500 {object} response.Response{error=response.ErrorInfo}
// @Router /shops/{id} [get]
func (h *ShopHandler) GetShopByID(c *fiber.Ctx) error {
	// Parse shop ID from URL parameter
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.Log.WithError(err).Error("Invalid shop ID format")
		return response.HandleError(c, fiber.NewError(fiber.StatusBadRequest, "Invalid shop ID format"), h.Log)
	}

	// Get shop with warehouses from use case
	shop, err := h.ShopUsecase.GetShopWithWarehouses(uint(id))
	if err != nil {
		h.Log.WithError(err).Error("Failed to get shop")
		return response.JSONError(c, err, h.Log)
	}

	// Convert to response model
	shopResponse := converter.ToShopDetailResponse(shop)

	// Return JSON response
	return response.JSONSuccess(c, shopResponse)
}

// GetShopWarehouses handles GET /shops/:id/warehouses to retrieve warehouses for a shop
// @Summary Get shop warehouses
// @Description Get all warehouses associated with a shop
// @Tags shops
// @Accept json
// @Produce json
// @Param id path int true "Shop ID"
// @Success 200 {object} response.Response{data=model.ShopWarehousesResponse}
// @Failure 400 {object} response.Response{error=response.ErrorInfo}
// @Failure 404 {object} response.Response{error=response.ErrorInfo}
// @Failure 500 {object} response.Response{error=response.ErrorInfo}
// @Router /shops/{id}/warehouses [get]
func (h *ShopHandler) GetShopWarehouses(c *fiber.Ctx) error {
	// Parse shop ID from URL parameter
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.Log.WithError(err).Error("Invalid shop ID format")
		return response.HandleError(c, fiber.NewError(fiber.StatusBadRequest, "Invalid shop ID format"), h.Log)
	}

	// Create context from fiber context
	ctx := c.Context()

	// Get warehouses for shop from use case
	warehousesResponse, err := h.ShopUsecase.GetShopWarehouses(ctx, uint(id))
	if err != nil {
		h.Log.WithError(err).Error("Failed to get shop warehouses")
		return response.JSONError(c, err, h.Log)
	}

	// Return JSON response
	return response.JSONSuccess(c, warehousesResponse)
}

// CreateShop handles POST /shops to create a new shop
// @Summary Create a new shop
// @Description Create a new shop with the provided information
// @Tags shops
// @Accept json
// @Produce json
// @Param shop body model.CreateShopRequest true "Shop information"
// @Success 201 {object} response.Response{data=model.ShopResponse}
// @Failure 400 {object} response.Response{error=response.ErrorInfo}
// @Failure 500 {object} response.Response{error=response.ErrorInfo}
// @Router /shops [post]
func (h *ShopHandler) CreateShop(c *fiber.Ctx) error {
	// Parse request body
	var req model.CreateShopRequest
	if err := c.BodyParser(&req); err != nil {
		h.Log.WithError(err).Error("Failed to parse shop creation request")
		return response.HandleError(c, fiber.NewError(fiber.StatusBadRequest, "Invalid request format"), h.Log)
	}

	// Create shop through use case
	shop, err := h.ShopUsecase.CreateShop(&req)
	if err != nil {
		h.Log.WithError(err).Error("Failed to create shop")
		return response.JSONError(c, err, h.Log)
	}

	// Convert to response model
	shopResponse := converter.ToShopResponse(shop)

	// Return JSON response with 201 Created status
	return c.Status(fiber.StatusCreated).JSON(response.Response{
		Success: true,
		Data:    shopResponse,
	})
}