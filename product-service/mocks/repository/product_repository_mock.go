package repository

import (
	"product-service/internal/entity"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(db *gorm.DB, product *entity.Product) error {
	args := m.Called(db, product)
	return args.Error(0)
}

func (m *MockProductRepository) FindAll(db *gorm.DB, limit, offset int) ([]entity.Product, int64, error) {
	args := m.Called(db, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) FindByID(db *gorm.DB, id string) (*entity.Product, error) {
	args := m.Called(db, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *MockProductRepository) Update(db *gorm.DB, product *entity.Product) error {
	args := m.Called(db, product)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(db *gorm.DB, id string) error {
	args := m.Called(db, id)
	return args.Error(0)
}

func (m *MockProductRepository) GetDB() *gorm.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*gorm.DB)
}

func (m *MockProductRepository) Search(db *gorm.DB, query string, limit, offset int) ([]entity.Product, int64, error) {
	args := m.Called(db, query, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) FindByCategory(db *gorm.DB, category string, limit, offset int) ([]entity.Product, int64, error) {
	args := m.Called(db, category, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) FindBySKU(db *gorm.DB, sku string) (*entity.Product, error) {
	args := m.Called(db, sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}