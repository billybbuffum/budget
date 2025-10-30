package application

import (
	"context"
	"fmt"
	"time"

	"github.com/billybbuffum/budget/internal/domain"
	"github.com/google/uuid"
)

// AllocationService handles allocation-related business logic with rollover support
type AllocationService struct {
	allocationRepo  domain.AllocationRepository
	categoryRepo    domain.CategoryRepository
	transactionRepo domain.TransactionRepository
	budgetStateRepo domain.BudgetStateRepository
}

// NewAllocationService creates a new allocation service
func NewAllocationService(
	allocationRepo domain.AllocationRepository,
	categoryRepo domain.CategoryRepository,
	transactionRepo domain.TransactionRepository,
	budgetStateRepo domain.BudgetStateRepository,
) *AllocationService {
	return &AllocationService{
		allocationRepo:  allocationRepo,
		categoryRepo:    categoryRepo,
		transactionRepo: transactionRepo,
		budgetStateRepo: budgetStateRepo,
	}
}

// CreateAllocation creates a new allocation or updates existing one for category+period
func (s *AllocationService) CreateAllocation(ctx context.Context, categoryID string, amount int64, period, notes string) (*domain.Allocation, error) {
	// Validate category exists and is an expense category
	category, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	if category.Type != domain.CategoryTypeExpense {
		return nil, fmt.Errorf("can only allocate to expense categories")
	}

	if amount < 0 {
		return nil, fmt.Errorf("allocation amount must be non-negative")
	}

	if period == "" {
		return nil, fmt.Errorf("period is required (e.g., '2024-11')")
	}

	// Check if allocation already exists for this category+period
	existing, err := s.allocationRepo.GetByCategoryAndPeriod(ctx, categoryID, period)
	if err == nil {
		// Update existing allocation
		existing.Amount = amount
		existing.Notes = notes
		existing.UpdatedAt = time.Now()
		if err := s.allocationRepo.Update(ctx, existing); err != nil {
			return nil, err
		}

		return existing, nil
	}

	// Create new allocation
	allocation := &domain.Allocation{
		ID:         uuid.New().String(),
		CategoryID: categoryID,
		Amount:     amount,
		Period:     period,
		Notes:      notes,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.allocationRepo.Create(ctx, allocation); err != nil {
		return nil, err
	}

	return allocation, nil
}

// GetAllocation retrieves an allocation by ID
func (s *AllocationService) GetAllocation(ctx context.Context, id string) (*domain.Allocation, error) {
	return s.allocationRepo.GetByID(ctx, id)
}

// ListAllocations retrieves all allocations
func (s *AllocationService) ListAllocations(ctx context.Context) ([]*domain.Allocation, error) {
	return s.allocationRepo.List(ctx)
}

// ListAllocationsByPeriod retrieves allocations for a specific period
func (s *AllocationService) ListAllocationsByPeriod(ctx context.Context, period string) ([]*domain.Allocation, error) {
	return s.allocationRepo.ListByPeriod(ctx, period)
}

// GetAllocationSummary calculates allocation summary for a period with rollover
// Shows: assigned this period, activity this period, available (with rollover)
func (s *AllocationService) GetAllocationSummary(ctx context.Context, period string) ([]*domain.AllocationSummary, error) {
	// Get all expense categories
	categories, err := s.categoryRepo.ListByType(ctx, domain.CategoryTypeExpense)
	if err != nil {
		return nil, err
	}

	var summaries []*domain.AllocationSummary

	for _, category := range categories {
		// Get allocation for this category+period (may not exist)
		allocation, _ := s.allocationRepo.GetByCategoryAndPeriod(ctx, category.ID, period)

		// Get activity for this period only
		activity, err := s.transactionRepo.GetCategoryActivity(ctx, category.ID, period)
		if err != nil {
			activity = 0 // If error, assume no activity
		}

		// Calculate available with rollover: sum ALL allocations - sum ALL transactions
		// Get all allocations for this category across all periods
		allAllocations, err := s.allocationRepo.List(ctx)
		if err != nil {
			continue
		}

		var totalAllocated int64
		for _, alloc := range allAllocations {
			if alloc.CategoryID == category.ID {
				totalAllocated += alloc.Amount
			}
		}

		// Get all transactions for this category (negative amounts are spending)
		allTransactions, err := s.transactionRepo.ListByCategory(ctx, category.ID)
		if err != nil {
			continue
		}

		var totalSpent int64
		for _, txn := range allTransactions {
			if txn.Amount < 0 {
				totalSpent += -txn.Amount // Convert to positive for display
			}
		}

		// Available = Total Allocated - Total Spent (includes rollover!)
		available := totalAllocated - totalSpent

		summary := &domain.AllocationSummary{
			Allocation: allocation, // May be nil if no allocation for this period
			Category:   category,
			Activity:   activity,   // Activity for THIS period only
			Available:  available,  // Includes rollover from previous periods
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// CalculateReadyToAssignForPeriod calculates Ready to Assign for a specific period
// Formula: (Total Income through period) - (Total Allocations through period)
// This allows future allocations without locking up current month's money
func (s *AllocationService) CalculateReadyToAssignForPeriod(ctx context.Context, period string) (int64, error) {
	// Get all transactions
	allTransactions, err := s.transactionRepo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list transactions: %w", err)
	}

	// Calculate total income through this period (positive transactions <= period)
	var totalIncome int64
	for _, txn := range allTransactions {
		// Extract period from transaction date (YYYY-MM)
		txnPeriod := txn.Date.Format("2006-01")

		// Only count income (positive amounts) through this period
		if txn.Amount > 0 && txnPeriod <= period {
			totalIncome += txn.Amount
		}
	}

	// Get all allocations
	allAllocations, err := s.allocationRepo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list allocations: %w", err)
	}

	// Calculate total allocations through this period
	var totalAllocations int64
	for _, alloc := range allAllocations {
		if alloc.Period <= period {
			totalAllocations += alloc.Amount
		}
	}

	// Ready to Assign = Income - Allocations
	// This can be negative if over-allocated!
	return totalIncome - totalAllocations, nil
}

// GetReadyToAssign reads the Ready to Assign amount from the database
// DEPRECATED: This now returns 0 as Ready to Assign is calculated per-period
// Use CalculateReadyToAssignForPeriod instead
func (s *AllocationService) GetReadyToAssign(ctx context.Context) (int64, error) {
	state, err := s.budgetStateRepo.Get(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get budget state: %w", err)
	}
	return state.ReadyToAssign, nil
}

// DeleteAllocation deletes an allocation
func (s *AllocationService) DeleteAllocation(ctx context.Context, id string) error {
	// Delete the allocation
	if err := s.allocationRepo.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}
