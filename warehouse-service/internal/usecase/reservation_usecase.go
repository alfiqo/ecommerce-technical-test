package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"
	appErrors "warehouse-service/internal/errors"
	"warehouse-service/internal/model"
	"warehouse-service/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ReservationUseCaseInterface interface {
	// ReserveStock reserves stock for a product in a warehouse
	ReserveStock(ctx context.Context, request *model.ReserveStockRequest) (*model.ReservationResponse, error)
	
	// CancelReservation cancels a previous reservation
	CancelReservation(ctx context.Context, request *model.CancelReservationRequest) error
	
	// CommitReservation confirms a reservation and removes stock
	CommitReservation(ctx context.Context, request *model.CommitReservationRequest) error
	
	// GetReservationHistory retrieves reservation history for a product
	GetReservationHistory(ctx context.Context, warehouseID, productID uint, page, limit int) (*model.ReservationHistoryResponse, error)
}

type ReservationUseCase struct {
	DB                  *gorm.DB
	Log                 *logrus.Logger
	Validate            *validator.Validate
	ReservationRepo     repository.ReservationRepositoryInterface
	WarehouseRepository repository.WarehouseRepositoryInterface
}

func NewReservationUseCase(
	db *gorm.DB,
	logger *logrus.Logger,
	validate *validator.Validate,
	reservationRepo repository.ReservationRepositoryInterface,
	warehouseRepo repository.WarehouseRepositoryInterface,
) ReservationUseCaseInterface {
	return &ReservationUseCase{
		DB:                  db,
		Log:                 logger,
		Validate:            validate,
		ReservationRepo:     reservationRepo,
		WarehouseRepository: warehouseRepo,
	}
}

// ReserveStock reserves stock for a product in a warehouse with database-level locking
func (u *ReservationUseCase) ReserveStock(ctx context.Context, request *model.ReserveStockRequest) (*model.ReservationResponse, error) {
	// Validate request
	if err := u.Validate.Struct(request); err != nil {
		u.Log.WithError(err).Warn("Invalid request body for stock reservation")
		return nil, fiber.ErrBadRequest
	}

	// Start a transaction with high isolation level for consistency
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Verify warehouse exists
	warehouse, err := u.WarehouseRepository.FindByID(tx, request.WarehouseID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrResourceNotFound
		}
		u.Log.WithError(err).Error("Failed to find warehouse")
		return nil, fiber.ErrInternalServerError
	}

	// Check if warehouse is active
	if !warehouse.IsActive {
		return nil, appErrors.WithMessage(appErrors.ErrBusinessRuleViolation, "Warehouse is not active")
	}

	// Call repository to reserve stock with locking
	stock, err := u.ReservationRepo.ReserveStock(tx, request.WarehouseID, request.ProductID, request.Quantity)
	if err != nil {
		u.Log.WithError(err).Error("Failed to reserve stock")
		
		// Check for specific error conditions
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrResourceNotFound
		}
		
		// Check for insufficient stock
		if err.Error()[0:16] == "insufficient stock" {
			return nil, appErrors.WithMessage(appErrors.ErrBusinessRuleViolation, err.Error())
		}
		
		return nil, fiber.ErrInternalServerError
	}

	// Create reservation reference
	reference := fmt.Sprintf("RSV-%d-%d-%d", request.WarehouseID, request.ProductID, time.Now().Unix())

	// Log the reservation
	err = u.ReservationRepo.CreateReservationLog(tx, request.WarehouseID, request.ProductID, 
		request.Quantity, string(model.ReservationStatusPending), reference)
	if err != nil {
		u.Log.WithError(err).Error("Failed to create reservation log")
		return nil, fiber.ErrInternalServerError
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		u.Log.WithError(err).Error("Failed to commit transaction")
		return nil, fiber.ErrInternalServerError
	}

	// Build response
	response := &model.ReservationResponse{
		WarehouseID:        stock.WarehouseID,
		ProductID:          stock.ProductID,
		ReservedQuantity:   stock.ReservedQuantity,
		AvailableQuantity:  stock.AvailableQuantity,
		TotalQuantity:      stock.Quantity,
		Reference:          reference,
		Status:             model.ReservationStatusPending,
		ReservationTime:    time.Now().Format(time.RFC3339),
	}

	return response, nil
}

