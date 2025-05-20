package response

// ErrorResponse represents the structure of error responses returned by the API.
// This type is used solely for Swagger documentation.
type ErrorResponse struct {
	Success bool      `json:"success" example:"false"`
	Error   ErrorInfo `json:"error"`
}

// SuccessResponse represents the structure of success responses returned by the API.
// This type is used solely for Swagger documentation.
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data,omitempty"`
}