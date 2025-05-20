package usecase

import (
	"context"
	"errors"
	"shop-service/internal/entity"
	appErrors "shop-service/internal/errors"
	"shop-service/internal/gateway"
	"shop-service/internal/model"
	"shop-service/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ShopUsecaseInterface defines the business logic methods for shop operations
type ShopUsecaseInterface interface {
	// ListShops retrieves a paginated list of shops with optional filtering
	ListShops(page, pageSize int, searchTerm string, includeInactive bool) ([]entity.Shop, int64, error)

	// GetShopByID retrieves a shop by its ID
	GetShopByID(id uint) (*entity.Shop, error)

	// GetShopWithWarehouses retrieves a shop with its warehouses by ID
	GetShopWithWarehouses(id uint) (*entity.Shop, error)

	// GetShopWarehouses retrieves detailed warehouse information for a shop
	GetShopWarehouses(ctx context.Context, shopID uint) (*model.ShopWarehousesResponse, error)
	
	// CreateShop creates a new shop
	CreateShop(req *model.CreateShopRequest) (*entity.Shop, error)
}

// ShopUsecase implements ShopUsecaseInterface
type ShopUsecase struct {
	DB                *gorm.DB
	Log               *logrus.Logger
	Validate          *validator.Validate
	ShopRepo          repository.ShopRepositoryInterface
	ShopWarehouseRepo repository.ShopWarehouseRepositoryInterface
	WarehouseGateway  gateway.WarehouseGatewayInterface
}

// NewShopUsecase creates a new shop usecase instance
func NewShopUsecase(
	db *gorm.DB,
	log *logrus.Logger,
	validate *validator.Validate,
	shopRepo repository.ShopRepositoryInterface,
	shopWarehouseRepo repository.ShopWarehouseRepositoryInterface,
	warehouseGateway gateway.WarehouseGatewayInterface,
) ShopUsecaseInterface {
	return &ShopUsecase{
		DB:                db,
		Log:               log,
		Validate:          validate,
		ShopRepo:          shopRepo,
		ShopWarehouseRepo: shopWarehouseRepo,
		WarehouseGateway:  warehouseGateway,
	}
}

// ListShops retrieves a paginated list of shops with optional filtering
func (u *ShopUsecase) ListShops(page, pageSize int, searchTerm string, includeInactive bool) ([]entity.Shop, int64, error) {
	// Validate page and pageSize
	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Get the shops from repository
	shops, totalCount, err := u.ShopRepo.FindAll(u.DB, page, pageSize, searchTerm, includeInactive)
	if err != nil {
		u.Log.WithError(err).Error("Failed to list shops")
		return nil, 0, err
	}

	return shops, totalCount, nil
}

// GetShopByID retrieves a shop by its ID
func (u *ShopUsecase) GetShopByID(id uint) (*entity.Shop, error) {
	if id == 0 {
		return nil, appErrors.ErrInvalidInput
	}

	shop, err := u.ShopRepo.FindByID(u.DB, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrShopNotFound
		}
		u.Log.WithError(err).Error("Failed to get shop by ID")
		return nil, appErrors.WithError(appErrors.ErrInternalServer, err)
	}

	return shop, nil
}

// GetShopWithWarehouses retrieves a shop with its warehouses by ID
func (u *ShopUsecase) GetShopWithWarehouses(id uint) (*entity.Shop, error) {
	if id == 0 {
		return nil, appErrors.ErrInvalidInput
	}

	shop, err := u.ShopRepo.FindByIDWithWarehouses(u.DB, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrShopNotFound
		}
		u.Log.WithError(err).Error("Failed to get shop with warehouses")
		return nil, appErrors.WithError(appErrors.ErrInternalServer, err)
	}

	return shop, nil
}

// GetShopWarehouses retrieves detailed warehouse information for a shop
func (u *ShopUsecase) GetShopWarehouses(ctx context.Context, shopID uint) (*model.ShopWarehousesResponse, error) {
	// Validate shopID
	if shopID == 0 {
		return nil, appErrors.ErrInvalidInput
	}

	// First check if the shop exists
	shop, err := u.ShopRepo.FindByID(u.DB, shopID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrShopNotFound
		}
		u.Log.WithError(err).Error("Failed to check if shop exists")
		return nil, appErrors.WithError(appErrors.ErrInternalServer, err)
	}

	// Get warehouse IDs for the specified shop using the dedicated repository
	warehouseIDs, err := u.ShopWarehouseRepo.FindWarehouseIDsByShopID(u.DB, shopID)
	if err != nil {
		u.Log.WithError(err).Error("Failed to get warehouse IDs for shop")
		return nil, appErrors.WithError(appErrors.ErrInternalServer, err)
	}

	// Prepare response
	response := &model.ShopWarehousesResponse{
		ShopID:     shop.ID,
		Warehouses: []model.WarehouseResponse{},
	}

	// If no warehouses associated, return empty response
	if len(warehouseIDs) == 0 {
		return response, nil
	}

	// Fetch detailed warehouse information for each warehouse ID
	warehouses := make([]model.WarehouseResponse, 0, len(warehouseIDs))

	for _, warehouseID := range warehouseIDs {
		warehouse, err := u.WarehouseGateway.GetWarehouseByID(ctx, warehouseID)
		if err != nil {
			// Log error but continue to fetch other warehouses if possible
			u.Log.WithFields(logrus.Fields{
				"error":        err.Error(),
				"shop_id":      shopID,
				"warehouse_id": warehouseID,
			}).Warn("Failed to get warehouse information from warehouse service")

			// Skip this warehouse and continue with others if it's a not found error
			if errors.Is(err, appErrors.ErrWarehouseNotFound) {
				continue
			}

			// For other errors, return the error
			return nil, err
		}

		// Add warehouse to the list
		warehouses = append(warehouses, *warehouse)
	}

	// Update response with warehouses
	response.Warehouses = warehouses

	return response, nil
}

// CreateShop creates a new shop
func (u *ShopUsecase) CreateShop(req *model.CreateShopRequest) (*entity.Shop, error) {
	// Validate request
	if err := u.Validate.Struct(req); err != nil {
		u.Log.WithError(err).Error("Invalid shop creation request")
		return nil, appErrors.WithError(appErrors.ErrInvalidInput, err)
	}

	// Create entity from request
	shop := &entity.Shop{
		Name:         req.Name,
		Description:  req.Description,
		Address:      req.Address,
		ContactEmail: req.ContactEmail,
		ContactPhone: req.ContactPhone,
		IsActive:     req.IsActive,
	}

	// Begin transaction
	tx := u.DB.Begin()
	if tx.Error != nil {
		u.Log.WithError(tx.Error).Error("Failed to begin transaction")
		return nil, appErrors.WithError(appErrors.ErrInternalServer, tx.Error)
	}

	// Create shop in database
	if err := u.ShopRepo.Create(tx, shop); err != nil {
		tx.Rollback()
		u.Log.WithError(err).Error("Failed to create shop")
		return nil, appErrors.WithError(appErrors.ErrInternalServer, err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		u.Log.WithError(err).Error("Failed to commit transaction")
		return nil, appErrors.WithError(appErrors.ErrInternalServer, err)
	}

	return shop, nil
}
