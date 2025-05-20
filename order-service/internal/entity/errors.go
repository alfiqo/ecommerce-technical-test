package entity

import "errors"

var (
	// ErrInsufficientStock is returned when there's not enough stock to fulfill a request
	ErrInsufficientStock = errors.New("insufficient stock available")
)