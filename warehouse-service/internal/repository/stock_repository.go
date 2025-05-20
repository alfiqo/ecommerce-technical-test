package repository

import (
	"fmt"
	"warehouse-service/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StockRepositoryInterface interface {
	// GetWarehouseStock retrieves stock in a warehouse with pagination
	GetWarehouseStock(tx *gorm.DB, warehouseID uint, productID uint, limit, offset int) ([]entity.WarehouseStock, int64, error)
	
	// AddStock adds stock to a warehouse
	AddStock(tx *gorm.DB, warehouseID, productID uint, productSKU string, quantity int, reference, notes string) (*entity.WarehouseStock, error)
	
	// TransferStock transfers stock between warehouses
	TransferStock(tx *gorm.DB, sourceWarehouseID, targetWarehouseID, productID uint, productSKU string, quantity int, reference string) (*entity.StockTransfer, error)
	
	// LogStockMovement records a stock movement
	LogStockMovement(tx *gorm.DB, warehouseID, productID uint, productSKU string, movementType entity.MovementType, quantity int, referenceType, referenceID, notes string) error
	
	// GetStock gets a single stock record with locking if requested
	GetStock(tx *gorm.DB, warehouseID, productID uint, forUpdate bool) (*entity.WarehouseStock, error)
}

type StockRepository struct {
	DB  *gorm.DB
	Log *logrus.Logger
}

func NewStockRepository(log *logrus.Logger, db *gorm.DB) StockRepositoryInterface {
	return &StockRepository{
		DB:  db,
		Log: log,
	}
}

// GetWarehouseStock retrieves stock in a warehouse with pagination
func (r *StockRepository) GetWarehouseStock(tx *gorm.DB, warehouseID uint, productID uint, limit, offset int) ([]entity.WarehouseStock, int64, error) {
	var stocks []entity.WarehouseStock
	var count int64
	
	query := tx.Model(&entity.WarehouseStock{}).Where("warehouse_id = ?", warehouseID)
	
	if productID > 0 {
		query = query.Where("product_id = ?", productID)
	}
	
	// Count total records
	err := query.Count(&count).Error
	if err != nil {
		r.Log.WithError(err).Error("Failed to count warehouse stock")
		return nil, 0, err
	}
	
	// Get paginated records
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}
	
	err = query.Find(&stocks).Error
	if err != nil {
		r.Log.WithError(err).Error("Failed to get warehouse stock")
		return nil, 0, err
	}
	
	// Calculate available quantity for each stock
	for i := range stocks {
		stocks[i].CalculateAvailableQuantity()
	}
	
	return stocks, count, nil
}

// AddStock adds stock to a warehouse
func (r *StockRepository) AddStock(tx *gorm.DB, warehouseID, productID uint, productSKU string, quantity int, reference, notes string) (*entity.WarehouseStock, error) {
	// Get the stock with locking
	stock, err := r.GetStock(tx, warehouseID, productID, true)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	
	// If stock doesn't exist, create it
	if err == gorm.ErrRecordNotFound {
		stock = &entity.WarehouseStock{
			WarehouseID:      warehouseID,
			ProductID:        productID,
			Quantity:         quantity,
			ReservedQuantity: 0,
		}
		
		if err := tx.Create(stock).Error; err != nil {
			return nil, err
		}
	} else {
		// Update existing stock
		stock.Quantity += quantity
		if err := tx.Save(stock).Error; err != nil {
			return nil, err
		}
	}
	
	// Log the stock movement
	if err := r.LogStockMovement(tx, warehouseID, productID, productSKU, entity.MovementTypeStockIn, quantity, "manual", reference, notes); err != nil {
		return nil, err
	}
	
	// Calculate available quantity
	stock.CalculateAvailableQuantity()
	
	return stock, nil
}

