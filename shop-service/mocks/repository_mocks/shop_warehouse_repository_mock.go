package repository_mocks

import (
	"shop-service/internal/entity"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// ShopWarehouseRepositoryMock is a mock for ShopWarehouseRepositoryInterface
type ShopWarehouseRepositoryMock struct {
	mock.Mock
}

// FindByShopID mocks the FindByShopID method
func (m *ShopWarehouseRepositoryMock) FindByShopID(db *gorm.DB, shopID uint) ([]entity.ShopWarehouse, error) {
	args := m.Called(db, shopID)
	
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	
	return args.Get(0).([]entity.ShopWarehouse), args.Error(1)
}

// FindWarehouseIDsByShopID mocks the FindWarehouseIDsByShopID method
func (m *ShopWarehouseRepositoryMock) FindWarehouseIDsByShopID(db *gorm.DB, shopID uint) ([]uint, error) {
	args := m.Called(db, shopID)
	
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	
	return args.Get(0).([]uint), args.Error(1)
}

// AssignWarehouseToShop mocks the AssignWarehouseToShop method
func (m *ShopWarehouseRepositoryMock) AssignWarehouseToShop(db *gorm.DB, shopID, warehouseID uint) error {
	args := m.Called(db, shopID, warehouseID)
	return args.Error(0)
}

// RemoveWarehouseFromShop mocks the RemoveWarehouseFromShop method
func (m *ShopWarehouseRepositoryMock) RemoveWarehouseFromShop(db *gorm.DB, shopID, warehouseID uint) error {
	args := m.Called(db, shopID, warehouseID)
	return args.Error(0)
}