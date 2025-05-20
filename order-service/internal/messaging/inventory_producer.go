package messaging

import (
	"context"
	"fmt"
	"order-service/internal/model"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// InventoryProducer sends inventory-related messages to the message queue
type InventoryProducer struct {
	mqClient *RabbitMQClient
	log      *logrus.Logger
}

// NewInventoryProducer creates a new InventoryProducer
func NewInventoryProducer(mqClient *RabbitMQClient, log *logrus.Logger) *InventoryProducer {
	return &InventoryProducer{
		mqClient: mqClient,
		log:      log,
	}
}

// PublishReserveStock publishes a stock reservation request
func (p *InventoryProducer) PublishReserveStock(ctx context.Context, orderID uint, items []model.OrderItemRequest) (string, error) {
	// Set reservation expiry to 24 hours from now
	reserveUntil := time.Now().Add(24 * time.Hour)
	
	// Create correlation ID for tracing
	correlationID := uuid.New().String()
	
	// Create context with correlation ID
	ctxWithCorrelation := context.WithValue(ctx, "correlation_id", correlationID)
	
	// Create message
	message := NewReserveStockMessage(orderID, items, reserveUntil)
	message.CorrelationID = correlationID
	
	// Publish message
	err := p.mqClient.Publish(ctxWithCorrelation, MessageTypeReserveStock, message)
	if err != nil {
		return "", fmt.Errorf("failed to publish reserve stock message: %w", err)
	}
	
	p.log.WithFields(logrus.Fields{
		"order_id":       orderID,
		"correlation_id": correlationID,
		"items_count":    len(items),
	}).Info("Published reserve stock message")
	
	return correlationID, nil
}

// PublishConfirmStock publishes a stock confirmation message
func (p *InventoryProducer) PublishConfirmStock(ctx context.Context, orderID uint, reservationID string) error {
	// Create correlation ID for tracing
	correlationID := uuid.New().String()
	
	// Create context with correlation ID
	ctxWithCorrelation := context.WithValue(ctx, "correlation_id", correlationID)
	
	// Create message
	message := NewConfirmStockMessage(orderID, reservationID)
	message.CorrelationID = correlationID
	
	// Publish message
	err := p.mqClient.Publish(ctxWithCorrelation, MessageTypeConfirmStock, message)
	if err != nil {
		return fmt.Errorf("failed to publish confirm stock message: %w", err)
	}
	
	p.log.WithFields(logrus.Fields{
		"order_id":       orderID,
		"reservation_id": reservationID,
		"correlation_id": correlationID,
	}).Info("Published confirm stock message")
	
	return nil
}

// PublishReleaseStock publishes a stock release message
func (p *InventoryProducer) PublishReleaseStock(ctx context.Context, orderID uint, reservationID string) error {
	// Create correlation ID for tracing
	correlationID := uuid.New().String()
	
	// Create context with correlation ID
	ctxWithCorrelation := context.WithValue(ctx, "correlation_id", correlationID)
	
	// Create message
	message := NewReleaseStockMessage(orderID, reservationID)
	message.CorrelationID = correlationID
	
	// Note: This message now contains the following fields based on the updated structure:
	// - WarehouseID (default: 1)
	// - ProductID (default: 1)
	// - Quantity (default: 1)
	// - Reference (set to reservationID)
	// These should be set by the caller when appropriate
	
	// Publish message
	err := p.mqClient.Publish(ctxWithCorrelation, MessageTypeReleaseStock, message)
	if err != nil {
		return fmt.Errorf("failed to publish release stock message: %w", err)
	}
	
	p.log.WithFields(logrus.Fields{
		"order_id":       orderID,
		"reservation_id": reservationID,
		"warehouse_id":   message.WarehouseID,
		"product_id":     message.ProductID,
		"quantity":       message.Quantity,
		"reference":      message.Reference,
		"correlation_id": correlationID,
	}).Info("Published release stock message")
	
	return nil
}