package usecase

import (
	"context"
	appContext "product-service/internal/context"
	"product-service/internal/entity"
	appErrors "product-service/internal/errors"
	"product-service/internal/model"
	"product-service/internal/model/converter"
	"product-service/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ProductUseCaseInterface interface {
	GetProducts(ctx context.Context, limit, offset int) (*model.ProductListResponse, error)
	GetProductByID(ctx context.Context, id string) (*model.ProductResponse, error)
	CreateProduct(ctx context.Context, request *model.CreateProductRequest) (*model.ProductResponse, error)
	UpdateProduct(ctx context.Context, id string, request *model.UpdateProductRequest) (*model.ProductResponse, error)
	DeleteProduct(ctx context.Context, id string) error
	SearchProducts(ctx context.Context, query string, limit, offset int) (*model.ProductListResponse, error)
	GetProductsByCategory(ctx context.Context, category string, limit, offset int) (*model.ProductListResponse, error)
}

type ProductUseCase struct {
	DB                *gorm.DB
	Log               *logrus.Logger
	Validate          *validator.Validate
	ProductRepository repository.ProductRepositoryInterface
}

func NewProductUseCase(
	db *gorm.DB,
	logger *logrus.Logger,
	validate *validator.Validate,
	productRepository repository.ProductRepositoryInterface,
) ProductUseCaseInterface {
	return &ProductUseCase{
		DB:                db,
		Log:               logger,
		Validate:          validate,
		ProductRepository: productRepository,
	}
}

func (c *ProductUseCase) GetProducts(ctx context.Context, limit, offset int) (*model.ProductListResponse, error) {
	requestID := appContext.GetRequestID(ctx)
	tx := c.DB.WithContext(ctx)
	
	// Default values for pagination
	if limit <= 0 {
		limit = 10 // Default limit
	}
	
	if offset < 0 {
		offset = 0
	}
	
	// Get products with pagination and count
	products, count, err := c.ProductRepository.FindAll(tx, limit, offset)
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"limit":      limit,
			"offset":     offset,
			"error":      err.Error(),
		}).Warn("Failed to get products")
		
		return nil, appErrors.WithError(appErrors.ErrInternalServer, err)
	}
	
	// Convert to response
	return converter.ProductsToResponse(products, count, limit, offset), nil
}

func (c *ProductUseCase) GetProductByID(ctx context.Context, id string) (*model.ProductResponse, error) {
	requestID := appContext.GetRequestID(ctx)
	tx := c.DB.WithContext(ctx)
	
	// Validate ID
	if id == "" {
		return nil, appErrors.ErrInvalidProductID
	}
	
	// Validate UUID format
	_, err := uuid.Parse(id)
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"product_id": id,
			"error":      err.Error(),
		}).Warn("Invalid product ID format")
		
		return nil, appErrors.WithError(appErrors.ErrInvalidProductID, err)
	}
	
	// Get product by ID
	product, err := c.ProductRepository.FindByID(tx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Log.WithFields(logrus.Fields{
				"request_id": requestID,
				"product_id": id,
			}).Info("Product not found")
			
			return nil, appErrors.ErrProductNotFound
		}
		
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"product_id": id,
			"error":      err.Error(),
		}).Warn("Failed to get product by ID")
		
		return nil, appErrors.WithError(appErrors.ErrInternalServer, err)
	}
	
	// Convert to response
	return converter.ProductToResponse(product), nil
}

func (c *ProductUseCase) CreateProduct(ctx context.Context, request *model.CreateProductRequest) (*model.ProductResponse, error) {
	requestID := appContext.GetRequestID(ctx)
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Validate request
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Invalid request body")
		return nil, appErrors.WithError(appErrors.ErrInvalidInput, err)
	}

	// Check if product with the same SKU already exists
	if request.SKU != "" {
		existingProduct, err := c.ProductRepository.FindBySKU(tx, request.SKU)
		if err == nil && existingProduct != nil {
			c.Log.WithFields(logrus.Fields{
				"request_id": requestID,
				"sku":        request.SKU,
			}).Warn("Product with this SKU already exists")
			return nil, appErrors.ErrDuplicateSKU
		}
	}

	// Create new product entity
	product := &entity.Product{
		Name:         request.Name,
		Description:  request.Description,
		BasePrice:    request.Price,
		Category:     request.Category,
		SKU:          request.SKU,
		ThumbnailURL: request.ImageURL,
		Status:       "active", // Default status for new products
	}

	// Save to database
	if err := c.ProductRepository.Create(tx, product); err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"name":       request.Name,
			"sku":        request.SKU,
			"error":      err.Error(),
		}).Warn("Failed to create product")
		return nil, appErrors.WithError(appErrors.ErrInternalServer, err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to commit transaction")
		return nil, appErrors.WithError(appErrors.ErrInternalServer, err)
	}

	// Convert to response and return
	return converter.ProductToResponse(product), nil
}

