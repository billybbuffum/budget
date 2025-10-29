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
		// Update existing allocation - adjust Ready to Assign by the difference
		oldAmount := existing.Amount
		delta := amount - oldAmount

		existing.Amount = amount
		existing.Notes = notes
		existing.UpdatedAt = time.Now()
		if err := s.allocationRepo.Update(ctx, existing); err != nil {
			return nil, err
		}

		// Adjust Ready to Assign by the difference (decrease if allocating more)
		if delta != 0 {
			if err := s.budgetStateRepo.AdjustReadyToAssign(ctx, -delta); err != nil {
				return nil, fmt.Errorf("failed to adjust ready to assign: %w", err)
			}
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

	// Decrease Ready to Assign by the allocated amount
	// Backend coordinates this automatically!
	if err := s.budgetStateRepo.AdjustReadyToAssign(ctx, -amount); err != nil {
		// Rollback if Ready to Assign update fails
		s.allocationRepo.Delete(ctx, allocation.ID)
		return nil, fmt.Errorf("failed to adjust ready to assign: %w", err)
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

// GetReadyToAssign reads the Ready to Assign amount from the database
// The backend automatically coordinates this value when transactions and allocations change
func (s *AllocationService) GetReadyToAssign(ctx context.Context, accountRepo domain.AccountRepository) (int64, error) {
	state, err := s.budgetStateRepo.Get(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get budget state: %w", err)
	}
	return state.ReadyToAssign, nil
}

// DeleteAllocation deletes an allocation and returns money to Ready to Assign
func (s *AllocationService) DeleteAllocation(ctx context.Context, id string) error {
	// Get the allocation to know how much to add back to Ready to Assign
	allocation, err := s.allocationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete the allocation
	if err := s.allocationRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Return the allocated amount to Ready to Assign
	if err := s.budgetStateRepo.AdjustReadyToAssign(ctx, allocation.Amount); err != nil {
		return fmt.Errorf("failed to adjust ready to assign: %w", err)
	}

	return nil
}
