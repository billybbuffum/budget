package validators

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var (
	// periodRegex validates YYYY-MM format with valid months (01-12)
	// Compiled once at package initialization for performance
	periodRegex = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)
)

// ValidateUUID checks if the provided string is a valid UUID format
func ValidateUUID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("invalid UUID format")
	}
	return nil
}

// ValidatePeriodFormat checks if the provided string is in YYYY-MM format
// with valid year (4 digits) and month (01-12)
func ValidatePeriodFormat(period string) error {
	// Regex: 4 digits for year, dash, 01-12 for month
	if !periodRegex.MatchString(period) {
		return fmt.Errorf("invalid period format, expected YYYY-MM")
	}
	return nil
}

// ValidatePeriodRange checks if the period is within reasonable bounds
// (2 years in the past, 5 years in the future)
func ValidatePeriodRange(period string) error {
	// First validate format
	if err := ValidatePeriodFormat(period); err != nil {
		return err
	}

	// Parse the period as a date (first day of month)
	periodTime, err := time.Parse("2006-01", period)
	if err != nil {
		return fmt.Errorf("invalid period format")
	}

	// Calculate acceptable range
	now := time.Now()
	minDate := now.AddDate(-2, 0, 0)  // 2 years ago
	maxDate := now.AddDate(5, 0, 0)   // 5 years in future

	if periodTime.Before(minDate) {
		return fmt.Errorf("period is too far in the past (more than 2 years)")
	}

	if periodTime.After(maxDate) {
		return fmt.Errorf("period is too far in the future (more than 5 years)")
	}

	return nil
}

// ValidateAmountPositive checks if the amount is positive and greater than zero
func ValidateAmountPositive(amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive and greater than zero")
	}
	return nil
}

// ValidateAmountBounds checks if the amount is within reasonable bounds
// Maximum: int64 max (about $92 quadrillion in cents)
func ValidateAmountBounds(amount int64) error {
	const maxAmount int64 = 9223372036854775807 // int64 max
	if amount > maxAmount || amount < 0 {
		return fmt.Errorf("amount out of bounds")
	}
	return nil
}
