package repository_mock

import (
	"order-service/internal/entity"
	"time"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// OrderRepositoryMock is a mock implementation of the OrderRepositoryInterface
type OrderRepositoryMock struct {
	mock.Mock
}

// CreateOrder mocks the CreateOrder method
func (m *OrderRepositoryMock) CreateOrder(tx *gorm.DB, order *entity.Order) error {
	args := m.Called(tx, order)
	return args.Error(0)
}

// CreateOrderItems mocks the CreateOrderItems method
func (m *OrderRepositoryMock) CreateOrderItems(tx *gorm.DB, items []entity.OrderItem) error {
	args := m.Called(tx, items)
	return args.Error(0)
}

// FindOrderByID mocks the FindOrderByID method
func (m *OrderRepositoryMock) FindOrderByID(tx *gorm.DB, orderID uint) (*entity.Order, error) {
	args := m.Called(tx, orderID)
	
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	
	return args.Get(0).(*entity.Order), args.Error(1)
}

// FindOrdersByUserID mocks the FindOrdersByUserID method
func (m *OrderRepositoryMock) FindOrdersByUserID(tx *gorm.DB, userID string, page, limit int) ([]entity.Order, int64, error) {
	args := m.Called(tx, userID, page, limit)
	
	return args.Get(0).([]entity.Order), args.Get(1).(int64), args.Error(2)
}

// FindOrdersByStatus mocks the FindOrdersByStatus method
func (m *OrderRepositoryMock) FindOrdersByStatus(tx *gorm.DB, status entity.OrderStatus, page, limit int) ([]entity.Order, int64, error) {
	args := m.Called(tx, status, page, limit)
	
	return args.Get(0).([]entity.Order), args.Get(1).(int64), args.Error(2)
}

// UpdateOrderStatus mocks the UpdateOrderStatus method
func (m *OrderRepositoryMock) UpdateOrderStatus(tx *gorm.DB, orderID uint, status entity.OrderStatus) error {
	args := m.Called(tx, orderID, status)
	return args.Error(0)
}

// FindExpiredOrders mocks the FindExpiredOrders method
func (m *OrderRepositoryMock) FindExpiredOrders(tx *gorm.DB, deadline time.Time) ([]entity.Order, error) {
	args := m.Called(tx, deadline)
	
	return args.Get(0).([]entity.Order), args.Error(1)
}