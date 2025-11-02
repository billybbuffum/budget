package validators

import (
	"testing"
	"time"
)

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "valid UUID v4",
			id:      "550e8400-e29b-41d4-a716-446655440000",
			wantErr: false,
		},
		{
			name:    "valid UUID v1",
			id:      "c56a4180-65aa-42ec-a945-5fd21dec0538",
			wantErr: false,
		},
		{
			name:    "empty string",
			id:      "",
			wantErr: true,
		},
		{
			name:    "malformed UUID - missing segments",
			id:      "550e8400-e29b-41d4-a716",
			wantErr: true,
		},
		{
			name:    "malformed UUID - wrong format",
			id:      "550e8400e29b41d4a716446655440000",
			wantErr: false, // uuid.Parse accepts UUIDs without dashes
		},
		{
			name:    "not a UUID - random string",
			id:      "not-a-valid-uuid",
			wantErr: true,
		},
		{
			name:    "not a UUID - integer",
			id:      "12345",
			wantErr: true,
		},
		{
			name:    "invalid characters in UUID",
			id:      "550e8400-e29b-41d4-a716-44665544000g",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUUID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUUID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != "invalid UUID format" {
				t.Errorf("ValidateUUID() error message = %v, want 'invalid UUID format'", err.Error())
			}
		})
	}
}

func TestValidatePeriodFormat(t *testing.T) {
	tests := []struct {
		name    string
		period  string
		wantErr bool
	}{
		{
			name:    "valid period - January",
			period:  "2024-01",
			wantErr: false,
		},
		{
			name:    "valid period - December",
			period:  "2024-12",
			wantErr: false,
		},
		{
			name:    "valid period - current year",
			period:  "2025-10",
			wantErr: false,
		},
		{
			name:    "invalid month - 13",
			period:  "2024-13",
			wantErr: true,
		},
		{
			name:    "invalid month - 00",
			period:  "2024-00",
			wantErr: true,
		},
		{
			name:    "invalid format - two digit year",
			period:  "24-01",
			wantErr: true,
		},
		{
			name:    "invalid format - single digit month",
			period:  "2024-1",
			wantErr: true,
		},
		{
			name:    "invalid format - slash separator",
			period:  "2024/01",
			wantErr: true,
		},
		{
			name:    "empty string",
			period:  "",
			wantErr: true,
		},
		{
			name:    "invalid format - no separator",
			period:  "202401",
			wantErr: true,
		},
		{
			name:    "invalid format - extra characters",
			period:  "2024-01-15",
			wantErr: true,
		},
		{
			name:    "invalid format - letters in year",
			period:  "ABCD-01",
			wantErr: true,
		},
		{
			name:    "invalid format - letters in month",
			period:  "2024-AB",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePeriodFormat(tt.period)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePeriodFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != "invalid period format, expected YYYY-MM" {
				t.Errorf("ValidatePeriodFormat() error message = %v, want 'invalid period format, expected YYYY-MM'", err.Error())
			}
		})
	}
}

func TestValidatePeriodRange(t *testing.T) {
	now := time.Now()

	// Calculate test periods
	currentPeriod := now.Format("2006-01")
	oneYearAgo := now.AddDate(-1, 0, 0).Format("2006-01")
	twoYearsAgo := now.AddDate(-2, 0, 0).Format("2006-01")
	threeYearsAgo := now.AddDate(-3, 0, 0).Format("2006-01") // Should fail
	fourYearsFuture := now.AddDate(4, 0, 0).Format("2006-01")
	fiveYearsFuture := now.AddDate(5, 0, 0).Format("2006-01")
	sixYearsFuture := now.AddDate(6, 0, 0).Format("2006-01") // Should fail

	tests := []struct {
		name    string
		period  string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid - current period",
			period:  currentPeriod,
			wantErr: false,
		},
		{
			name:    "valid - one year ago",
			period:  oneYearAgo,
			wantErr: false,
		},
		{
			name:    "valid - two years ago (boundary)",
			period:  twoYearsAgo,
			wantErr: false,
		},
		{
			name:    "invalid - three years ago",
			period:  threeYearsAgo,
			wantErr: true,
			errMsg:  "period is too far in the past (more than 2 years)",
		},
		{
			name:    "valid - four years in future",
			period:  fourYearsFuture,
			wantErr: false,
		},
		{
			name:    "valid - five years in future (boundary)",
			period:  fiveYearsFuture,
			wantErr: false,
		},
		{
			name:    "invalid - six years in future",
			period:  sixYearsFuture,
			wantErr: true,
			errMsg:  "period is too far in the future (more than 5 years)",
		},
		{
			name:    "invalid - bad format should fail",
			period:  "2024-13",
			wantErr: true,
			errMsg:  "invalid period format, expected YYYY-MM",
		},
		{
			name:    "invalid - empty string",
			period:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePeriodRange(tt.period)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePeriodRange() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("ValidatePeriodRange() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidateAmountPositive(t *testing.T) {
	tests := []struct {
		name    string
		amount  int64
		wantErr bool
	}{
		{
			name:    "valid - small positive amount",
			amount:  1,
			wantErr: false,
		},
		{
			name:    "valid - typical amount",
			amount:  100,
			wantErr: false,
		},
		{
			name:    "valid - large amount",
			amount:  1000000000,
			wantErr: false,
		},
		{
			name:    "valid - max int64",
			amount:  9223372036854775807,
			wantErr: false,
		},
		{
			name:    "invalid - zero",
			amount:  0,
			wantErr: true,
		},
		{
			name:    "invalid - negative small",
			amount:  -1,
			wantErr: true,
		},
		{
			name:    "invalid - negative large",
			amount:  -100,
			wantErr: true,
		},
		{
			name:    "invalid - negative max",
			amount:  -9223372036854775808,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAmountPositive(tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAmountPositive() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != "amount must be positive and greater than zero" {
				t.Errorf("ValidateAmountPositive() error message = %v, want 'amount must be positive and greater than zero'", err.Error())
			}
		})
	}
}

func TestValidateAmountBounds(t *testing.T) {
	tests := []struct {
		name    string
		amount  int64
		wantErr bool
	}{
		{
			name:    "valid - zero",
			amount:  0,
			wantErr: false,
		},
		{
			name:    "valid - small positive",
			amount:  100,
			wantErr: false,
		},
		{
			name:    "valid - large positive",
			amount:  1000000000,
			wantErr: false,
		},
		{
			name:    "valid - max int64",
			amount:  9223372036854775807,
			wantErr: false,
		},
		{
			name:    "invalid - negative",
			amount:  -1,
			wantErr: true,
		},
		{
			name:    "invalid - negative large",
			amount:  -100,
			wantErr: true,
		},
		{
			name:    "invalid - min int64",
			amount:  -9223372036854775808,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAmountBounds(tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAmountBounds() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && err.Error() != "amount out of bounds" {
				t.Errorf("ValidateAmountBounds() error message = %v, want 'amount out of bounds'", err.Error())
			}
		})
	}
}