// CancelReservation cancels a previous reservation
func (u *ReservationUseCase) CancelReservation(ctx context.Context, request *model.CancelReservationRequest) error {
	// Validate request
	if err := u.Validate.Struct(request); err != nil {
		u.Log.WithError(err).Warn("Invalid request body for cancellation")
		return fiber.ErrBadRequest
	}

	// Start a transaction
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Cancel the reservation
	err := u.ReservationRepo.CancelReservation(tx, request.WarehouseID, request.ProductID, request.Quantity)
	if err != nil {
		u.Log.WithError(err).Error("Failed to cancel reservation")
		
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrResourceNotFound
		}
		
		// Check for specific error message
		if err.Error()[0:24] == "cannot cancel more than" {
			return appErrors.WithMessage(appErrors.ErrBusinessRuleViolation, err.Error())
		}
		
		return fiber.ErrInternalServerError
	}

	// Log the cancellation
	err = u.ReservationRepo.CreateReservationLog(tx, request.WarehouseID, request.ProductID, 
		request.Quantity, string(model.ReservationStatusCancelled), request.Reference)
	if err != nil {
		u.Log.WithError(err).Error("Failed to create cancellation log")
		return fiber.ErrInternalServerError
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		u.Log.WithError(err).Error("Failed to commit transaction")
		return fiber.ErrInternalServerError
	}

	return nil
}

// CommitReservation confirms a reservation and removes stock
func (u *ReservationUseCase) CommitReservation(ctx context.Context, request *model.CommitReservationRequest) error {
	// Validate request
	if err := u.Validate.Struct(request); err != nil {
		u.Log.WithError(err).Warn("Invalid request body for commit")
		return fiber.ErrBadRequest
	}

	// Start a transaction
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Commit the reservation
	err := u.ReservationRepo.CommitReservation(tx, request.WarehouseID, request.ProductID, request.Quantity)
	if err != nil {
		u.Log.WithError(err).Error("Failed to commit reservation")
		
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrResourceNotFound
		}
		
		// Check for specific error message
		if err.Error()[0:24] == "cannot commit more than" {
			return appErrors.WithMessage(appErrors.ErrBusinessRuleViolation, err.Error())
		}
		
		return fiber.ErrInternalServerError
	}

	// Log the commit
	err = u.ReservationRepo.CreateReservationLog(tx, request.WarehouseID, request.ProductID, 
		request.Quantity, string(model.ReservationStatusCommitted), request.Reference)
	if err != nil {
		u.Log.WithError(err).Error("Failed to create commit log")
		return fiber.ErrInternalServerError
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		u.Log.WithError(err).Error("Failed to commit transaction")
		return fiber.ErrInternalServerError
	}

	return nil
}

// GetReservationHistory retrieves reservation history for a product
func (u *ReservationUseCase) GetReservationHistory(ctx context.Context, warehouseID, productID uint, page, limit int) (*model.ReservationHistoryResponse, error) {
	// Calculate offset
	offset := (page - 1) * limit

	// Start a transaction (read-only)
	tx := u.DB.WithContext(ctx)

	// Get reservation logs
	logs, count, err := u.ReservationRepo.GetReservationLogs(tx, warehouseID, productID, limit, offset)
	if err != nil {
		u.Log.WithError(err).Error("Failed to get reservation logs")
		return nil, fiber.ErrInternalServerError
	}

	// Build response
	response := &model.ReservationHistoryResponse{
		WarehouseID: warehouseID,
		ProductID:   productID,
		Total:       count,
		Page:        page,
		Limit:       limit,
		Logs:        make([]model.ReservationLogResponse, 0, len(logs)),
	}

	// Map logs to response
	for _, log := range logs {
		response.Logs = append(response.Logs, model.ReservationLogResponse{
			Quantity:    log.Quantity,
			Status:      model.ReservationStatus(log.Status),
			Reference:   log.Reference,
			CreatedAt:   log.CreatedAt.Format(time.RFC3339),
		})
	}

	return response, nil
}