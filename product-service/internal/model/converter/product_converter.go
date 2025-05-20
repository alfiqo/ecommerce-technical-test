package converter

import (
	"product-service/internal/entity"
	"product-service/internal/model"
	"time"
)

// ProductToResponse converts a product entity to product response
func ProductToResponse(product *entity.Product) *model.ProductResponse {
	return &model.ProductResponse{
		ID:          product.ID.String(),
		Name:        product.Name,
		Description: product.Description,
		Price:       product.BasePrice,
		Stock:       0, // Stock will be managed in inventory service
		Category:    product.Category,
		SKU:         product.SKU,
		ImageURL:    product.ThumbnailURL, // Using ThumbnailURL as main image for simplicity
		CreatedAt:   product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   product.UpdatedAt.Format(time.RFC3339),
	}
}

// ProductsToResponse converts a slice of product entities to product list response
func ProductsToResponse(products []entity.Product, count int64, limit, offset int) *model.ProductListResponse {
	var productResponses []model.ProductResponse
	
	for _, product := range products {
		productResponse := model.ProductResponse{
			ID:          product.ID.String(),
			Name:        product.Name,
			Description: product.Description,
			Price:       product.BasePrice,
			Stock:       0, // Stock will be managed in inventory service
			Category:    product.Category,
			SKU:         product.SKU,
			ImageURL:    product.ThumbnailURL, // Using ThumbnailURL as main image
			CreatedAt:   product.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   product.UpdatedAt.Format(time.RFC3339),
		}
		productResponses = append(productResponses, productResponse)
	}
	
	return &model.ProductListResponse{
		Products: productResponses,
		Count:    count,
		Limit:    limit,
		Offset:   offset,
	}
}