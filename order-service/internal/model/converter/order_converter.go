package converter

import (
	"order-service/internal/entity"
	"order-service/internal/model"
)

// OrderToResponse converts an order entity to response model
func OrderToResponse(order *entity.Order) *model.OrderResponse {
	response := &model.OrderResponse{
		ID:              order.ID,
		UserID:          order.UserID,
		Status:          string(order.Status),
		TotalAmount:     order.TotalAmount,
		ShippingAddress: order.ShippingAddress,
		PaymentMethod:   order.PaymentMethod,
		PaymentDeadline: order.PaymentDeadline.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt:       order.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       order.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if len(order.OrderItems) > 0 {
		response.Items = make([]model.OrderItemResponse, len(order.OrderItems))
		for i, item := range order.OrderItems {
			response.Items[i] = model.OrderItemResponse{
				ID:          item.ID,
				ProductID:   item.ProductID,
				WarehouseID: item.WarehouseID,
				Quantity:    item.Quantity,
				UnitPrice:   item.UnitPrice,
				TotalPrice:  item.TotalPrice,
			}
		}
	}

	return response
}

// OrdersToResponse converts a slice of order entities to response models
func OrdersToResponse(orders []entity.Order) []model.OrderResponse {
	responses := make([]model.OrderResponse, len(orders))
	for i, order := range orders {
		response := OrderToResponse(&order)
		responses[i] = *response
	}
	return responses
}

// ReservationToResponse converts a reservation entity to response model
func ReservationToResponse(reservation *entity.Reservation) *model.ReservationResponse {
	return &model.ReservationResponse{
		ID:          reservation.ID,
		OrderID:     reservation.OrderID,
		ProductID:   reservation.ProductID,
		WarehouseID: reservation.WarehouseID,
		Quantity:    reservation.Quantity,
		ExpiresAt:   reservation.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		IsActive:    reservation.IsActive,
		CreatedAt:   reservation.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ReservationsToResponse converts a slice of reservation entities to response models
func ReservationsToResponse(reservations []entity.Reservation) []model.ReservationResponse {
	responses := make([]model.ReservationResponse, len(reservations))
	for i, reservation := range reservations {
		response := ReservationToResponse(&reservation)
		responses[i] = *response
	}
	return responses
}