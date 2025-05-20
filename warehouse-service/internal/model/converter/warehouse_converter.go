package converter

import (
	"warehouse-service/internal/entity"
	"warehouse-service/internal/model"
)

func WarehouseToResponse(warehouse *entity.Warehouse, stats *model.WarehouseStatsDTO) *model.WarehouseResponse {
	return &model.WarehouseResponse{
		ID:        warehouse.ID,
		Name:      warehouse.Name,
		Location:  warehouse.Location,
		Address:   warehouse.Address,
		IsActive:  warehouse.IsActive,
		Stats:     stats,
		CreatedAt: warehouse.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: warehouse.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func WarehouseRequestToEntity(request *model.CreateWarehouseRequest) *entity.Warehouse {
	return &entity.Warehouse{
		Name:     request.Name,
		Location: request.Location,
		Address:  request.Address,
		IsActive: request.IsActive,
	}
}

func UpdateWarehouseFromRequest(warehouse *entity.Warehouse, request *model.UpdateWarehouseRequest) {
	warehouse.Name = request.Name
	warehouse.Location = request.Location
	warehouse.Address = request.Address
	warehouse.IsActive = request.IsActive
}