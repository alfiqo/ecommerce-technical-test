package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Product is a struct that represents a product entity
type Product struct {
	ID              uuid.UUID `gorm:"column:uuid;primaryKey"`
	Name            string    `gorm:"column:name;type:varchar(255);not null"`
	Description     string    `gorm:"column:description;type:text"`
	BasePrice       float64   `gorm:"column:base_price;type:decimal(15,2);not null"`
	SKU             string    `gorm:"column:sku;type:varchar(50);uniqueIndex"`
	Barcode         string    `gorm:"column:barcode;type:varchar(50);uniqueIndex"`
	Weight          float64   `gorm:"column:weight;type:decimal(10,3)"`
	Dimensions      string    `gorm:"column:dimensions;type:varchar(100)"`
	Brand           string    `gorm:"column:brand;type:varchar(100)"`
	Manufacturer    string    `gorm:"column:manufacturer;type:varchar(100)"`
	Category        string    `gorm:"column:category;type:varchar(100)"`
	Tags            string    `gorm:"column:tags;type:varchar(255)"`
	Status          string    `gorm:"column:status;type:varchar(50);not null;default:active"`
	ImageURLs       string    `gorm:"column:image_urls;type:text"` // Comma-separated list of image URLs
	ThumbnailURL    string    `gorm:"column:thumbnail_url;type:varchar(255)"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoCreateTime;autoUpdateTime"`
	MetaTitle       string    `gorm:"column:meta_title;type:varchar(255)"`
	MetaDescription string    `gorm:"column:meta_description;type:text"`
	MetaKeywords    string    `gorm:"column:meta_keywords;type:varchar(255)"`
}

// ProductVariant represents a specific variant of a product (e.g., size, color)
type ProductVariant struct {
	ID           uuid.UUID `gorm:"column:uuid;primaryKey"`
	ProductID    uuid.UUID `gorm:"column:product_uuid;type:char(36);not null"`
	SKU          string    `gorm:"column:sku;type:varchar(50);uniqueIndex"`
	Name         string    `gorm:"column:name;type:varchar(255);not null"`
	Attributes   string    `gorm:"column:attributes;type:text"` // JSON string of attributes
	PriceDiff    float64   `gorm:"column:price_diff;type:decimal(15,2);default:0"`
	ThumbnailURL string    `gorm:"column:thumbnail_url;type:varchar(255)"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoCreateTime;autoUpdateTime"`
	Product      Product   `gorm:"foreignKey:ProductID;references:ID"`
}

// ProductCategory represents product categories
type ProductCategory struct {
	ID          uuid.UUID `gorm:"column:uuid;primaryKey"`
	Name        string    `gorm:"column:name;type:varchar(100);not null;uniqueIndex"`
	Description string    `gorm:"column:description;type:text"`
	ParentID    *uuid.UUID `gorm:"column:parent_uuid;type:char(36)"`
	Level       int       `gorm:"column:level;type:int;not null;default:0"`
	ImageURL    string    `gorm:"column:image_url;type:varchar(255)"`
	Status      string    `gorm:"column:status;type:varchar(50);not null;default:active"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoCreateTime;autoUpdateTime"`
	Parent      *ProductCategory `gorm:"foreignKey:ParentID;references:ID"`
}

// ProductCategoryMapping maps products to categories (many-to-many)
type ProductCategoryMapping struct {
	ID         uuid.UUID      `gorm:"column:uuid;primaryKey"`
	ProductID  uuid.UUID      `gorm:"column:product_uuid;type:char(36);not null"`
	CategoryID uuid.UUID      `gorm:"column:category_uuid;type:char(36);not null"`
	CreatedAt  time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;autoCreateTime;autoUpdateTime"`
	Product    Product        `gorm:"foreignKey:ProductID;references:ID"`
	Category   ProductCategory `gorm:"foreignKey:CategoryID;references:ID"`
}

func (p *Product) TableName() string {
	return "products"
}

func (p *Product) BeforeCreate(tx *gorm.DB) (err error) {
	p.ID = uuid.New()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	return
}

func (pv *ProductVariant) TableName() string {
	return "product_variants"
}

func (pv *ProductVariant) BeforeCreate(tx *gorm.DB) (err error) {
	pv.ID = uuid.New()
	pv.CreatedAt = time.Now()
	pv.UpdatedAt = time.Now()
	return
}

func (pc *ProductCategory) TableName() string {
	return "product_categories"
}

func (pc *ProductCategory) BeforeCreate(tx *gorm.DB) (err error) {
	pc.ID = uuid.New()
	pc.CreatedAt = time.Now()
	pc.UpdatedAt = time.Now()
	return
}

func (pcm *ProductCategoryMapping) TableName() string {
	return "product_category_mappings"
}

func (pcm *ProductCategoryMapping) BeforeCreate(tx *gorm.DB) (err error) {
	pcm.ID = uuid.New()
	pcm.CreatedAt = time.Now()
	pcm.UpdatedAt = time.Now()
	return
}