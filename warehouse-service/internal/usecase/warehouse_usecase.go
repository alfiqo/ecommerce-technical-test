package usecase

import (
	"context"
	"errors"
	appErrors "warehouse-service/internal/errors"
	"warehouse-service/internal/model"
	"warehouse-service/internal/model/converter"
	"warehouse-service/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	defaultPage  = 1
	defaultLimit = 20
)

type WarehouseUseCaseInterface interface {
	GetWarehouse(ctx context.Context, id uint) (*model.WarehouseResponse, error)
	CreateWarehouse(ctx context.Context, request *model.CreateWarehouseRequest) (*model.WarehouseResponse, error)
	UpdateWarehouse(ctx context.Context, request *model.UpdateWarehouseRequest) (*model.WarehouseResponse, error)
	DeleteWarehouse(ctx context.Context, id uint) error
	ListWarehouses(ctx context.Context, request *model.ListWarehouseRequest) (*model.WarehouseListResponse, error)
}

type WarehouseUseCase struct {
	DB                 *gorm.DB
	Log                *logrus.Logger
	Validate           *validator.Validate
	WarehouseRepository repository.WarehouseRepositoryInterface
}

func NewWarehouseUseCase(
	db *gorm.DB,
	logger *logrus.Logger,
	validate *validator.Validate,
	warehouseRepository repository.WarehouseRepositoryInterface,
) WarehouseUseCaseInterface {
	return &WarehouseUseCase{
		DB:                 db,
		Log:                logger,
		Validate:           validate,
		WarehouseRepository: warehouseRepository,
	}
}

// GetWarehouse retrieves a warehouse by ID along with its statistics
func (c *WarehouseUseCase) GetWarehouse(ctx context.Context, id uint) (*model.WarehouseResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Find the warehouse
	warehouse, err := c.WarehouseRepository.FindByID(tx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrResourceNotFound
		}
		c.Log.WithError(err).Error("Failed to find warehouse")
		return nil, fiber.ErrInternalServerError
	}

	// Get warehouse statistics
	stats, err := c.getWarehouseStats(tx, warehouse.ID)
	if err != nil {
		c.Log.WithError(err).Error("Failed to get warehouse statistics")
		return nil, fiber.ErrInternalServerError
	}

	return converter.WarehouseToResponse(warehouse, stats), nil
}

// CreateWarehouse creates a new warehouse
func (c *WarehouseUseCase) CreateWarehouse(ctx context.Context, request *model.CreateWarehouseRequest) (*model.WarehouseResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Validate request
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Warn("Invalid request body")
		return nil, fiber.ErrBadRequest
	}

	// Convert request to entity
	warehouse := converter.WarehouseRequestToEntity(request)

	// Create the warehouse
	if err := c.WarehouseRepository.Create(tx, warehouse); err != nil {
		c.Log.WithError(err).Error("Failed to create warehouse")
		return nil, fiber.ErrInternalServerError
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("Failed to commit transaction")
		return nil, fiber.ErrInternalServerError
	}

	// Create empty stats for new warehouse
	stats := &model.WarehouseStatsDTO{
		TotalProducts: 0,
		TotalItems:    0,
	}

	return converter.WarehouseToResponse(warehouse, stats), nil
}

// UpdateWarehouse updates an existing warehouse
func (c *WarehouseUseCase) UpdateWarehouse(ctx context.Context, request *model.UpdateWarehouseRequest) (*model.WarehouseResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Validate request
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Warn("Invalid request body")
		return nil, fiber.ErrBadRequest
	}

	// Find the warehouse
	warehouse, err := c.WarehouseRepository.FindByID(tx, request.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrResourceNotFound
		}
		c.Log.WithError(err).Error("Failed to find warehouse")
		return nil, fiber.ErrInternalServerError
	}

	// Update the warehouse
	converter.UpdateWarehouseFromRequest(warehouse, request)
	if err := c.WarehouseRepository.Update(tx, warehouse); err != nil {
		c.Log.WithError(err).Error("Failed to update warehouse")
		return nil, fiber.ErrInternalServerError
	}

	// Get warehouse statistics
	stats, err := c.getWarehouseStats(tx, warehouse.ID)
	if err != nil {
		c.Log.WithError(err).Error("Failed to get warehouse statistics")
		return nil, fiber.ErrInternalServerError
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("Failed to commit transaction")
		return nil, fiber.ErrInternalServerError
	}

	return converter.WarehouseToResponse(warehouse, stats), nil
}

// DeleteWarehouse deletes a warehouse by ID
func (c *WarehouseUseCase) DeleteWarehouse(ctx context.Context, id uint) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Find the warehouse to ensure it exists
	_, err := c.WarehouseRepository.FindByID(tx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrResourceNotFound
		}
		c.Log.WithError(err).Error("Failed to find warehouse")
		return fiber.ErrInternalServerError
	}

	// Delete the warehouse
	if err := c.WarehouseRepository.Delete(tx, id); err != nil {
		c.Log.WithError(err).Error("Failed to delete warehouse")
		return fiber.ErrInternalServerError
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("Failed to commit transaction")
		return fiber.ErrInternalServerError
	}

	return nil
}

// ListWarehouses lists warehouses with pagination
func (c *WarehouseUseCase) ListWarehouses(ctx context.Context, request *model.ListWarehouseRequest) (*model.WarehouseListResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Set default pagination values if not provided
	page := defaultPage
	limit := defaultLimit
	if request != nil {
		if request.Page > 0 {
			page = request.Page
		}
		if request.Limit > 0 && request.Limit <= 100 {
			limit = request.Limit
		}
	}

	offset := (page - 1) * limit

	// Get warehouses
	warehouses, total, err := c.WarehouseRepository.List(tx, limit, offset)
	if err != nil {
		c.Log.WithError(err).Error("Failed to list warehouses")
		return nil, fiber.ErrInternalServerError
	}

	// Build response
	response := &model.WarehouseListResponse{
		Warehouses: make([]model.WarehouseResponse, 0, len(warehouses)),
		Total:      total,
		Page:       page,
		Limit:      limit,
	}

	// Get stats for each warehouse and build response
	for _, warehouse := range warehouses {
		stats, err := c.getWarehouseStats(tx, warehouse.ID)
		if err != nil {
			c.Log.WithError(err).Error("Failed to get warehouse statistics")
			continue
		}
		
		warehouseResponse := converter.WarehouseToResponse(&warehouse, stats)
		response.Warehouses = append(response.Warehouses, *warehouseResponse)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("Failed to commit transaction")
		return nil, fiber.ErrInternalServerError
	}

	return response, nil
}

// getWarehouseStats gets statistics for a warehouse
func (c *WarehouseUseCase) getWarehouseStats(tx *gorm.DB, warehouseID uint) (*model.WarehouseStatsDTO, error) {
	// Get product count
	productCount, err := c.WarehouseRepository.GetProductCount(tx, warehouseID)
	if err != nil {
		return nil, err
	}

	// Get total item count
	totalItems, err := c.WarehouseRepository.GetTotalItemCount(tx, warehouseID)
	if err != nil {
		return nil, err
	}

	return &model.WarehouseStatsDTO{
		TotalProducts: productCount,
		TotalItems:    totalItems,
	}, nil
}