func (c *ProductUseCase) UpdateProduct(ctx context.Context, id string, request *model.UpdateProductRequest) (*model.ProductResponse, error) {
	requestID := appContext.GetRequestID(ctx)
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Validate request
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"product_id": id,
			"error":      err.Error(),
		}).Warn("Invalid request body")
		return nil, appErrors.WithError(appErrors.ErrInvalidInput, err)
	}

	// Validate ID format
	if id == "" {
		return nil, appErrors.ErrInvalidProductID
	}

	_, err := uuid.Parse(id)
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"product_id": id,
			"error":      err.Error(),
		}).Warn("Invalid product ID format")
		return nil, appErrors.WithError(appErrors.ErrInvalidProductID, err)
	}

	// Find existing product
	product, err := c.ProductRepository.FindByID(tx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Log.WithFields(logrus.Fields{
				"request_id": requestID,
				"product_id": id,
			}).Info("Product not found")
			return nil, appErrors.ErrProductNotFound
		}
		
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"product_id": id,
			"error":      err.Error(),
		}).Warn("Failed to get product by ID")
		
		return nil, appErrors.WithError(appErrors.ErrInternalServer, err)
	}

	// Check if SKU is being updated and already exists
	if request.SKU != "" && request.SKU != product.SKU {
		existingProduct, err := c.ProductRepository.FindBySKU(tx, request.SKU)
		if err == nil && existingProduct != nil {
			c.Log.WithFields(logrus.Fields{
				"request_id": requestID,
				"product_id": id,
				"sku":        request.SKU,
			}).Warn("Product with this SKU already exists")
			return nil, appErrors.ErrDuplicateSKU
		}
	}

	// Update fields only if they are provided
	if request.Name != "" {
		product.Name = request.Name
	}

	if request.Description != "" {
		product.Description = request.Description
	}

	if request.Price > 0 {
		product.BasePrice = request.Price
	}

	if request.Category != "" {
		product.Category = request.Category
	}

	if request.SKU != "" {
		product.SKU = request.SKU
	}

	if request.ImageURL != "" {
		product.ThumbnailURL = request.ImageURL
	}

	// Save updates
	if err := c.ProductRepository.Update(tx, product); err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"product_id": id,
			"error":      err.Error(),
		}).Warn("Failed to update product")
		return nil, appErrors.WithError(appErrors.ErrInternalServer, err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to commit transaction")
		return nil, appErrors.WithError(appErrors.ErrInternalServer, err)
	}

	// Convert to response and return
	return converter.ProductToResponse(product), nil
}

func (c *ProductUseCase) DeleteProduct(ctx context.Context, id string) error {
	requestID := appContext.GetRequestID(ctx)
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Validate ID format
	if id == "" {
		return appErrors.ErrInvalidProductID
	}

	_, err := uuid.Parse(id)
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"product_id": id,
			"error":      err.Error(),
		}).Warn("Invalid product ID format")
		return appErrors.WithError(appErrors.ErrInvalidProductID, err)
	}

	// Check if product exists
	_, err = c.ProductRepository.FindByID(tx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Log.WithFields(logrus.Fields{
				"request_id": requestID,
				"product_id": id,
			}).Info("Product not found")
			return appErrors.ErrProductNotFound
		}
		
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"product_id": id,
			"error":      err.Error(),
		}).Warn("Failed to get product by ID")
		
		return appErrors.WithError(appErrors.ErrInternalServer, err)
	}

	// Delete the product
	if err := c.ProductRepository.Delete(tx, id); err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"product_id": id,
			"error":      err.Error(),
		}).Warn("Failed to delete product")
		return appErrors.WithError(appErrors.ErrInternalServer, err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to commit transaction")
		return appErrors.WithError(appErrors.ErrInternalServer, err)
	}

	return nil
}

func (c *ProductUseCase) SearchProducts(ctx context.Context, query string, limit, offset int) (*model.ProductListResponse, error) {
	requestID := appContext.GetRequestID(ctx)
	tx := c.DB.WithContext(ctx)

	// Default values for pagination
	if limit <= 0 {
		limit = 10 // Default limit
	}
	
	if offset < 0 {
		offset = 0
	}

	// Search products
	products, count, err := c.ProductRepository.Search(tx, query, limit, offset)
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"query":      query,
			"limit":      limit,
			"offset":     offset,
			"error":      err.Error(),
		}).Warn("Failed to search products")
		
		return nil, appErrors.WithError(appErrors.ErrInternalServer, err)
	}

	// Convert to response
	return converter.ProductsToResponse(products, count, limit, offset), nil
}

func (c *ProductUseCase) GetProductsByCategory(ctx context.Context, category string, limit, offset int) (*model.ProductListResponse, error) {
	requestID := appContext.GetRequestID(ctx)
	tx := c.DB.WithContext(ctx)

	// Validate category
	if category == "" {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
		}).Warn("Empty category provided")
		return nil, appErrors.WithMessage(appErrors.ErrInvalidInput, "Category cannot be empty")
	}

	// Default values for pagination
	if limit <= 0 {
		limit = 10 // Default limit
	}
	
	if offset < 0 {
		offset = 0
	}

	// Get products by category
	products, count, err := c.ProductRepository.FindByCategory(tx, category, limit, offset)
	if err != nil {
		c.Log.WithFields(logrus.Fields{
			"request_id": requestID,
			"category":   category,
			"limit":      limit,
			"offset":     offset,
			"error":      err.Error(),
		}).Warn("Failed to get products by category")
		
		return nil, appErrors.WithError(appErrors.ErrInternalServer, err)
	}

	// Convert to response
	return converter.ProductsToResponse(products, count, limit, offset), nil
}