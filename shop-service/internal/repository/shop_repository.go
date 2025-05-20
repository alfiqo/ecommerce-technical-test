package repository

import (
	"shop-service/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ShopRepositoryInterface defines the methods for shop repository
type ShopRepositoryInterface interface {
	// FindAll retrieves a paginated list of shops
	FindAll(db *gorm.DB, page, pageSize int, searchTerm string, includeInactive bool) ([]entity.Shop, int64, error)
	
	// FindByID finds a shop by its ID
	FindByID(db *gorm.DB, id uint) (*entity.Shop, error)

	// FindByIDWithWarehouses finds a shop by its ID and includes its warehouses
	FindByIDWithWarehouses(db *gorm.DB, id uint) (*entity.Shop, error)
	
	// Create creates a new shop
	Create(db *gorm.DB, shop *entity.Shop) error
	
	// Update updates an existing shop
	Update(db *gorm.DB, shop *entity.Shop) error
	
	// Delete deletes a shop by its ID
	Delete(db *gorm.DB, id uint) error
}

// ShopRepository implements ShopRepositoryInterface
type ShopRepository struct {
	Log *logrus.Logger
}

// NewShopRepository creates a new shop repository instance
func NewShopRepository(log *logrus.Logger) ShopRepositoryInterface {
	return &ShopRepository{
		Log: log,
	}
}

// FindAll retrieves a paginated list of shops
func (r *ShopRepository) FindAll(db *gorm.DB, page, pageSize int, searchTerm string, includeInactive bool) ([]entity.Shop, int64, error) {
	var shops []entity.Shop
	var totalCount int64

	query := db
	
	// Apply filters
	if searchTerm != "" {
		searchPattern := "%" + searchTerm + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}
	
	// Filter by active status if needed
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}
	
	// Count total results
	err := query.Model(&entity.Shop{}).Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}
	
	// Apply pagination
	offset := (page - 1) * pageSize
	
	// Execute the query
	err = query.
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&shops).
		Error
	
	if err != nil {
		return nil, 0, err
	}
	
	return shops, totalCount, nil
}

// FindByID finds a shop by its ID
func (r *ShopRepository) FindByID(db *gorm.DB, id uint) (*entity.Shop, error) {
	var shop entity.Shop
	
	err := db.Where("id = ?", id).First(&shop).Error
	if err != nil {
		return nil, err
	}
	
	return &shop, nil
}

// FindByIDWithWarehouses finds a shop by its ID and includes its related shop_warehouses
func (r *ShopRepository) FindByIDWithWarehouses(db *gorm.DB, id uint) (*entity.Shop, error) {
	var shop entity.Shop
	
	err := db.
		Where("id = ?", id).
		First(&shop).
		Error
	
	if err != nil {
		return nil, err
	}
	
	// Load shop_warehouses separately without preload
	var shopWarehouses []entity.ShopWarehouse
	err = db.
		Where("shop_id = ?", id).
		Find(&shopWarehouses).
		Error
	
	if err != nil {
		return nil, err
	}
	
	shop.Warehouses = shopWarehouses
	
	return &shop, nil
}

// Create creates a new shop
func (r *ShopRepository) Create(db *gorm.DB, shop *entity.Shop) error {
	return db.Create(shop).Error
}

// Update updates an existing shop
func (r *ShopRepository) Update(db *gorm.DB, shop *entity.Shop) error {
	return db.Save(shop).Error
}

// Delete deletes a shop by its ID
func (r *ShopRepository) Delete(db *gorm.DB, id uint) error {
	return db.Delete(&entity.Shop{}, id).Error
}

