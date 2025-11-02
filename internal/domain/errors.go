package domain

import "errors"

// Domain errors for allocation operations
var (
	// ErrInsufficientFunds indicates there isn't enough Ready to Assign to cover an allocation
	ErrInsufficientFunds = errors.New("insufficient funds in Ready to Assign")

	// ErrNotPaymentCategory indicates the category is not a payment category
	ErrNotPaymentCategory = errors.New("category is not a payment category")

	// ErrNotUnderfunded indicates the payment category is not underfunded
	ErrNotUnderfunded = errors.New("payment category is not underfunded")

	// ErrCategoryNotFound indicates the category doesn't exist
	ErrCategoryNotFound = errors.New("category not found")
)
