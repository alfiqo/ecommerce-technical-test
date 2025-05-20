package model

// WebResponse is a struct for web response
type WebResponse[T any] struct {
	Data   T      `json:"data,omitempty"`
	Errors string `json:"errors,omitempty"`
}