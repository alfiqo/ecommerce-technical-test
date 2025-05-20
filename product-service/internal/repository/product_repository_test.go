package repository

import (
	"fmt"
	"product-service/internal/entity"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ProductRepositoryTestSuite struct {
	suite.Suite
	DB         *gorm.DB
	repository ProductRepositoryInterface
	mockProduct *entity.Product
}

func (suite *ProductRepositoryTestSuite) SetupTest() {
	// Setup in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	assert.NoError(suite.T(), err)
	
	// Migrate the schema
	err = db.AutoMigrate(&entity.Product{})
	assert.NoError(suite.T(), err)
	
	// Setup repository
	logger := logrus.New()
	suite.DB = db
	suite.repository = NewProductRepository(logger, db)
	
	// Create a mock product - create a unique SKU and barcode per test
	randomUUID := uuid.New().String()
	randomSKU := fmt.Sprintf("TEST-SKU-%s", randomUUID)
	randomBarcode := fmt.Sprintf("BAR-%s", randomUUID)
	
	suite.mockProduct = &entity.Product{
		ID:              uuid.New(),
		Name:            "Test Product",
		Description:     "Test Description",
		BasePrice:       99.99,
		Category:        "Test Category",
		SKU:             randomSKU,
		ThumbnailURL:    "http://example.com/image.jpg",
		Status:          "active",
		Barcode:         randomBarcode,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	
	// Insert the mock product
	err = db.Create(suite.mockProduct).Error
	assert.NoError(suite.T(), err)
}

func (suite *ProductRepositoryTestSuite) TestCreate() {
	t := suite.T()
	
	// Create a new product with unique SKU and barcode
	randomUUID := uuid.New().String()
	newSKU := fmt.Sprintf("NEW-SKU-%s", randomUUID)
	newBarcode := fmt.Sprintf("NEW-BAR-%s", randomUUID)
	
	newProduct := &entity.Product{
		Name:            "New Product",
		Description:     "New Description",
		BasePrice:       149.99,
		Category:        "New Category",
		SKU:             newSKU,
		ThumbnailURL:    "http://example.com/new-image.jpg",
		Status:          "active",
		Barcode:         newBarcode,
	}
	
	// Save to db
	err := suite.repository.Create(suite.DB, newProduct)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, newProduct.ID)
	
	// Verify the product was saved
	var savedProduct entity.Product
	err = suite.DB.Where("sku = ?", newSKU).First(&savedProduct).Error
	assert.NoError(t, err)
	assert.Equal(t, "New Product", savedProduct.Name)
	assert.Equal(t, float64(149.99), savedProduct.BasePrice)
}

func (suite *ProductRepositoryTestSuite) TestFindAll() {
	t := suite.T()
	
	// Add another product with unique SKU
	anotherSKU := fmt.Sprintf("ANOTHER-SKU-%s", uuid.New().String())
	newProduct := &entity.Product{
		Name:            "Another Product",
		Description:     "Another Description",
		BasePrice:       199.99,
		Category:        "Another Category",
		SKU:             anotherSKU,
		ThumbnailURL:    "http://example.com/another-image.jpg",
		Status:          "active",
		Barcode:         "456789123",
	}
	
	err := suite.repository.Create(suite.DB, newProduct)
	assert.NoError(t, err)
	
	// Find all products
	products, count, err := suite.repository.FindAll(suite.DB, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(products))
	assert.Equal(t, int64(2), count)
	
	// Test pagination
	limitedProducts, limitedCount, err := suite.repository.FindAll(suite.DB, 1, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(limitedProducts))
	assert.Equal(t, int64(2), limitedCount)
}

func (suite *ProductRepositoryTestSuite) TestFindByID() {
	t := suite.T()
	
	// Find by ID
	product, err := suite.repository.FindByID(suite.DB, suite.mockProduct.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, suite.mockProduct.ID, product.ID)
	assert.Equal(t, suite.mockProduct.Name, product.Name)
	
	// Test not found
	_, err = suite.repository.FindByID(suite.DB, uuid.New().String())
	assert.Error(t, err)
}

func (suite *ProductRepositoryTestSuite) TestUpdate() {
	t := suite.T()
	
	// Update the mock product
	suite.mockProduct.Name = "Updated Product"
	suite.mockProduct.BasePrice = 129.99
	
	err := suite.repository.Update(suite.DB, suite.mockProduct)
	assert.NoError(t, err)
	
	// Verify the update
	var updatedProduct entity.Product
	err = suite.DB.First(&updatedProduct, "uuid = ?", suite.mockProduct.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated Product", updatedProduct.Name)
	assert.Equal(t, float64(129.99), updatedProduct.BasePrice)
}

func (suite *ProductRepositoryTestSuite) TestDelete() {
	t := suite.T()
	
	// Delete the product
	err := suite.repository.Delete(suite.DB, suite.mockProduct.ID.String())
	assert.NoError(t, err)
	
	// Verify it's deleted
	var deletedProduct entity.Product
	err = suite.DB.First(&deletedProduct, "uuid = ?", suite.mockProduct.ID).Error
	assert.Error(t, err)
	assert.Equal(t, "record not found", err.Error())
}

func (suite *ProductRepositoryTestSuite) TestFindBySKU() {
	t := suite.T()
	
	// Find by SKU
	product, err := suite.repository.FindBySKU(suite.DB, suite.mockProduct.SKU)
	assert.NoError(t, err)
	assert.Equal(t, suite.mockProduct.SKU, product.SKU)
	assert.Equal(t, suite.mockProduct.Name, product.Name)
	
	// Test not found
	_, err = suite.repository.FindBySKU(suite.DB, "NONEXISTENT-SKU")
	assert.Error(t, err)
	assert.Equal(t, "record not found", err.Error())
}

func (suite *ProductRepositoryTestSuite) TestSearch() {
	t := suite.T()
	
	// Add products with different attributes for search testing
	products := []*entity.Product{
		{
			Name:         "Apple iPhone",
			Description:  "Smartphone with iOS",
			BasePrice:    999.99,
			Category:     "Electronics",
			SKU:          fmt.Sprintf("IPHONE-%s", uuid.New().String()),
			Brand:        "Apple",
			ThumbnailURL: "http://example.com/iphone.jpg",
			Status:       "active",
			Barcode:      fmt.Sprintf("BAR-IPHONE-%s", uuid.New().String()),
		},
		{
			Name:         "Samsung Galaxy",
			Description:  "Smartphone with Android",
			BasePrice:    899.99,
			Category:     "Electronics",
			SKU:          fmt.Sprintf("GALAXY-%s", uuid.New().String()),
			Brand:        "Samsung",
			ThumbnailURL: "http://example.com/galaxy.jpg",
			Status:       "active",
			Barcode:      fmt.Sprintf("BAR-GALAXY-%s", uuid.New().String()),
		},
		{
			Name:         "Apple MacBook",
			Description:  "Laptop with macOS",
			BasePrice:    1299.99,
			Category:     "Computers",
			SKU:          fmt.Sprintf("MACBOOK-%s", uuid.New().String()),
			Brand:        "Apple",
			ThumbnailURL: "http://example.com/macbook.jpg",
			Status:       "active",
			Barcode:      fmt.Sprintf("BAR-MACBOOK-%s", uuid.New().String()),
		},
	}
	
	// Insert test products
	for _, p := range products {
		err := suite.repository.Create(suite.DB, p)
		assert.NoError(t, err)
	}
	
	// Search by brand
	appleProducts, appleCount, err := suite.repository.Search(suite.DB, "Apple", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(appleProducts))
	assert.Equal(t, int64(2), appleCount)
	
	// Search by category
	electronicsProducts, electronicsCount, err := suite.repository.Search(suite.DB, "Electronics", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(electronicsProducts))
	assert.Equal(t, int64(2), electronicsCount)
	
	// Search with limit
	limitedProducts, limitedCount, err := suite.repository.Search(suite.DB, "Apple", 1, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(limitedProducts))
	assert.Equal(t, int64(2), limitedCount)
	
	// Search with no results
	noProducts, noCount, err := suite.repository.Search(suite.DB, "Nonexistent", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(noProducts))
	assert.Equal(t, int64(0), noCount)
}

func (suite *ProductRepositoryTestSuite) TestFindByCategory() {
	t := suite.T()
	
	// Add products with different categories
	products := []*entity.Product{
		{
			Name:         "Canon EOS R5",
			Description:  "Mirrorless Camera",
			BasePrice:    3899.99,
			Category:     "Cameras",
			SKU:          fmt.Sprintf("CANON-%s", uuid.New().String()),
			Brand:        "Canon",
			ThumbnailURL: "http://example.com/canon.jpg",
			Status:       "active",
			Barcode:      fmt.Sprintf("BAR-CANON-%s", uuid.New().String()),
		},
		{
			Name:         "Nikon Z7",
			Description:  "Mirrorless Camera",
			BasePrice:    2999.99,
			Category:     "Cameras",
			SKU:          fmt.Sprintf("NIKON-%s", uuid.New().String()),
			Brand:        "Nikon",
			ThumbnailURL: "http://example.com/nikon.jpg",
			Status:       "active",
			Barcode:      fmt.Sprintf("BAR-NIKON-%s", uuid.New().String()),
		},
		{
			Name:         "Logitech Mouse",
			Description:  "Wireless Mouse",
			BasePrice:    49.99,
			Category:     "Accessories",
			SKU:          fmt.Sprintf("LOGITECH-%s", uuid.New().String()),
			Brand:        "Logitech",
			ThumbnailURL: "http://example.com/logitech.jpg",
			Status:       "active",
			Barcode:      fmt.Sprintf("BAR-LOGITECH-%s", uuid.New().String()),
		},
	}
	
	// Insert test products
	for _, p := range products {
		err := suite.repository.Create(suite.DB, p)
		assert.NoError(t, err)
	}
	
	// Find by category
	cameraProducts, cameraCount, err := suite.repository.FindByCategory(suite.DB, "Cameras", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(cameraProducts))
	assert.Equal(t, int64(2), cameraCount)
	
	// Test pagination
	limitedProducts, limitedCount, err := suite.repository.FindByCategory(suite.DB, "Cameras", 1, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(limitedProducts))
	assert.Equal(t, int64(2), limitedCount)
	
	// Test category with no products
	noProducts, noCount, err := suite.repository.FindByCategory(suite.DB, "NonexistentCategory", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(noProducts))
	assert.Equal(t, int64(0), noCount)
}

func TestProductRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(ProductRepositoryTestSuite))
}