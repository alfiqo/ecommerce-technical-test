package repository

import (
	"context"
	"fmt"
	"time"
	"warehouse-service/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReservationRepositoryInterface interface {
	// ReserveStock reserves stock with database locking to prevent race conditions
	ReserveStock(tx *gorm.DB, warehouseID, productID uint, quantity int) (*entity.WarehouseStock, error)
	
	// CancelReservation cancels a previously made reservation
	CancelReservation(tx *gorm.DB, warehouseID, productID uint, quantity int) error
	
	// CommitReservation converts a reservation to a confirmed withdrawal
	CommitReservation(tx *gorm.DB, warehouseID, productID uint, quantity int) error
	
	// CreateReservationLog logs a reservation event
	CreateReservationLog(tx *gorm.DB, warehouseID, productID uint, quantity int, status string, reference string) error
	
	// GetReservationLogs retrieves reservation logs for a product
	GetReservationLogs(tx *gorm.DB, warehouseID, productID uint, limit, offset int) ([]entity.ReservationLog, int64, error)
}

type ReservationRepository struct {
	DB  *gorm.DB
	Log *logrus.Logger
}

func NewReservationRepository(log *logrus.Logger, db *gorm.DB) ReservationRepositoryInterface {
	return &ReservationRepository{
		DB:  db,
		Log: log,
	}
}

// ReserveStock locks and reserves stock with pessimistic locking to prevent race conditions
func (r *ReservationRepository) ReserveStock(tx *gorm.DB, warehouseID, productID uint, quantity int) (*entity.WarehouseStock, error) {
	// First, lock the stock record with FOR UPDATE to prevent concurrent modifications
	stock := new(entity.WarehouseStock)
	result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).
		First(stock)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("stock not found for warehouseID %d and productID %d", warehouseID, productID)
		}
		return nil, result.Error
	}

	// Calculate available quantity
	stock.CalculateAvailableQuantity()

	// Check if there's enough available quantity to reserve
	if stock.AvailableQuantity < quantity {
		return nil, fmt.Errorf("insufficient stock: requested %d, available %d", quantity, stock.AvailableQuantity)
	}

	// Update the reserved quantity
	stock.ReservedQuantity += quantity

	// Save the updated stock
	if err := tx.Save(stock).Error; err != nil {
		return nil, err
	}

	// Recalculate available quantity after update
	stock.CalculateAvailableQuantity()

	return stock, nil
}

// CancelReservation cancels a previously made reservation by decreasing the reserved quantity
func (r *ReservationRepository) CancelReservation(tx *gorm.DB, warehouseID, productID uint, quantity int) error {
	// Set a short timeout for the query to prevent long-running locks
	queryTimeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	
	// Lock the stock record with a timeout to prevent deadlocks
	stock := new(entity.WarehouseStock)
	result := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).
		First(stock)

	if result.Error != nil {
		r.Log.WithError(result.Error).WithFields(logrus.Fields{
			"warehouse_id": warehouseID,
			"product_id": productID,
		}).Error("Failed to lock stock record for cancellation")
		return result.Error
	}

	// Check if the requested quantity can be canceled
	if stock.ReservedQuantity < quantity {
		r.Log.WithFields(logrus.Fields{
			"warehouse_id": warehouseID,
			"product_id": productID,
			"reserved": stock.ReservedQuantity,
			"requested": quantity,
		}).Warn("Cannot cancel more than reserved")
		return fmt.Errorf("cannot cancel more than reserved: reserved %d, cancel request %d", 
			stock.ReservedQuantity, quantity)
	}

	// Update the reserved quantity
	stock.ReservedQuantity -= quantity

	// Save the updated stock
	err := tx.Save(stock).Error
	if err != nil {
		r.Log.WithError(err).WithFields(logrus.Fields{
			"warehouse_id": warehouseID,
			"product_id": productID,
		}).Error("Failed to save stock after cancellation")
	}
	return err
}

// CommitReservation converts a reservation to a confirmed withdrawal by reducing both quantity and reserved quantity
func (r *ReservationRepository) CommitReservation(tx *gorm.DB, warehouseID, productID uint, quantity int) error {
	// Set a short timeout for the query to prevent long-running locks
	queryTimeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()
	
	// Lock the stock record
	stock := new(entity.WarehouseStock)
	result := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).
		First(stock)

	if result.Error != nil {
		r.Log.WithError(result.Error).WithFields(logrus.Fields{
			"warehouse_id": warehouseID,
			"product_id": productID,
		}).Error("Failed to lock stock record for commit")
		return result.Error
	}

	// Check if there's enough reserved quantity to commit
	if stock.ReservedQuantity < quantity {
		r.Log.WithFields(logrus.Fields{
			"warehouse_id": warehouseID,
			"product_id": productID,
			"reserved": stock.ReservedQuantity,
			"requested": quantity,
		}).Warn("Cannot commit more than reserved")
		return fmt.Errorf("cannot commit more than reserved: reserved %d, commit request %d",
			stock.ReservedQuantity, quantity)
	}

	// Update the reserved and total quantity
	stock.ReservedQuantity -= quantity
	stock.Quantity -= quantity

	// Save the updated stock
	err := tx.Save(stock).Error
	if err != nil {
		r.Log.WithError(err).WithFields(logrus.Fields{
			"warehouse_id": warehouseID,
			"product_id": productID,
		}).Error("Failed to save stock after commit")
	}
	return err
}

// CreateReservationLog logs a reservation event
func (r *ReservationRepository) CreateReservationLog(tx *gorm.DB, warehouseID, productID uint, quantity int, status string, reference string) error {
	log := entity.ReservationLog{
		WarehouseID: warehouseID,
		ProductID:   productID,
		Quantity:    quantity,
		Status:      status,
		Reference:   reference,
		CreatedAt:   time.Now(),
	}

	return tx.Create(&log).Error
}

// GetReservationLogs retrieves reservation logs for a product
func (r *ReservationRepository) GetReservationLogs(tx *gorm.DB, warehouseID, productID uint, limit, offset int) ([]entity.ReservationLog, int64, error) {
	var logs []entity.ReservationLog
	var count int64

	// Count the total records
	err := tx.Model(&entity.ReservationLog{}).
		Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).
		Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	// Query with pagination
	query := tx.Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	// Execute the query
	err = query.Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, count, nil
}