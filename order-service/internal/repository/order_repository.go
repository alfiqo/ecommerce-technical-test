package repository

import (
	"order-service/internal/entity"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type OrderRepositoryInterface interface {
	CreateOrder(tx *gorm.DB, order *entity.Order) error
	CreateOrderItems(tx *gorm.DB, items []entity.OrderItem) error
	FindOrderByID(tx *gorm.DB, orderID uint) (*entity.Order, error)
	FindOrdersByUserID(tx *gorm.DB, userID string, page, limit int) ([]entity.Order, int64, error)
	FindOrdersByStatus(tx *gorm.DB, status entity.OrderStatus, page, limit int) ([]entity.Order, int64, error)
	UpdateOrderStatus(tx *gorm.DB, orderID uint, status entity.OrderStatus) error
	FindExpiredOrders(tx *gorm.DB, deadline time.Time) ([]entity.Order, error)
}

type OrderRepository struct {
	DB  *gorm.DB
	Log *logrus.Logger
}

func NewOrderRepository(log *logrus.Logger, db *gorm.DB) OrderRepositoryInterface {
	return &OrderRepository{
		DB:  db,
		Log: log,
	}
}

func (r *OrderRepository) CreateOrder(tx *gorm.DB, order *entity.Order) error {
	return tx.Create(order).Error
}

func (r *OrderRepository) CreateOrderItems(tx *gorm.DB, items []entity.OrderItem) error {
	return tx.Create(&items).Error
}

func (r *OrderRepository) FindOrderByID(tx *gorm.DB, orderID uint) (*entity.Order, error) {
	order := new(entity.Order)
	if err := tx.Preload("OrderItems").Where("id = ?", orderID).First(order).Error; err != nil {
		return nil, err
	}
	return order, nil
}

func (r *OrderRepository) FindOrdersByUserID(tx *gorm.DB, userID string, page, limit int) ([]entity.Order, int64, error) {
	var orders []entity.Order
	var total int64
	
	offset := (page - 1) * limit
	
	// Count total matching records
	err := tx.Model(&entity.Order{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// Get paginated data
	err = tx.Preload("OrderItems").Where("user_id = ?", userID).
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&orders).Error
	
	if err != nil {
		return nil, 0, err
	}
	
	return orders, total, nil
}

func (r *OrderRepository) FindOrdersByStatus(tx *gorm.DB, status entity.OrderStatus, page, limit int) ([]entity.Order, int64, error) {
	var orders []entity.Order
	var total int64
	
	offset := (page - 1) * limit
	
	// Count total matching records
	err := tx.Model(&entity.Order{}).Where("status = ?", status).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// Get paginated data
	err = tx.Preload("OrderItems").Where("status = ?", status).
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&orders).Error
	
	if err != nil {
		return nil, 0, err
	}
	
	return orders, total, nil
}

func (r *OrderRepository) UpdateOrderStatus(tx *gorm.DB, orderID uint, status entity.OrderStatus) error {
	return tx.Model(&entity.Order{}).Where("id = ?", orderID).Update("status", status).Error
}

func (r *OrderRepository) FindExpiredOrders(tx *gorm.DB, deadline time.Time) ([]entity.Order, error) {
	var orders []entity.Order
	
	err := tx.Preload("OrderItems").
		Where("status = ? AND payment_deadline < ?", entity.OrderStatusPending, deadline).
		Find(&orders).Error
	
	if err != nil {
		return nil, err
	}
	
	return orders, nil
}