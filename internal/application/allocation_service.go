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
	accountRepo     domain.AccountRepository
}

// NewAllocationService creates a new allocation service
func NewAllocationService(
	allocationRepo domain.AllocationRepository,
	categoryRepo domain.CategoryRepository,
	transactionRepo domain.TransactionRepository,
	budgetStateRepo domain.BudgetStateRepository,
	accountRepo domain.AccountRepository,
) *AllocationService {
	return &AllocationService{
		allocationRepo:  allocationRepo,
		categoryRepo:    categoryRepo,
		transactionRepo: transactionRepo,
		budgetStateRepo: budgetStateRepo,
		accountRepo:     accountRepo,
	}
}

// CreateAllocation creates a new allocation or updates existing one for category+period
func (s *AllocationService) CreateAllocation(ctx context.Context, categoryID string, amount int64, period, notes string) (*domain.Allocation, error) {
	// Validate category exists
	_, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
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

// AllocateToCoverUnderfunded creates an allocation to cover an underfunded payment category
// This method manually allocates funds from Ready to Assign to a payment category that
// doesn't have enough money to cover its associated credit card balance.
//
// Returns:
//   - allocation: The created or updated allocation
//   - underfundedAmount: The amount that was underfunded (and now covered)
//   - error: Any error that occurred
func (s *AllocationService) AllocateToCoverUnderfunded(
	ctx context.Context,
	paymentCategoryID string,
	period string,
) (*domain.Allocation, int64, error) {
	// 1. Validate that the category exists and is a payment category
	category, err := s.categoryRepo.GetByID(ctx, paymentCategoryID)
	if err != nil {
		return nil, 0, fmt.Errorf("payment category not found")
	}

	if category.PaymentForAccountID == nil || *category.PaymentForAccountID == "" {
		return nil, 0, fmt.Errorf("category is not a payment category")
	}

	// 2. Calculate the underfunded amount using GetAllocationSummary
	summaries, err := s.GetAllocationSummary(ctx, period)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to calculate allocation summary: %w", err)
	}

	var underfundedAmount int64
	var found bool
	for _, summary := range summaries {
		if summary.Category.ID == paymentCategoryID {
			found = true
			if summary.Underfunded != nil && *summary.Underfunded > 0 {
				underfundedAmount = *summary.Underfunded
			}
			break
		}
	}

	if !found {
		return nil, 0, fmt.Errorf("payment category not found in summary")
	}

	// 3. Check that there is actually an underfunded amount
	if underfundedAmount <= 0 {
		return nil, 0, fmt.Errorf("payment category is not underfunded")
	}

	// 4. Calculate Ready to Assign to verify sufficient funds
	readyToAssign, err := s.CalculateReadyToAssignForPeriod(ctx, period)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to calculate Ready to Assign: %w", err)
	}

	// 5. Verify that Ready to Assign has sufficient funds
	if readyToAssign < underfundedAmount {
		return nil, 0, fmt.Errorf(
			"insufficient funds: Ready to Assign: $%.2f, Underfunded: $%.2f",
			float64(readyToAssign)/100,
			float64(underfundedAmount)/100,
		)
	}

	// 6. Create or update the allocation (upsert behavior)
	// Use CreateAllocation which already has upsert logic
	allocation, err := s.CreateAllocation(
		ctx,
		paymentCategoryID,
		underfundedAmount,
		period,
		"Cover underfunded credit card spending",
	)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create allocation: %w", err)
	}

	return allocation, underfundedAmount, nil
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
	// Get all categories
	categories, err := s.categoryRepo.List(ctx)
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
			// For category-specific available: COUNT all spending including transfers
			// Transfers DO reduce what's available in a specific category
			// (This is different from Ready to Assign, which excludes transfers)
			if txn.Amount < 0 {
				totalSpent += -txn.Amount // Convert to positive for display
			}
		}

		// Available = Total Allocated - Total Spent (includes rollover!)
		available := totalAllocated - totalSpent

		// For payment categories, check if underfunded (available < credit card balance)
		var underfunded *int64
		var underfundedCategories []string
		if category.PaymentForAccountID != nil && *category.PaymentForAccountID != "" {
			// Get the credit card account balance
			account, err := s.accountRepo.GetByID(ctx, *category.PaymentForAccountID)
			if err == nil && account != nil {
				// Credit card balance is negative (you owe money)
				// We need enough AVAILABLE (not just allocated) to cover the balance
				amountOwed := -account.Balance // Convert to positive

				if amountOwed > 0 && available < amountOwed {
					// Underfunded: need more money
					shortfall := amountOwed - available
					underfunded = &shortfall

					// Find which expense categories are underfunded
					// Get all transactions on this credit card
					ccTransactions, err := s.transactionRepo.ListByAccount(ctx, *category.PaymentForAccountID)
					if err == nil {
						// Group by category and calculate spending per category
						categorySpending := make(map[string]int64)
						categoryNames := make(map[string]string)

						for _, txn := range ccTransactions {
							if txn.CategoryID != nil && *txn.CategoryID != "" && txn.Amount < 0 {
								// This is spending on an expense category
								categorySpending[*txn.CategoryID] += -txn.Amount // Convert to positive

								// Get category name
								if _, exists := categoryNames[*txn.CategoryID]; !exists {
									cat, err := s.categoryRepo.GetByID(ctx, *txn.CategoryID)
									if err == nil {
										categoryNames[*txn.CategoryID] = cat.Name
									}
								}
							}
						}

						// Check each category to see if it has enough allocated
						for catID, spending := range categorySpending {
							// Get all allocations for this category
							allAllocForCat, err := s.allocationRepo.List(ctx)
							if err == nil {
								var catTotalAllocated int64
								for _, alloc := range allAllocForCat {
									if alloc.CategoryID == catID {
										catTotalAllocated += alloc.Amount
									}
								}

								// If spending exceeds allocation, this category is underfunded
								if spending > catTotalAllocated {
									if name, exists := categoryNames[catID]; exists {
										underfundedCategories = append(underfundedCategories, name)
									}
								}
							}
						}
					}
				}
			}
		}

		summary := &domain.AllocationSummary{
			Allocation:            allocation,             // May be nil if no allocation for this period
			Category:              category,
			Activity:              activity,               // Activity for THIS period only
			Available:             available,              // Includes rollover from previous periods
			Underfunded:           underfunded,            // Amount needed to cover CC balance (nil if not underfunded)
			UnderfundedCategories: underfundedCategories,  // List of categories needing more allocation
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// CalculateReadyToAssignForPeriod calculates Ready to Assign for a specific period
// Formula: Total Account Balance - (Total Allocations through period - Total Spent through period)
// This represents: "How much money do I have that isn't allocated to a category?"
// Note: This calculation ignores future periods to allow forward budgeting
func (s *AllocationService) CalculateReadyToAssignForPeriod(ctx context.Context, period string) (int64, error) {
	// Ready to Assign = Total Inflows - Total Allocated
	// This shows how much INCOME is available to allocate, not account balance.
	// Account balance is lower due to spending, but inflows are what you budget from.

	// Get all transactions to calculate inflows
	allTransactions, err := s.transactionRepo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list transactions: %w", err)
	}

	// Calculate total inflows through this period
	// Only count positive amounts (inflows), exclude transfers
	var totalInflows int64
	for _, txn := range allTransactions {
		txnPeriod := txn.Date.Format("2006-01")
		if txn.Amount > 0 && txnPeriod <= period && txn.Type != "transfer" {
			totalInflows += txn.Amount
		}
	}

	// Get all allocations through this period
	allAllocations, err := s.allocationRepo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list allocations: %w", err)
	}

	// Get all categories to check which are payment categories
	allCategories, err := s.categoryRepo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list categories: %w", err)
	}

	// Build map of payment category IDs
	paymentCategoryIDs := make(map[string]bool)
	for _, cat := range allCategories {
		if cat.PaymentForAccountID != nil && *cat.PaymentForAccountID != "" {
			paymentCategoryIDs[cat.ID] = true
		}
	}

	// Calculate total allocations through this period
	// EXCLUDE payment category allocations - they don't represent "new" money allocation
	// Payment allocations just track CC debt covered by expense budgets
	var totalAllocations int64
	for _, alloc := range allAllocations {
		if alloc.Period <= period && !paymentCategoryIDs[alloc.CategoryID] {
			totalAllocations += alloc.Amount
		}
	}

	// Ready to Assign = Total Inflows - Total Allocated
	// This can be negative if you over-allocated!
	// When you categorize unbudgeted spending, categories go negative (overspent).
	// You must then allocate money to cover the overspending, reducing RTA.
	readyToAssign := totalInflows - totalAllocations

	// Check for underfunded credit cards and subtract their shortfalls
	// This ensures the indicator reflects CC debt that isn't properly covered
	summaries, err := s.GetAllocationSummary(ctx, period)
	if err == nil {
		for _, summary := range summaries {
			if summary.Underfunded != nil && *summary.Underfunded > 0 {
				// Reduce ready to assign by the underfunded amount
				// This makes underfunded CC spending equivalent to over-allocation
				readyToAssign -= *summary.Underfunded
			}
		}
	}

	return readyToAssign, nil
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
