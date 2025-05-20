package model

// Swagger doesn't support generics, so we need to define concrete types for documentation

// ProductResponseWrapper is a wrapper for WebResponse[ProductResponse]
type ProductResponseWrapper struct {
	Data   ProductResponse `json:"data,omitempty"`
	Errors string         `json:"errors,omitempty"`
}

// ProductListResponseWrapper is a wrapper for WebResponse[ProductListResponse]
type ProductListResponseWrapper struct {
	Data   ProductListResponse `json:"data,omitempty"`
	Errors string             `json:"errors,omitempty"`
}

// ErrorResponse is a wrapper for WebResponse[string]
type ErrorResponse struct {
	Errors string `json:"errors,omitempty"`
}