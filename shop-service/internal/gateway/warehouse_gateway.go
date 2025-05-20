package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"shop-service/internal/config/services"
	appErrors "shop-service/internal/errors"
	"shop-service/internal/model"
	"time"

	"github.com/sirupsen/logrus"
)

// WarehouseGatewayInterface defines the interface for warehouse service operations
type WarehouseGatewayInterface interface {
	// GetWarehouseByID retrieves warehouse details by ID
	GetWarehouseByID(ctx context.Context, warehouseID uint) (*model.WarehouseResponse, error)
}

// WarehouseGateway implements WarehouseGatewayInterface
type WarehouseGateway struct {
	Log      *logrus.Logger
	Services *services.ServicesConfig
	Client   *http.Client
}

// NewWarehouseGateway creates a new warehouse gateway instance
func NewWarehouseGateway(log *logrus.Logger, services *services.ServicesConfig) WarehouseGatewayInterface {
	client := &http.Client{
		Timeout: services.Warehouse.Timeout,
	}

	return &WarehouseGateway{
		Log:      log,
		Services: services,
		Client:   client,
	}
}

// GetWarehouseByID retrieves warehouse details by ID
func (g *WarehouseGateway) GetWarehouseByID(ctx context.Context, warehouseID uint) (*model.WarehouseResponse, error) {
	// Create the endpoint URL
	endpoint := fmt.Sprintf("warehouses/%d", warehouseID)
	url := g.Services.Warehouse.GetEndpointURL(endpoint)

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		g.Log.WithFields(logrus.Fields{
			"error":        err.Error(),
			"warehouse_id": warehouseID,
		}).Error("Failed to create request for warehouse service")
		return nil, appErrors.WithError(appErrors.ErrInternalServer, err)
	}

	// Set request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Execute the request
	start := time.Now()
	resp, err := g.Client.Do(req)
	requestDuration := time.Since(start)

	g.Log.WithFields(logrus.Fields{
		"warehouse_id":     warehouseID,
		"request_duration": requestDuration.Milliseconds(),
		"method":           http.MethodGet,
		"url":              url,
	}).Debug("Warehouse service request completed")

	if err != nil {
		g.Log.WithFields(logrus.Fields{
			"error":        err.Error(),
			"warehouse_id": warehouseID,
		}).Error("Failed to send request to warehouse service")
		return nil, appErrors.WithError(appErrors.ErrExternalServiceUnavailable, err)
	}
	defer resp.Body.Close()

	// Handle response status codes
	if resp.StatusCode == http.StatusNotFound {
		return nil, appErrors.ErrWarehouseNotFound
	}

	// Handle other non-successful status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		g.Log.WithFields(logrus.Fields{
			"status_code":  resp.StatusCode,
			"warehouse_id": warehouseID,
		}).Error("Warehouse service returned non-success status code")
		return nil, appErrors.ErrExternalServiceError
	}

	// Parse the response
	var response struct {
		Success bool                     `json:"success"`
		Data    model.WarehouseResponse `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		g.Log.WithFields(logrus.Fields{
			"error":        err.Error(),
			"warehouse_id": warehouseID,
		}).Error("Failed to parse warehouse service response")
		return nil, appErrors.WithError(appErrors.ErrInternalServer, err)
	}

	// Check response success
	if !response.Success {
		g.Log.WithFields(logrus.Fields{
			"warehouse_id": warehouseID,
		}).Error("Warehouse service returned success=false")
		return nil, appErrors.ErrExternalServiceError
	}

	return &response.Data, nil
}