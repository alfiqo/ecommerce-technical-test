package repository

import (
	"shop-service/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ShopWarehouseRepositoryInterface defines the methods for shop_warehouses repository
type ShopWarehouseRepositoryInterface interface {
	// FindByShopID finds all warehouses associated with a shop
	FindByShopID(db *gorm.DB, shopID uint) ([]entity.ShopWarehouse, error)
	
	// FindWarehouseIDsByShopID finds all warehouse IDs associated with a shop
	FindWarehouseIDsByShopID(db *gorm.DB, shopID uint) ([]uint, error)
	
	// AssignWarehouseToShop adds a warehouse to a shop
	AssignWarehouseToShop(db *gorm.DB, shopID, warehouseID uint) error
	
	// RemoveWarehouseFromShop removes a warehouse from a shop
	RemoveWarehouseFromShop(db *gorm.DB, shopID, warehouseID uint) error
}

// ShopWarehouseRepository implements ShopWarehouseRepositoryInterface
type ShopWarehouseRepository struct {
	Log *logrus.Logger
}

// NewShopWarehouseRepository creates a new shop warehouse repository instance
func NewShopWarehouseRepository(log *logrus.Logger) ShopWarehouseRepositoryInterface {
	return &ShopWarehouseRepository{
		Log: log,
	}
}

// FindByShopID finds all warehouses associated with a shop
func (r *ShopWarehouseRepository) FindByShopID(db *gorm.DB, shopID uint) ([]entity.ShopWarehouse, error) {
	var shopWarehouses []entity.ShopWarehouse
	
	// Check if shop exists
	var shopExists bool
	err := db.Model(&entity.Shop{}).
		Select("COUNT(*) > 0").
		Where("id = ?", shopID).
		Find(&shopExists).
		Error
	
	if err != nil {
		r.Log.WithError(err).Error("Failed to check if shop exists")
		return nil, err
	}
	
	if !shopExists {
		return nil, gorm.ErrRecordNotFound
	}
	
	// Query shop_warehouses for the shop ID
	err = db.
		Where("shop_id = ?", shopID).
		Find(&shopWarehouses).
		Error
	
	if err != nil {
		r.Log.WithError(err).Error("Failed to find warehouse associations for shop")
		return nil, err
	}
	
	return shopWarehouses, nil
}

// FindWarehouseIDsByShopID finds all warehouse IDs associated with a shop
func (r *ShopWarehouseRepository) FindWarehouseIDsByShopID(db *gorm.DB, shopID uint) ([]uint, error) {
	// First check if the shop exists
	shopWarehouses, err := r.FindByShopID(db, shopID)
	if err != nil {
		return nil, err
	}
	
	// Extract warehouse IDs
	warehouseIDs := make([]uint, 0, len(shopWarehouses))
	for _, sw := range shopWarehouses {
		warehouseIDs = append(warehouseIDs, sw.WarehouseID)
	}
	
	return warehouseIDs, nil
}

// AssignWarehouseToShop adds a warehouse to a shop
func (r *ShopWarehouseRepository) AssignWarehouseToShop(db *gorm.DB, shopID, warehouseID uint) error {
	// Check if the association already exists
	var count int64
	err := db.Model(&entity.ShopWarehouse{}).
		Where("shop_id = ? AND warehouse_id = ?", shopID, warehouseID).
		Count(&count).
		Error
	
	if err != nil {
		r.Log.WithError(err).Error("Failed to check existing shop-warehouse association")
		return err
	}
	
	// If association already exists, return without error
	if count > 0 {
		return nil
	}
	
	// Create new association
	shopWarehouse := entity.ShopWarehouse{
		ShopID:      shopID,
		WarehouseID: warehouseID,
	}
	
	err = db.Create(&shopWarehouse).Error
	if err != nil {
		r.Log.WithError(err).Error("Failed to create shop-warehouse association")
		return err
	}
	
	return nil
}

// RemoveWarehouseFromShop removes a warehouse from a shop
func (r *ShopWarehouseRepository) RemoveWarehouseFromShop(db *gorm.DB, shopID, warehouseID uint) error {
	result := db.
		Where("shop_id = ? AND warehouse_id = ?", shopID, warehouseID).
		Delete(&entity.ShopWarehouse{})
	
	if result.Error != nil {
		r.Log.WithError(result.Error).Error("Failed to remove shop-warehouse association")
		return result.Error
	}
	
	return nil
}