package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// RabbitMQConfig contains configuration for RabbitMQ connection
type RabbitMQConfig struct {
	URL       string
	Exchange  string
	QueueName string
	Username  string
	Password  string
}

// RabbitMQClient provides methods to interact with RabbitMQ
type RabbitMQClient struct {
	config     RabbitMQConfig
	connection *amqp.Connection
	channel    *amqp.Channel
	log        *logrus.Logger
}

// NewRabbitMQClient creates a new RabbitMQ client
func NewRabbitMQClient(config RabbitMQConfig, log *logrus.Logger) (*RabbitMQClient, error) {
	client := &RabbitMQClient{
		config: config,
		log:    log,
	}

	if err := client.Connect(); err != nil {
		return nil, err
	}

	return client, nil
}

// Connect establishes a connection to RabbitMQ
func (c *RabbitMQClient) Connect() error {
	var err error

	// Create connection URL with credentials if provided
	url := c.config.URL
	if c.config.Username != "" && c.config.Password != "" {
		// Parse the URL to inject credentials - handle both formats with/without amqp://
		if url[:4] != "amqp" {
			url = fmt.Sprintf("amqp://%s:%s@%s", c.config.Username, c.config.Password, c.config.URL)
		} else {
			// URL already has protocol
			url = fmt.Sprintf("amqp://%s:%s@%s", c.config.Username, c.config.Password, url[7:])
		}
	} else if url[:4] != "amqp" {
		// Add protocol if it's missing
		url = "amqp://" + url
	}

	c.log.Infof("Connecting to RabbitMQ at %s", url)

	// Connect to RabbitMQ
	c.connection, err = amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create channel
	c.channel, err = c.connection.Channel()
	if err != nil {
		c.connection.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange if provided
	if c.config.Exchange != "" {
		err = c.channel.ExchangeDeclare(
			c.config.Exchange, // exchange name
			"topic",           // type
			true,              // durable
			false,             // auto-deleted
			false,             // internal
			false,             // no-wait
			nil,               // arguments
		)
		if err != nil {
			c.Close()
			return fmt.Errorf("failed to declare exchange: %w", err)
		}
	}

	// Declare queue if provided
	if c.config.QueueName != "" {
		_, err = c.channel.QueueDeclare(
			c.config.QueueName, // queue name
			true,               // durable
			false,              // delete when unused
			false,              // exclusive
			false,              // no-wait
			nil,                // arguments
		)
		if err != nil {
			c.Close()
			return fmt.Errorf("failed to declare queue: %w", err)
		}
	}

	return nil
}

// Close closes the connection to RabbitMQ
func (c *RabbitMQClient) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.connection != nil {
		c.connection.Close()
	}
}

// Publish publishes a message to RabbitMQ
func (c *RabbitMQClient) Publish(ctx context.Context, routingKey string, body interface{}) error {
	// Convert body to JSON
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal message body: %w", err)
	}

	// Create message
	msg := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
		Body:         jsonBody,
	}

	// Get correlation ID and other metadata from context if available
	if ctxVal := ctx.Value("correlation_id"); ctxVal != nil {
		if correlationID, ok := ctxVal.(string); ok {
			msg.CorrelationId = correlationID
		}
	}

	// Publish message
	err = c.channel.Publish(
		c.config.Exchange, // exchange
		routingKey,        // routing key
		false,             // mandatory
		false,             // immediate
		msg,               // message
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Consume starts consuming messages from RabbitMQ
func (c *RabbitMQClient) Consume(routingKey string) (<-chan amqp.Delivery, error) {
	// First bind queue to exchange with routing key if both are provided
	if c.config.Exchange != "" && c.config.QueueName != "" {
		err := c.channel.QueueBind(
			c.config.QueueName, // queue name
			routingKey,         // routing key
			c.config.Exchange,  // exchange
			false,              // no-wait
			nil,                // arguments
		)
		if err != nil {
			return nil, fmt.Errorf("failed to bind queue: %w", err)
		}
	}

	// Start consuming
	return c.channel.Consume(
		c.config.QueueName, // queue
		"",                 // consumer
		false,              // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
}