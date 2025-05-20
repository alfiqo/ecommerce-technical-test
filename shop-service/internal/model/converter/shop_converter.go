package converter

import (
	"shop-service/internal/entity"
	"shop-service/internal/model"
)

// ToShopResponse converts a Shop entity to a ShopResponse model
func ToShopResponse(shop *entity.Shop) *model.ShopResponse {
	if shop == nil {
		return nil
	}

	return &model.ShopResponse{
		ID:           shop.ID,
		Name:         shop.Name,
		Description:  shop.Description,
		Address:      shop.Address,
		ContactEmail: shop.ContactEmail,
		ContactPhone: shop.ContactPhone,
		IsActive:     shop.IsActive,
		CreatedAt:    shop.CreatedAt,
		UpdatedAt:    shop.UpdatedAt,
	}
}

// ToShopDetailResponse converts a Shop entity to a ShopDetailResponse model
func ToShopDetailResponse(shop *entity.Shop) *model.ShopDetailResponse {
	if shop == nil {
		return nil
	}

	shopResponse := ToShopResponse(shop)
	warehouseIDs := make([]model.WarehouseID, 0, len(shop.Warehouses))

	for _, warehouse := range shop.Warehouses {
		warehouseIDs = append(warehouseIDs, model.WarehouseID{
			ID: warehouse.WarehouseID,
		})
	}

	return &model.ShopDetailResponse{
		ShopResponse: *shopResponse,
		WarehouseIDs: warehouseIDs,
	}
}

// ToShopListResponse converts a slice of Shop entities to a ShopListResponse model
func ToShopListResponse(shops []entity.Shop, totalCount int64, page, pageSize int) *model.ShopListResponse {
	shopResponses := make([]model.ShopResponse, 0, len(shops))

	for _, shop := range shops {
		shopResponses = append(shopResponses, *ToShopResponse(&shop))
	}

	return &model.ShopListResponse{
		Shops:      shopResponses,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}
}