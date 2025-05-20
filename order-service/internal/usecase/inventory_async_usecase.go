package usecase

import (
	"context"
	"errors"
	"fmt"
	"order-service/internal/entity"
	"order-service/internal/messaging"
	"order-service/internal/model"
	"order-service/internal/repository"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// InventoryAsyncUseCase implements inventory business logic with asynchronous messaging
type InventoryAsyncUseCase struct {
	DB                  *gorm.DB
	Log                 *logrus.Logger
	InventoryProducer   *messaging.InventoryProducer
	OrderRepository     repository.OrderRepositoryInterface
	ReservationRepository repository.ReservationRepositoryInterface
	
	// For tracking correlations between requests and responses
	reservations     map[uint]string // Map of order IDs to reservation IDs
	reservationMutex sync.RWMutex
	
	// Channel for response notifications
	responses        map[string]chan bool // Map of correlation IDs to response channels
	responsesMutex   sync.RWMutex
}

// NewInventoryAsyncUseCase creates a new InventoryAsyncUseCase instance
func NewInventoryAsyncUseCase(
	db *gorm.DB,
	log *logrus.Logger,
	producer *messaging.InventoryProducer,
	orderRepository repository.OrderRepositoryInterface,
	reservationRepository repository.ReservationRepositoryInterface,
) *InventoryAsyncUseCase {
	return &InventoryAsyncUseCase{
		DB:                  db,
		Log:                 log,
		InventoryProducer:   producer,
		OrderRepository:     orderRepository,
		ReservationRepository: reservationRepository,
		reservations:        make(map[uint]string),
		responses:           make(map[string]chan bool),
	}
}

// CheckAndReserveStock checks and reserves stock for multiple items asynchronously
func (uc *InventoryAsyncUseCase) CheckAndReserveStock(ctx context.Context, items []model.OrderItemRequest) error {
	// Extract order ID if available
	var orderID uint
	for _, item := range items {
		if item.OrderID > 0 {
			orderID = item.OrderID
			break
		}
	}
	
	// Create a response channel and correlationID
	responseChan := make(chan bool, 1)
	correlationID, err := uc.InventoryProducer.PublishReserveStock(ctx, orderID, items)
	if err != nil {
		return fmt.Errorf("failed to publish reserve stock message: %w", err)
	}
	
	// Store the correlation ID for tracking
	uc.responsesMutex.Lock()
	uc.responses[correlationID] = responseChan
	uc.responsesMutex.Unlock()
	
	// Set a timeout for the response
	const timeout = 10 * time.Second
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	
	// Wait for the response or timeout
	select {
	case success := <-responseChan:
		if !success {
			return entity.ErrInsufficientStock
		}
		return nil
	case <-timer.C:
		// Clean up
		uc.responsesMutex.Lock()
		delete(uc.responses, correlationID)
		uc.responsesMutex.Unlock()
		return errors.New("timeout waiting for stock reservation response")
	case <-ctx.Done():
		// Clean up
		uc.responsesMutex.Lock()
		delete(uc.responses, correlationID)
		uc.responsesMutex.Unlock()
		return ctx.Err()
	}
}

// ConfirmStockDeduction commits reserved stock as sold (after payment)
func (uc *InventoryAsyncUseCase) ConfirmStockDeduction(ctx context.Context, orderItems []entity.OrderItem) error {
	if len(orderItems) == 0 {
		return errors.New("no order items provided")
	}
	
	orderID := orderItems[0].OrderID
	
	// Get the reservation ID
	uc.reservationMutex.RLock()
	reservationID, exists := uc.reservations[orderID]
	uc.reservationMutex.RUnlock()
	
	if !exists {
		// Use a default format if no reservation ID is found
		reservationID = fmt.Sprintf("res_%d", orderID)
	}
	
	// Publish the confirmation message
	err := uc.InventoryProducer.PublishConfirmStock(ctx, orderID, reservationID)
	if err != nil {
		return fmt.Errorf("failed to publish confirm stock message: %w", err)
	}
	
	// For now, we're not waiting for a response to the confirmation
	// since this is typically a "fire and forget" operation
	
	return nil
}

// ReleaseReservation releases stock back to available inventory (e.g., cancelled order)
func (uc *InventoryAsyncUseCase) ReleaseReservation(ctx context.Context, orderItems []entity.OrderItem) error {
	if len(orderItems) == 0 {
		return errors.New("no order items provided")
	}
	
	orderID := orderItems[0].OrderID
	
	// Get the reservation ID
	uc.reservationMutex.RLock()
	reservationID, exists := uc.reservations[orderID]
	uc.reservationMutex.RUnlock()
	
	if !exists {
		// Use a default format if no reservation ID is found
		reservationID = fmt.Sprintf("res_%d", orderID)
	}
	
	// Get order item details
	orderItem := orderItems[0]
	
	// Create a new release stock message with order item details
	message := &messaging.ReleaseStockMessage{
		InventoryMessage: messaging.InventoryMessage{
			Type:          messaging.MessageTypeReleaseStock,
			OrderID:       orderID,
			ReservationID: reservationID,
			Timestamp:     time.Now(),
		},
		WarehouseID: orderItem.WarehouseID,
		ProductID:   orderItem.ProductID,
		Quantity:    orderItem.Quantity,
		Reference:   reservationID,
	}
	
	// Set correlation ID
	correlationID := fmt.Sprintf("rel_%s_%d", uuid.New().String(), orderID)
	message.CorrelationID = correlationID
	
	// Publish the message using the client directly
	err := uc.InventoryProducer.PublishReleaseStock(ctx, orderID, reservationID)
	if err != nil {
		return fmt.Errorf("failed to publish release stock message: %w", err)
	}
	
	// Clean up the reservation
	uc.reservationMutex.Lock()
	delete(uc.reservations, orderID)
	uc.reservationMutex.Unlock()
	
	return nil
}

// GetInventory gets current inventory level for a product (not implemented for async version)
func (uc *InventoryAsyncUseCase) GetInventory(ctx context.Context, productID, warehouseID uint) (*entity.Inventory, error) {
	return nil, errors.New("get inventory not implemented for async inventory use case")
}

// UpdateInventory updates inventory quantity (not implemented for async version)
func (uc *InventoryAsyncUseCase) UpdateInventory(ctx context.Context, inventory *entity.Inventory) error {
	return errors.New("update inventory not implemented for async inventory use case")
}

// HandleReservationResponse handles responses from the warehouse service for stock reservations
func (uc *InventoryAsyncUseCase) HandleReservationResponse(ctx context.Context, message *messaging.ReserveStockResponseMessage) error {
	// Extract the correlation ID
	var correlationID string
	if ctxVal := ctx.Value("correlation_id"); ctxVal != nil {
		correlationID, _ = ctxVal.(string)
	}
	
	// Store the reservation ID for the order if successful
	if message.Success && message.ReservationID != "" {
		uc.reservationMutex.Lock()
		uc.reservations[message.OrderID] = message.ReservationID
		uc.reservationMutex.Unlock()
	}
	
	// Find and notify the waiting response channel
	if correlationID != "" {
		uc.responsesMutex.RLock()
		responseChan, exists := uc.responses[correlationID]
		uc.responsesMutex.RUnlock()
		
		if exists {
			// Send the result
			select {
			case responseChan <- message.Success:
				// Message sent
			default:
				// Channel is full or closed, which shouldn't happen
				uc.Log.Warnf("Failed to send to response channel for correlation ID %s", correlationID)
			}
			
			// Clean up
			uc.responsesMutex.Lock()
			delete(uc.responses, correlationID)
			uc.responsesMutex.Unlock()
		}
	}
	
	return nil
}

// HandleConfirmationResponse handles responses from the warehouse service for stock confirmations
func (uc *InventoryAsyncUseCase) HandleConfirmationResponse(ctx context.Context, message *messaging.InventoryResponseMessage) error {
	// Log the response but no action needed
	uc.Log.WithFields(logrus.Fields{
		"order_id":      message.OrderID,
		"success":       message.Success,
		"error_message": message.Message,
	}).Info("Received stock confirmation response")
	
	return nil
}

// HandleReleaseResponse handles responses from the warehouse service for stock releases
func (uc *InventoryAsyncUseCase) HandleReleaseResponse(ctx context.Context, message *messaging.InventoryResponseMessage) error {
	// Log the response but no action needed
	uc.Log.WithFields(logrus.Fields{
		"order_id":      message.OrderID,
		"success":       message.Success,
		"error_message": message.Message,
	}).Info("Received stock release response")
	
	return nil
}