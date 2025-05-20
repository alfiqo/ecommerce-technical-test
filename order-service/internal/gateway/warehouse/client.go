package warehouse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// HTTPClient defines the interface for HTTP operations
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client implements an HTTP client for the warehouse service
type Client struct {
	BaseURL    string
	HTTPClient HTTPClient
	Timeout    time.Duration
	Log        *logrus.Logger
}

// NewClient creates a new warehouse service client
func NewClient(baseURL string, timeout time.Duration, log *logrus.Logger) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		Timeout: timeout,
		Log:     log,
	}
}

// doRequest performs an HTTP request and unmarshals the response
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	// Create request URL
	url := fmt.Sprintf("%s%s", c.BaseURL, path)

	// Create request body if provided
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("error marshaling request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", "warehouse-service-api-key") // Example API key

	// Execute request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(respBody))
	}

	// Unmarshal response if result container provided
	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("error unmarshaling response body: %w", err)
		}
	}

	return nil
}
