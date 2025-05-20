package usecase

import (
	"context"
	"errors"
	"order-service/internal/entity"
	"order-service/internal/model"
	repository_mock "order-service/mocks/repository"
	usecase_mock "order-service/mocks/usecase"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/mock/gomock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestOrderUseCase_CreateOrder(t *testing.T) {
	// Create SQL mock
	sqlDB, sqlMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create SQL mock: %v", err)
	}
	defer sqlDB.Close()

	// Create mock DB with transaction expectations
	sqlMock.ExpectBegin()
	sqlMock.ExpectCommit()

	// Configure GORM to use the mock
	dialector := mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open GORM DB: %v", err)
	}

	// Create mocks
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockOrderRepo := new(repository_mock.OrderRepositoryMock)
	mockReservationRepo := new(repository_mock.ReservationRepositoryMock)
	mockInventoryUseCase := usecase_mock.NewMockInventoryUseCaseInterface(ctrl)
	
	// Setup inventory use case expectations - AnyTimes to prevent conflicts between subtests
	mockInventoryUseCase.EXPECT().
		CheckAndReserveStock(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
		
	// For error recovery in error case
	mockInventoryUseCase.EXPECT().
		ReleaseReservation(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
	
	// Create use case
	logger := logrus.New()
	validate := validator.New()
	orderUseCase := NewOrderUseCase(db, logger, validate, mockOrderRepo, mockReservationRepo, mockInventoryUseCase)
	
	// Test case 1: Successful order creation
	t.Run("SuccessfulOrderCreation", func(t *testing.T) {
		// Create test data
		createRequest := &model.CreateOrderRequest{
			UserID:          "test-user-id",
			ShippingAddress: "123 Test St",
			PaymentMethod:   "credit_card",
			Items: []model.OrderItemRequest{
				{
					ProductID:   1,
					WarehouseID: 1,
					Quantity:    2,
					UnitPrice:   10.0,
				},
			},
		}
		
		// Set up expectations for the mock
		mockOrderRepo.On("CreateOrder", mock.Anything, mock.MatchedBy(func(order *entity.Order) bool {
			return order.UserID == createRequest.UserID && 
				order.ShippingAddress == createRequest.ShippingAddress &&
				order.PaymentMethod == createRequest.PaymentMethod &&
				order.Status == entity.OrderStatusPending
		})).Run(func(args mock.Arguments) {
			// Set the ID when creating the order
			order := args.Get(1).(*entity.Order)
			order.ID = 1
		}).Return(nil).Once()
		
		mockOrderRepo.On("CreateOrderItems", mock.Anything, mock.MatchedBy(func(items []entity.OrderItem) bool {
			return len(items) == 1 &&
				items[0].OrderID == 1 &&
				items[0].ProductID == createRequest.Items[0].ProductID &&
				items[0].Quantity == createRequest.Items[0].Quantity
		})).Return(nil).Once()
		
		mockReservationRepo.On("CreateReservationBatch", mock.Anything, mock.MatchedBy(func(reservations []entity.Reservation) bool {
			return len(reservations) == 1 &&
				reservations[0].OrderID == 1 &&
				reservations[0].ProductID == createRequest.Items[0].ProductID &&
				reservations[0].Quantity == createRequest.Items[0].Quantity
		})).Return(nil).Once()
		
		createdOrder := &entity.Order{
			ID:              1,
			UserID:          createRequest.UserID,
			Status:          entity.OrderStatusPending,
			TotalAmount:     20.0,
			ShippingAddress: createRequest.ShippingAddress,
			PaymentMethod:   createRequest.PaymentMethod,
			PaymentDeadline: time.Now().Add(24 * time.Hour),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			OrderItems: []entity.OrderItem{
				{
					ID:          1,
					OrderID:     1,
					ProductID:   createRequest.Items[0].ProductID,
					WarehouseID: createRequest.Items[0].WarehouseID,
					Quantity:    createRequest.Items[0].Quantity,
					UnitPrice:   createRequest.Items[0].UnitPrice,
					TotalPrice:  20.0,
				},
			},
		}
		
		mockOrderRepo.On("FindOrderByID", mock.Anything, uint(1)).Return(createdOrder, nil).Once()
		
		// Call the method
		ctx := context.Background()
		response, err := orderUseCase.CreateOrder(ctx, createRequest)
		
		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, uint(1), response.ID)
		assert.Equal(t, createRequest.UserID, response.UserID)
		assert.Equal(t, "pending", response.Status)
		assert.Equal(t, 20.0, response.TotalAmount)
		assert.Equal(t, 1, len(response.Items))
		
		// Verify mock expectations
		mockOrderRepo.AssertExpectations(t)
		mockReservationRepo.AssertExpectations(t)
	})
	
	// Test case 2: Empty items list
	t.Run("EmptyItemsList", func(t *testing.T) {
		// Create test data with empty items
		createRequest := &model.CreateOrderRequest{
			UserID:          "test-user-id",
			ShippingAddress: "123 Test St",
			PaymentMethod:   "credit_card",
			Items:           []model.OrderItemRequest{},
		}
		
		// Call the method
		ctx := context.Background()
		response, err := orderUseCase.CreateOrder(ctx, createRequest)
		
		// Assertions
		assert.Error(t, err)
		assert.Equal(t, fiber.ErrBadRequest, err)
		assert.Nil(t, response)
	})
	
	// Test case 3: Error creating order
	t.Run("ErrorCreatingOrder", func(t *testing.T) {
		// Create test data
		createRequest := &model.CreateOrderRequest{
			UserID:          "test-user-id",
			ShippingAddress: "123 Test St",
			PaymentMethod:   "credit_card",
			Items: []model.OrderItemRequest{
				{
					ProductID:   1,
					WarehouseID: 1,
					Quantity:    2,
					UnitPrice:   10.0,
				},
			},
		}
		
		// Set up expectations for the mock
		mockOrderRepo.On("CreateOrder", mock.Anything, mock.MatchedBy(func(order *entity.Order) bool {
			return order.UserID == createRequest.UserID
		})).Return(errors.New("database error")).Once()
		
		// Call the method
		ctx := context.Background()
		response, err := orderUseCase.CreateOrder(ctx, createRequest)
		
		// Assertions
		assert.Error(t, err)
		assert.Equal(t, fiber.ErrInternalServerError, err)
		assert.Nil(t, response)
		
		// Verify mock expectations
		mockOrderRepo.AssertExpectations(t)
	})
}