// TransferStock transfers stock between warehouses
func (r *StockRepository) TransferStock(tx *gorm.DB, sourceWarehouseID, targetWarehouseID, productID uint, productSKU string, quantity int, reference string) (*entity.StockTransfer, error) {
	// Create transfer record
	transfer := &entity.StockTransfer{
		SourceWarehouseID: sourceWarehouseID,
		TargetWarehouseID: targetWarehouseID,
		ProductID:         productID,
		Quantity:          quantity,
		Status:            entity.StatusPending,
		TransferReference: reference,
	}
	
	if err := tx.Create(transfer).Error; err != nil {
		return nil, err
	}
	
	// Lock both source and target stocks consistently (by warehouse ID to prevent deadlocks)
	// Order the locks by warehouse ID to prevent deadlocks
	isSourceFirst := sourceWarehouseID < targetWarehouseID
	
	if !isSourceFirst {
		// First lock the warehouse with smaller ID to prevent deadlocks
		// This is a common deadlock prevention technique
	}
	
	// Get first stock with lock
	var firstStock, secondStock *entity.WarehouseStock
	var err error
	
	if isSourceFirst {
		firstStock, err = r.GetStock(tx, sourceWarehouseID, productID, true)
	} else {
		firstStock, err = r.GetStock(tx, targetWarehouseID, productID, true)
	}
	
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	
	// If first is source and not found or has insufficient stock
	if isSourceFirst && (err == gorm.ErrRecordNotFound || firstStock.Quantity-firstStock.ReservedQuantity < quantity) {
		transfer.Status = entity.StatusFailed
		tx.Save(transfer)
		return nil, fmt.Errorf("insufficient stock in source warehouse")
	}
	
	// Get second stock with lock
	if isSourceFirst {
		secondStock, err = r.GetStock(tx, targetWarehouseID, productID, true)
	} else {
		secondStock, err = r.GetStock(tx, sourceWarehouseID, productID, true)
	}
	
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	
	// If second is source and not found or has insufficient stock
	if !isSourceFirst && (err == gorm.ErrRecordNotFound || secondStock.Quantity-secondStock.ReservedQuantity < quantity) {
		transfer.Status = entity.StatusFailed
		tx.Save(transfer)
		return nil, fmt.Errorf("insufficient stock in source warehouse")
	}
	
	// Source and target are now identified
	var sourceStock, targetStock *entity.WarehouseStock
	if isSourceFirst {
		sourceStock = firstStock
		targetStock = secondStock
	} else {
		sourceStock = secondStock
		targetStock = firstStock
	}
	
	// Ensure source has sufficient stock
	if sourceStock.Quantity-sourceStock.ReservedQuantity < quantity {
		transfer.Status = entity.StatusFailed
		tx.Save(transfer)
		return nil, fmt.Errorf("insufficient stock in source warehouse")
	}
	
	// Decrease source stock
	sourceStock.Quantity -= quantity
	if err := tx.Save(sourceStock).Error; err != nil {
		return nil, err
	}
	
	// Create or update target stock
	if targetStock == nil || err == gorm.ErrRecordNotFound {
		targetStock = &entity.WarehouseStock{
			WarehouseID:      targetWarehouseID,
			ProductID:        productID,
			Quantity:         quantity,
			ReservedQuantity: 0,
		}
		if err := tx.Create(targetStock).Error; err != nil {
			return nil, err
		}
	} else {
		targetStock.Quantity += quantity
		if err := tx.Save(targetStock).Error; err != nil {
			return nil, err
		}
	}
	
	// Log source movement (out)
	if err := r.LogStockMovement(tx, sourceWarehouseID, productID, productSKU, entity.MovementTypeTransferOut, quantity, "transfer", reference, ""); err != nil {
		return nil, err
	}
	
	// Log target movement (in)
	if err := r.LogStockMovement(tx, targetWarehouseID, productID, productSKU, entity.MovementTypeTransferIn, quantity, "transfer", reference, ""); err != nil {
		return nil, err
	}
	
	// Mark transfer as completed
	transfer.Status = entity.StatusCompleted
	if err := tx.Save(transfer).Error; err != nil {
		return nil, err
	}
	
	return transfer, nil
}

// LogStockMovement records a stock movement
func (r *StockRepository) LogStockMovement(tx *gorm.DB, warehouseID, productID uint, productSKU string, movementType entity.MovementType, quantity int, referenceType, referenceID, notes string) error {
	movement := &entity.StockMovement{
		WarehouseID:   warehouseID,
		ProductID:     productID,
		ProductSKU:    productSKU,
		MovementType:  movementType,
		Quantity:      quantity,
		ReferenceType: referenceType,
		ReferenceID:   referenceID,
		Notes:         notes,
	}
	
	return tx.Create(movement).Error
}

// GetStock gets a single stock record with locking if requested
func (r *StockRepository) GetStock(tx *gorm.DB, warehouseID, productID uint, forUpdate bool) (*entity.WarehouseStock, error) {
	stock := new(entity.WarehouseStock)
	
	query := tx
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	
	result := query.Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).First(stock)
	if result.Error != nil {
		return nil, result.Error
	}
	
	stock.CalculateAvailableQuantity()
	return stock, nil
}