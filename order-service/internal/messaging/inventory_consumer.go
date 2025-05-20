package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// InventoryResponseHandler defines the contract for handling inventory response messages
type InventoryResponseHandler interface {
	HandleReservationResponse(ctx context.Context, message *ReserveStockResponseMessage) error
	HandleConfirmationResponse(ctx context.Context, message *InventoryResponseMessage) error
	HandleReleaseResponse(ctx context.Context, message *InventoryResponseMessage) error
}

// InventoryConsumer consumes inventory-related messages from the message queue
type InventoryConsumer struct {
	mqClient  *RabbitMQClient
	log       *logrus.Logger
	handler   InventoryResponseHandler
	routingKey string
	quit      chan struct{}
}

// NewInventoryConsumer creates a new InventoryConsumer
func NewInventoryConsumer(mqClient *RabbitMQClient, log *logrus.Logger, handler InventoryResponseHandler, routingKey string) *InventoryConsumer {
	return &InventoryConsumer{
		mqClient:  mqClient,
		log:       log,
		handler:   handler,
		routingKey: routingKey,
		quit:      make(chan struct{}),
	}
}

// Start starts consuming messages
func (c *InventoryConsumer) Start(ctx context.Context) error {
	msgs, err := c.mqClient.Consume(c.routingKey)
	if err != nil {
		return fmt.Errorf("failed to start consuming messages: %w", err)
	}

	c.log.WithField("routing_key", c.routingKey).Info("Started consuming inventory messages")

	go c.processMessages(ctx, msgs)

	return nil
}

// Stop stops consuming messages
func (c *InventoryConsumer) Stop() {
	close(c.quit)
}

// processMessages processes received messages
func (c *InventoryConsumer) processMessages(ctx context.Context, msgs <-chan amqp.Delivery) {
	for {
		select {
		case <-c.quit:
			c.log.Info("Stopping inventory consumer")
			return
		case msg, ok := <-msgs:
			if !ok {
				c.log.Warn("Message channel closed, attempting to reconnect")
				// Attempt to reconnect
				if err := c.reconnect(ctx); err != nil {
					c.log.WithError(err).Error("Failed to reconnect to message queue")
					time.Sleep(5 * time.Second)
				}
				continue
			}

			// Process the message
			if err := c.processMessage(ctx, msg); err != nil {
				c.log.WithError(err).Error("Failed to process message")
				// Negative acknowledgement
				if err := msg.Nack(false, true); err != nil {
					c.log.WithError(err).Error("Failed to nack message")
				}
			} else {
				// Acknowledge the message
				if err := msg.Ack(false); err != nil {
					c.log.WithError(err).Error("Failed to ack message")
				}
			}
		}
	}
}

// reconnect attempts to reconnect to the message queue
func (c *InventoryConsumer) reconnect(ctx context.Context) error {
	if err := c.mqClient.Connect(); err != nil {
		return err
	}

	msgs, err := c.mqClient.Consume(c.routingKey)
	if err != nil {
		return err
	}

	go c.processMessages(ctx, msgs)
	return nil
}

// processMessage processes a single message
func (c *InventoryConsumer) processMessage(ctx context.Context, msg amqp.Delivery) error {
	// Decode the base message to determine the type
	var baseMessage struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(msg.Body, &baseMessage); err != nil {
		return fmt.Errorf("failed to unmarshal base message: %w", err)
	}

	// Create a context with the correlation ID
	ctxWithCorrelation := ctx
	if msg.CorrelationId != "" {
		ctxWithCorrelation = context.WithValue(ctx, "correlation_id", msg.CorrelationId)
	}

	// Process the message based on its type
	switch baseMessage.Type {
	case "inventory.reserve.response":
		var responseMsg ReserveStockResponseMessage
		if err := json.Unmarshal(msg.Body, &responseMsg); err != nil {
			return fmt.Errorf("failed to unmarshal reserve stock response: %w", err)
		}

		c.log.WithFields(logrus.Fields{
			"order_id":       responseMsg.OrderID,
			"reservation_id": responseMsg.ReservationID,
			"correlation_id": msg.CorrelationId,
			"success":        responseMsg.Success,
		}).Info("Received reserve stock response")

		return c.handler.HandleReservationResponse(ctxWithCorrelation, &responseMsg)

	case "inventory.confirm.response":
		var responseMsg InventoryResponseMessage
		if err := json.Unmarshal(msg.Body, &responseMsg); err != nil {
			return fmt.Errorf("failed to unmarshal confirm stock response: %w", err)
		}

		c.log.WithFields(logrus.Fields{
			"order_id":       responseMsg.OrderID,
			"reservation_id": responseMsg.ReservationID,
			"correlation_id": msg.CorrelationId,
			"success":        responseMsg.Success,
		}).Info("Received confirm stock response")

		return c.handler.HandleConfirmationResponse(ctxWithCorrelation, &responseMsg)

	case "inventory.release.response":
		var responseMsg InventoryResponseMessage
		if err := json.Unmarshal(msg.Body, &responseMsg); err != nil {
			return fmt.Errorf("failed to unmarshal release stock response: %w", err)
		}

		c.log.WithFields(logrus.Fields{
			"order_id":       responseMsg.OrderID,
			"reservation_id": responseMsg.ReservationID,
			"correlation_id": msg.CorrelationId,
			"success":        responseMsg.Success,
		}).Info("Received release stock response")

		return c.handler.HandleReleaseResponse(ctxWithCorrelation, &responseMsg)

	default:
		return errors.New("unknown message type: " + baseMessage.Type)
	}
}