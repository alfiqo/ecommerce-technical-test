package usecase

import (
	"context"
	"errors"
	"order-service/internal/entity"
	"order-service/internal/model"
	"order-service/internal/model/converter"
	"order-service/internal/repository"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ReservationUseCaseInterface interface {
	CreateReservation(ctx context.Context, request *model.ReservationRequest) (*model.ReservationResponse, error)
	GetReservationsByOrderID(ctx context.Context, orderID uint) ([]model.ReservationResponse, error)
	DeactivateReservation(ctx context.Context, reservationID uint) error
	CleanupExpiredReservations(ctx context.Context) error
}

type ReservationUseCase struct {
	DB                    *gorm.DB
	Log                   *logrus.Logger
	Validate              *validator.Validate
	ReservationRepository repository.ReservationRepositoryInterface
	OrderRepository       repository.OrderRepositoryInterface
}

func NewReservationUseCase(
	db *gorm.DB,
	logger *logrus.Logger,
	validate *validator.Validate,
	reservationRepository repository.ReservationRepositoryInterface,
	orderRepository repository.OrderRepositoryInterface,
) ReservationUseCaseInterface {
	return &ReservationUseCase{
		DB:                    db,
		Log:                   logger,
		Validate:              validate,
		ReservationRepository: reservationRepository,
		OrderRepository:       orderRepository,
	}
}

func (c *ReservationUseCase) CreateReservation(ctx context.Context, request *model.ReservationRequest) (*model.ReservationResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Validate request
	err := c.Validate.Struct(request)
	if err != nil {
		c.Log.Warnf("Invalid request body: %+v", err)
		return nil, fiber.ErrBadRequest
	}

	// Check if order exists
	order, err := c.OrderRepository.FindOrderByID(tx, request.OrderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.Log.Warnf("Order not found: %d", request.OrderID)
			return nil, fiber.ErrNotFound
		}
		c.Log.Warnf("Failed to find order: %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	// Only allow reservations for pending orders
	if order.Status != entity.OrderStatusPending {
		c.Log.Warnf("Cannot create reservation for non-pending order: %d", request.OrderID)
		return nil, fiber.ErrBadRequest
	}

	// Create reservation
	reservation := &entity.Reservation{
		OrderID:     request.OrderID,
		ProductID:   request.ProductID,
		WarehouseID: request.WarehouseID,
		Quantity:    request.Quantity,
		ExpiresAt:   request.ExpiresAt,
		IsActive:    true,
	}

	if err := c.ReservationRepository.CreateReservation(tx, reservation); err != nil {
		c.Log.Warnf("Failed to create reservation: %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed to commit transaction: %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.ReservationToResponse(reservation), nil
}

func (c *ReservationUseCase) GetReservationsByOrderID(ctx context.Context, orderID uint) ([]model.ReservationResponse, error) {
	reservations, err := c.ReservationRepository.FindReservationsByOrderID(c.DB.WithContext(ctx), orderID)
	if err != nil {
		c.Log.Warnf("Failed to find reservations by order ID: %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	return converter.ReservationsToResponse(reservations), nil
}

func (c *ReservationUseCase) DeactivateReservation(ctx context.Context, reservationID uint) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.ReservationRepository.UpdateReservationStatus(tx, reservationID, false); err != nil {
		c.Log.Warnf("Failed to deactivate reservation: %+v", err)
		return fiber.ErrInternalServerError
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed to commit transaction: %+v", err)
		return fiber.ErrInternalServerError
	}

	return nil
}

func (c *ReservationUseCase) CleanupExpiredReservations(ctx context.Context) error {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	currentTime := time.Now()

	// Find expired reservations that are still active
	expiredReservations, err := c.ReservationRepository.FindExpiredReservations(tx, currentTime)
	if err != nil {
		c.Log.Warnf("Failed to find expired reservations: %+v", err)
		return fiber.ErrInternalServerError
	}

	// Deactivate expired reservations
	for _, reservation := range expiredReservations {
		if err := c.ReservationRepository.UpdateReservationStatus(tx, reservation.ID, false); err != nil {
			c.Log.Warnf("Failed to deactivate reservation: %+v", err)
			return fiber.ErrInternalServerError
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed to commit transaction: %+v", err)
		return fiber.ErrInternalServerError
	}

	return nil
}