func TestOrderUseCase_GetOrderByID(t *testing.T) {
	// Create SQL mock
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create SQL mock: %v", err)
	}
	defer sqlDB.Close()

	// Configure GORM to use the mock
	dialector := mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	})
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open GORM DB: %v", err)
	}

	// Create mocks
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockOrderRepo := new(repository_mock.OrderRepositoryMock)
	mockReservationRepo := new(repository_mock.ReservationRepositoryMock)
	mockInventoryUseCase := usecase_mock.NewMockInventoryUseCaseInterface(ctrl)
	
	// Create use case
	logger := logrus.New()
	validate := validator.New()
	orderUseCase := NewOrderUseCase(db, logger, validate, mockOrderRepo, mockReservationRepo, mockInventoryUseCase)
	
	// Test case 1: Order found
	t.Run("OrderFound", func(t *testing.T) {
		// Create test data
		order := &entity.Order{
			ID:              1,
			UserID:          "test-user-id",
			Status:          entity.OrderStatusPending,
			TotalAmount:     20.0,
			ShippingAddress: "123 Test St",
			PaymentMethod:   "credit_card",
			PaymentDeadline: time.Now().Add(24 * time.Hour),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			OrderItems: []entity.OrderItem{
				{
					ID:          1,
					OrderID:     1,
					ProductID:   1,
					WarehouseID: 1,
					Quantity:    2,
					UnitPrice:   10.0,
					TotalPrice:  20.0,
				},
			},
		}
		
		// Set up expectations for the mock
		mockOrderRepo.On("FindOrderByID", mock.Anything, uint(1)).Return(order, nil).Once()
		
		// Call the method
		ctx := context.Background()
		response, err := orderUseCase.GetOrderByID(ctx, 1)
		
		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, uint(1), response.ID)
		assert.Equal(t, "test-user-id", response.UserID)
		assert.Equal(t, "pending", response.Status)
		assert.Equal(t, 20.0, response.TotalAmount)
		assert.Equal(t, 1, len(response.Items))
		
		// Verify mock expectations
		mockOrderRepo.AssertExpectations(t)
	})
	
	// Test case 2: Order not found
	t.Run("OrderNotFound", func(t *testing.T) {
		// Set up expectations for the mock
		mockOrderRepo.On("FindOrderByID", mock.Anything, uint(999)).
			Return(nil, gorm.ErrRecordNotFound).Once()
		
		// Call the method
		ctx := context.Background()
		response, err := orderUseCase.GetOrderByID(ctx, 999)
		
		// Assertions
		assert.Error(t, err)
		assert.Equal(t, fiber.ErrNotFound, err)
		assert.Nil(t, response)
		
		// Verify mock expectations
		mockOrderRepo.AssertExpectations(t)
	})
	
	// Test case 3: Database error
	t.Run("DatabaseError", func(t *testing.T) {
		// Set up expectations for the mock
		mockOrderRepo.On("FindOrderByID", mock.Anything, uint(1)).
			Return(nil, errors.New("database error")).Once()
		
		// Call the method
		ctx := context.Background()
		response, err := orderUseCase.GetOrderByID(ctx, 1)
		
		// Assertions
		assert.Error(t, err)
		assert.Equal(t, fiber.ErrInternalServerError, err)
		assert.Nil(t, response)
		
		// Verify mock expectations
		mockOrderRepo.AssertExpectations(t)
	})
}

