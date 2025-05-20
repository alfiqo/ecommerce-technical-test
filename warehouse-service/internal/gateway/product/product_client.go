package product

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// ProductInfo represents product information from the external product service
type ProductInfo struct {
	ID          uint   `json:"id"`
	SKU         string `json:"sku"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// ProductClientInterface defines the interface for interacting with the product service
type ProductClientInterface interface {
	GetProductByID(ctx context.Context, productID uint) (*ProductInfo, error)
	GetProductBySKU(ctx context.Context, sku string) (*ProductInfo, error)
	ValidateProduct(ctx context.Context, productID uint) (bool, error)
}

// ProductClient implements ProductClientInterface for the external product service
type ProductClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Log        *logrus.Logger
}

// NewProductClient creates a new ProductClient with the given configuration
func NewProductClient(log *logrus.Logger) *ProductClient {
	baseURL := os.Getenv("PRODUCT_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://product-service:8080/api/v1" // Default URL
	}

	return &ProductClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		Log: log,
	}
}

// GetProductByID fetches product information by ID from the product service
func (c *ProductClient) GetProductByID(ctx context.Context, productID uint) (*ProductInfo, error) {
	url := fmt.Sprintf("%s/products/%d", c.BaseURL, productID)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.Log.WithError(err).Error("Failed to create request for product service")
		return nil, err
	}
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Log.WithError(err).Error("Failed to fetch product from product service")
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("product with ID %d not found", productID)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code from product service: %d", resp.StatusCode)
	}
	
	var product ProductInfo
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		c.Log.WithError(err).Error("Failed to decode product response")
		return nil, err
	}
	
	return &product, nil
}

// GetProductBySKU fetches product information by SKU from the product service
func (c *ProductClient) GetProductBySKU(ctx context.Context, sku string) (*ProductInfo, error) {
	url := fmt.Sprintf("%s/products/sku/%s", c.BaseURL, sku)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.Log.WithError(err).Error("Failed to create request for product service")
		return nil, err
	}
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		c.Log.WithError(err).Error("Failed to fetch product from product service")
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("product with SKU %s not found", sku)
	}
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code from product service: %d", resp.StatusCode)
	}
	
	var product ProductInfo
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		c.Log.WithError(err).Error("Failed to decode product response")
		return nil, err
	}
	
	return &product, nil
}

// ValidateProduct checks if a product exists and is valid
func (c *ProductClient) ValidateProduct(ctx context.Context, productID uint) (bool, error) {
	product, err := c.GetProductByID(ctx, productID)
	if err != nil {
		return false, err
	}
	
	return product != nil, nil
}