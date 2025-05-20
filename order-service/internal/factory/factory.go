package factory

import (
	"context"
	"order-service/internal/config"
	"order-service/internal/gateway/warehouse"
	"order-service/internal/messaging"
	"order-service/internal/repository"
	"order-service/internal/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Factory provides methods to create dependencies
type Factory struct {
	DB       *gorm.DB
	Config   *config.AppConfig
	Log      *logrus.Logger
	Validate *validator.Validate
}

// NewFactory creates a new Factory instance
func NewFactory(db *gorm.DB, config *config.AppConfig, log *logrus.Logger, validate *validator.Validate) *Factory {
	return &Factory{
		DB:       db,
		Config:   config,
		Log:      log,
		Validate: validate,
	}
}

// CreateRabbitMQClient creates a new RabbitMQ client
func (f *Factory) CreateRabbitMQClient() (*messaging.RabbitMQClient, error) {
	rabbitConfig := f.Config.GetRabbitMQConfig()
	return messaging.NewRabbitMQClient(
		rabbitConfig.GetRabbitMQMessagingConfig(),
		f.Log,
	)
}

// CreateInventoryProducer creates a new inventory message producer
func (f *Factory) CreateInventoryProducer() (*messaging.InventoryProducer, error) {
	mqClient, err := f.CreateRabbitMQClient()
	if err != nil {
		return nil, err
	}
	
	return messaging.NewInventoryProducer(mqClient, f.Log), nil
}

// CreateInventoryConsumer creates a new inventory message consumer
func (f *Factory) CreateInventoryConsumer(handler messaging.InventoryResponseHandler) (*messaging.InventoryConsumer, error) {
	mqClient, err := f.CreateRabbitMQClient()
	if err != nil {
		return nil, err
	}
	
	routingKey := "inventory.*.response"
	
	return messaging.NewInventoryConsumer(mqClient, f.Log, handler, routingKey), nil
}

// CreateWarehouseClient creates a new warehouse client
func (f *Factory) CreateWarehouseClient() *warehouse.Client {
	warehouseConfig := f.Config.GetWarehouseConfig()
	return warehouse.NewClient(
		warehouseConfig.BaseURL,
		warehouseConfig.Timeout,
		f.Log,
	)
}

// CreateWarehouseGateway creates a new warehouse gateway
func (f *Factory) CreateWarehouseGateway() warehouse.WarehouseGatewayInterface {
	client := f.CreateWarehouseClient()
	return warehouse.NewWarehouseGateway(client, f.Log)
}

// CreateReservationRepository creates a new reservation repository
func (f *Factory) CreateReservationRepository() repository.ReservationRepositoryInterface {
	return repository.NewReservationRepository(f.Log, f.DB)
}

// CreateOrderRepository creates a new order repository
func (f *Factory) CreateOrderRepository() repository.OrderRepositoryInterface {
	return repository.NewOrderRepository(f.Log, f.DB)
}

// CreateInventoryUseCase creates a new inventory usecase
func (f *Factory) CreateInventoryUseCase() usecase.InventoryUseCaseInterface {
	warehouseConfig := f.Config.GetWarehouseConfig()
	
	// Check the async mode config to determine which implementation to use
	if warehouseConfig.AsyncMode {
		// If async mode is enabled, use the async inventory use case with messaging
		producer, err := f.CreateInventoryProducer()
		if err != nil {
			f.Log.WithError(err).Error("Failed to create inventory producer, falling back to synchronous mode")
			// Fall back to synchronous mode
			return usecase.NewInventoryWarehouseUseCase(
				f.DB,
				f.Log,
				nil,
				f.CreateWarehouseGateway(),
			)
		}
		
		asyncUseCase := usecase.NewInventoryAsyncUseCase(
			f.DB,
			f.Log,
			producer,
			f.CreateOrderRepository(),
			f.CreateReservationRepository(),
		)
		
		// Create and start the consumer with the async use case as the handler
		consumer, err := f.CreateInventoryConsumer(asyncUseCase)
		if err != nil {
			f.Log.WithError(err).Error("Failed to create inventory consumer, falling back to synchronous mode")
			// Fall back to synchronous mode
			return usecase.NewInventoryWarehouseUseCase(
				f.DB,
				f.Log,
				nil,
				f.CreateWarehouseGateway(),
			)
		}
		
		// Start consuming messages
		if err := consumer.Start(context.Background()); err != nil {
			f.Log.WithError(err).Error("Failed to start inventory consumer, falling back to synchronous mode")
			// Fall back to synchronous mode
			return usecase.NewInventoryWarehouseUseCase(
				f.DB,
				f.Log,
				nil,
				f.CreateWarehouseGateway(),
			)
		}
		
		return asyncUseCase
	}
	
	// Use the synchronous mode with direct HTTP calls
	return usecase.NewInventoryWarehouseUseCase(
		f.DB,
		f.Log,
		nil,
		f.CreateWarehouseGateway(),
	)
}

// CreateOrderUseCase creates a new order usecase
func (f *Factory) CreateOrderUseCase() usecase.OrderUseCaseInterface {
	return usecase.NewOrderUseCase(
		f.DB,
		f.Log,
		f.Validate,
		f.CreateOrderRepository(),
		f.CreateReservationRepository(),
		f.CreateInventoryUseCase(),
	)
}