func TestOrderUseCase_UpdateOrderStatus(t *testing.T) {
	// Create SQL mock for first test case
	sqlDB1, sqlMock1, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create SQL mock: %v", err)
	}
	defer sqlDB1.Close()

	// Create mock DB with transaction expectations for first test
	sqlMock1.ExpectBegin()
	sqlMock1.ExpectCommit()

	// Configure GORM to use the first mock
	dialector1 := mysql.New(mysql.Config{
		Conn:                      sqlDB1,
		SkipInitializeWithVersion: true,
	})
	db1, err := gorm.Open(dialector1, &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open GORM DB: %v", err)
	}

	// Create separate SQL mock for second test case
	sqlDB2, sqlMock2, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create SQL mock: %v", err)
	}
	defer sqlDB2.Close()

	// Create mock DB with transaction expectations for second test
	sqlMock2.ExpectBegin()
	sqlMock2.ExpectCommit()

	// Configure GORM to use the second mock
	dialector2 := mysql.New(mysql.Config{
		Conn:                      sqlDB2,
		SkipInitializeWithVersion: true,
	})
	db2, err := gorm.Open(dialector2, &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open GORM DB: %v", err)
	}

	// Test case 1: Successfully update to paid
	t.Run("SuccessfulUpdateToPaid", func(t *testing.T) {
		// Create mocks
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		
		mockOrderRepo := new(repository_mock.OrderRepositoryMock)
		mockReservationRepo := new(repository_mock.ReservationRepositoryMock)
		mockInventoryUseCase := usecase_mock.NewMockInventoryUseCaseInterface(ctrl)
		
		// Setup inventory use case expectations
		mockInventoryUseCase.EXPECT().
			ConfirmStockDeduction(gomock.Any(), gomock.Any()).
			Return(nil).
			AnyTimes()
		
		// Create use case with first DB
		logger := logrus.New()
		validate := validator.New()
		orderUseCase := NewOrderUseCase(db1, logger, validate, mockOrderRepo, mockReservationRepo, mockInventoryUseCase)

		order := &entity.Order{
			ID:              1,
			UserID:          "test-user-id",
			Status:          entity.OrderStatusPending,
			TotalAmount:     20.0,
			ShippingAddress: "123 Test St",
			PaymentMethod:   "credit_card",
			OrderItems: []entity.OrderItem{
				{
					ID:          1,
					OrderID:     1,
					ProductID:   1,
					WarehouseID: 1,
					Quantity:    2,
					UnitPrice:   10.0,
					TotalPrice:  20.0,
				},
			},
		}
		
		// Set up expectations for the mock
		mockOrderRepo.On("FindOrderByID", mock.Anything, uint(1)).Return(order, nil).Once()
		mockOrderRepo.On("UpdateOrderStatus", mock.Anything, uint(1), entity.OrderStatusPaid).Return(nil).Once()
		
		// Call the method
		ctx := context.Background()
		err := orderUseCase.UpdateOrderStatus(ctx, 1, "paid")
		
		// Assertions
		assert.NoError(t, err)
		
		// Verify mock expectations
		mockOrderRepo.AssertExpectations(t)
		mockReservationRepo.AssertExpectations(t)
	})
	
	// Test case 2: Successfully update to cancelled (deactivate reservations)
	t.Run("SuccessfulUpdateToCancelled", func(t *testing.T) {
		// Create new mocks
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		
		mockOrderRepo := new(repository_mock.OrderRepositoryMock)
		mockReservationRepo := new(repository_mock.ReservationRepositoryMock)
		mockInventoryUseCase := usecase_mock.NewMockInventoryUseCaseInterface(ctrl)
		
		// Setup inventory use case expectations
		mockInventoryUseCase.EXPECT().
			ReleaseReservation(gomock.Any(), gomock.Any()).
			Return(nil).
			AnyTimes()
		
		// Create use case with second DB
		logger := logrus.New()
		validate := validator.New()
		orderUseCase := NewOrderUseCase(db2, logger, validate, mockOrderRepo, mockReservationRepo, mockInventoryUseCase)

		order := &entity.Order{
			ID:              1,
			UserID:          "test-user-id",
			Status:          entity.OrderStatusPending,
			TotalAmount:     20.0,
			ShippingAddress: "123 Test St",
			PaymentMethod:   "credit_card",
			OrderItems: []entity.OrderItem{
				{
					ID:          1,
					OrderID:     1,
					ProductID:   1,
					WarehouseID: 1,
					Quantity:    2,
					UnitPrice:   10.0,
					TotalPrice:  20.0,
				},
			},
		}
		
		// Set up expectations for the mock
		mockOrderRepo.On("FindOrderByID", mock.Anything, uint(1)).Return(order, nil).Once()
		mockReservationRepo.On("DeactivateReservationsByOrderID", mock.Anything, uint(1)).Return(nil).Once()
		mockOrderRepo.On("UpdateOrderStatus", mock.Anything, uint(1), entity.OrderStatusCancelled).Return(nil).Once()
		
		// Call the method
		ctx := context.Background()
		err := orderUseCase.UpdateOrderStatus(ctx, 1, "cancelled")
		
		// Assertions
		assert.NoError(t, err)
		
		// Verify mock expectations
		mockOrderRepo.AssertExpectations(t)
		mockReservationRepo.AssertExpectations(t)
	})
	
	// Use a DB with no transaction expectations for the remaining tests
	sqlDB3, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create SQL mock: %v", err)
	}
	defer sqlDB3.Close()

	// Configure GORM to use the third mock
	dialector3 := mysql.New(mysql.Config{
		Conn:                      sqlDB3,
		SkipInitializeWithVersion: true,
	})
	db3, err := gorm.Open(dialector3, &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open GORM DB: %v", err)
	}

	// Create mocks for remaining tests
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockOrderRepo := new(repository_mock.OrderRepositoryMock)
	mockReservationRepo := new(repository_mock.ReservationRepositoryMock)
	mockInventoryUseCase := usecase_mock.NewMockInventoryUseCaseInterface(ctrl)
	
	// Create use case with third DB
	logger := logrus.New()
	validate := validator.New()
	orderUseCase := NewOrderUseCase(db3, logger, validate, mockOrderRepo, mockReservationRepo, mockInventoryUseCase)
	
	// Test case 3: Invalid status
	t.Run("InvalidStatus", func(t *testing.T) {
		// Call the method with invalid status
		ctx := context.Background()
		err := orderUseCase.UpdateOrderStatus(ctx, 1, "invalid_status")
		
		// Assertions
		assert.Error(t, err)
		assert.Equal(t, fiber.ErrBadRequest, err)
	})
	
	// Test case 4: Order not found
	t.Run("OrderNotFound", func(t *testing.T) {
		// Set up expectations for the mock
		mockOrderRepo.On("FindOrderByID", mock.Anything, uint(999)).
			Return(nil, gorm.ErrRecordNotFound).Once()
		
		// Call the method
		ctx := context.Background()
		err := orderUseCase.UpdateOrderStatus(ctx, 999, "paid")
		
		// Assertions
		assert.Error(t, err)
		assert.Equal(t, fiber.ErrNotFound, err)
		
		// Verify mock expectations
		mockOrderRepo.AssertExpectations(t)
	})
}