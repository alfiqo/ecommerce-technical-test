package response

// ErrorResponse represents the structure of error responses returned by the API.
// This type is used solely for Swagger documentation.
type ErrorResponse struct {
	Success bool      `json:"success" example:"false"`
	Error   ErrorInfo `json:"error"`
}

// SuccessMessageResponse represents a successful response with a message.
// This type is used solely for Swagger documentation.
type SuccessMessageResponse struct {
	Success bool   `json:"success" example:"true"`
	Data    struct {
		Message string `json:"message" example:"Operation completed successfully"`
	} `json:"data"`
}