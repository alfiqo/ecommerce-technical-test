package repository

import (
	"product-service/internal/entity"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ProductRepositoryInterface interface {
	Create(db *gorm.DB, product *entity.Product) error
	FindAll(db *gorm.DB, limit, offset int) ([]entity.Product, int64, error)
	FindByID(db *gorm.DB, id string) (*entity.Product, error)
	FindBySKU(db *gorm.DB, sku string) (*entity.Product, error)
	Update(db *gorm.DB, product *entity.Product) error
	Delete(db *gorm.DB, id string) error
	Search(db *gorm.DB, query string, limit, offset int) ([]entity.Product, int64, error)
	FindByCategory(db *gorm.DB, category string, limit, offset int) ([]entity.Product, int64, error)
	GetDB() *gorm.DB
}

type ProductRepository struct {
	DB  *gorm.DB
	Log *logrus.Logger
}

func NewProductRepository(log *logrus.Logger, db *gorm.DB) ProductRepositoryInterface {
	return &ProductRepository{
		DB:  db,
		Log: log,
	}
}

func (r *ProductRepository) GetDB() *gorm.DB {
	return r.DB
}

func (r *ProductRepository) Create(db *gorm.DB, product *entity.Product) error {
	return db.Create(product).Error
}

func (r *ProductRepository) FindAll(db *gorm.DB, limit, offset int) ([]entity.Product, int64, error) {
	var products []entity.Product
	var count int64
	
	// Get total count
	if err := db.Model(&entity.Product{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}
	
	// Apply pagination
	query := db
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	// Get products with ordering
	err := query.Order("created_at DESC").Find(&products).Error
	return products, count, err
}

func (r *ProductRepository) FindByID(db *gorm.DB, id string) (*entity.Product, error) {
	product := new(entity.Product)
	
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	
	if err := db.Where("uuid = ?", parsedID).First(product).Error; err != nil {
		return nil, err
	}
	
	return product, nil
}

func (r *ProductRepository) FindBySKU(db *gorm.DB, sku string) (*entity.Product, error) {
	product := new(entity.Product)
	
	if err := db.Where("sku = ?", sku).First(product).Error; err != nil {
		return nil, err
	}
	
	return product, nil
}

func (r *ProductRepository) Update(db *gorm.DB, product *entity.Product) error {
	return db.Save(product).Error
}

func (r *ProductRepository) Delete(db *gorm.DB, id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	
	return db.Where("uuid = ?", parsedID).Delete(&entity.Product{}).Error
}

func (r *ProductRepository) Search(db *gorm.DB, query string, limit, offset int) ([]entity.Product, int64, error) {
	var products []entity.Product
	var count int64
	searchQuery := "%" + query + "%"

	// Get total count for the search
	countQuery := db.Model(&entity.Product{}).Where(
		"name LIKE ? OR description LIKE ? OR sku LIKE ? OR category LIKE ? OR brand LIKE ?",
		searchQuery, searchQuery, searchQuery, searchQuery, searchQuery,
	)
	if err := countQuery.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Apply search, pagination and get products
	searchDB := db.Where(
		"name LIKE ? OR description LIKE ? OR sku LIKE ? OR category LIKE ? OR brand LIKE ?",
		searchQuery, searchQuery, searchQuery, searchQuery, searchQuery,
	)
	
	if limit > 0 {
		searchDB = searchDB.Limit(limit)
	}
	
	if offset > 0 {
		searchDB = searchDB.Offset(offset)
	}
	
	if err := searchDB.Order("created_at DESC").Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, count, nil
}

func (r *ProductRepository) FindByCategory(db *gorm.DB, category string, limit, offset int) ([]entity.Product, int64, error) {
	var products []entity.Product
	var count int64

	// For testing purposes, log the search parameters
	r.Log.WithFields(logrus.Fields{
		"category": category,
	}).Info("Finding products by category")

	// Use case-insensitive comparison and LIKE for more flexible category matching
	// Get total count for the category
	if err := db.Model(&entity.Product{}).Where("category LIKE ?", category).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Apply category filter, pagination and get products
	query := db.Where("category LIKE ?", category)
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	if err := query.Order("created_at DESC").Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, count, nil
}