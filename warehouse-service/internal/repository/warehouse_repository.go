package repository

import (
	"warehouse-service/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type WarehouseRepositoryInterface interface {
	// Warehouse operations
	FindByID(db *gorm.DB, id uint) (*entity.Warehouse, error)
	Create(db *gorm.DB, warehouse *entity.Warehouse) error
	Update(db *gorm.DB, warehouse *entity.Warehouse) error
	Delete(db *gorm.DB, id uint) error
	List(db *gorm.DB, limit, offset int) ([]entity.Warehouse, int64, error)
	
	// Stock operations
	GetProductCount(db *gorm.DB, warehouseID uint) (int64, error)
	GetTotalItemCount(db *gorm.DB, warehouseID uint) (int64, error)
	GetWarehouseStock(db *gorm.DB, warehouseID uint, productID uint) (*entity.WarehouseStock, error)
	ListWarehouseStock(db *gorm.DB, warehouseID uint, limit, offset int) ([]entity.WarehouseStock, int64, error)
	UpdateStock(db *gorm.DB, stock *entity.WarehouseStock) error
}

type WarehouseRepository struct {
	DB  *gorm.DB
	Log *logrus.Logger
}

func NewWarehouseRepository(log *logrus.Logger, db *gorm.DB) WarehouseRepositoryInterface {
	return &WarehouseRepository{
		DB:  db,
		Log: log,
	}
}

// FindByID retrieves a warehouse by ID
func (r *WarehouseRepository) FindByID(db *gorm.DB, id uint) (*entity.Warehouse, error) {
	warehouse := new(entity.Warehouse)
	if err := db.Where("id = ?", id).Limit(1).First(warehouse).Error; err != nil {
		return nil, err
	}
	return warehouse, nil
}

// Create creates a new warehouse
func (r *WarehouseRepository) Create(db *gorm.DB, warehouse *entity.Warehouse) error {
	return db.Create(warehouse).Error
}

// Update updates an existing warehouse
func (r *WarehouseRepository) Update(db *gorm.DB, warehouse *entity.Warehouse) error {
	return db.Save(warehouse).Error
}

// Delete deletes a warehouse by ID
func (r *WarehouseRepository) Delete(db *gorm.DB, id uint) error {
	return db.Delete(&entity.Warehouse{}, id).Error
}

// List retrieves warehouses with pagination
func (r *WarehouseRepository) List(db *gorm.DB, limit, offset int) ([]entity.Warehouse, int64, error) {
	var warehouses []entity.Warehouse
	var count int64
	
	err := db.Model(&entity.Warehouse{}).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	
	if limit > 0 {
		err = db.Limit(limit).Offset(offset).Find(&warehouses).Error
	} else {
		err = db.Find(&warehouses).Error
	}
	
	if err != nil {
		return nil, 0, err
	}
	
	return warehouses, count, nil
}

// GetProductCount returns the count of unique products in a warehouse
func (r *WarehouseRepository) GetProductCount(db *gorm.DB, warehouseID uint) (int64, error) {
	var count int64
	err := db.Model(&entity.WarehouseStock{}).
		Where("warehouse_id = ?", warehouseID).
		Count(&count).Error
	return count, err
}

// GetTotalItemCount returns the sum of all items in a warehouse
func (r *WarehouseRepository) GetTotalItemCount(db *gorm.DB, warehouseID uint) (int64, error) {
	var total int64
	err := db.Model(&entity.WarehouseStock{}).
		Where("warehouse_id = ?", warehouseID).
		Select("COALESCE(SUM(quantity), 0) as total_items").
		Pluck("total_items", &total).Error
	return total, err
}

// GetWarehouseStock retrieves stock for a specific product in a warehouse
func (r *WarehouseRepository) GetWarehouseStock(db *gorm.DB, warehouseID uint, productID uint) (*entity.WarehouseStock, error) {
	stock := new(entity.WarehouseStock)
	if err := db.Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).First(stock).Error; err != nil {
		return nil, err
	}
	
	// Calculate virtual field
	stock.CalculateAvailableQuantity()
	return stock, nil
}

// ListWarehouseStock retrieves all stock items for a warehouse with pagination
func (r *WarehouseRepository) ListWarehouseStock(db *gorm.DB, warehouseID uint, limit, offset int) ([]entity.WarehouseStock, int64, error) {
	var stocks []entity.WarehouseStock
	var count int64
	
	err := db.Model(&entity.WarehouseStock{}).
		Where("warehouse_id = ?", warehouseID).
		Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	
	query := db.Where("warehouse_id = ?", warehouseID)
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}
	
	err = query.Find(&stocks).Error
	if err != nil {
		return nil, 0, err
	}
	
	// Calculate virtual fields
	for i := range stocks {
		stocks[i].CalculateAvailableQuantity()
	}
	
	return stocks, count, nil
}

// UpdateStock updates warehouse stock
func (r *WarehouseRepository) UpdateStock(db *gorm.DB, stock *entity.WarehouseStock) error {
	return db.Save(stock).Error
}