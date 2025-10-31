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

	// Build category name map to avoid N database calls in underfunded calculation
	categoryNames := make(map[string]string)
	for _, cat := range categories {
		categoryNames[cat.ID] = cat.Name
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
			return nil, fmt.Errorf("failed to list allocations for category %s: %w", category.ID, err)
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
			return nil, fmt.Errorf("failed to list transactions for category %s: %w", category.ID, err)
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
				// Get all transactions on this credit card
				ccTransactions, err := s.transactionRepo.ListByAccount(ctx, *category.PaymentForAccountID)
				if err == nil {
					// Group by category and calculate spending per category
					categorySpending := make(map[string]int64)

					for _, txn := range ccTransactions {
						if txn.CategoryID != nil && *txn.CategoryID != "" && txn.Amount < 0 {
							// This is spending on an expense category
							categorySpending[*txn.CategoryID] += -txn.Amount // Convert to positive
						}
					}

					// Get all allocations (we'll need this for multiple categories)
					allAllocations, err := s.allocationRepo.List(ctx)
					if err == nil {
						// Calculate budgeted CC spending per category
						// This is the amount of CC spending that's covered by expense allocations
						budgetedPerCategory := make(map[string]int64)
						for catID, spending := range categorySpending {
							var catAllocation int64
							for _, alloc := range allAllocations {
								if alloc.CategoryID == catID {
									catAllocation += alloc.Amount
								}
							}

							// Category's contribution to covering CC debt
							// It's the minimum of spending and allocation
							contribution := spending
							if catAllocation < spending {
								contribution = catAllocation
							}
							budgetedPerCategory[catID] = contribution
						}

						// Sum total budgeted vs total spending
						var totalBudgeted, totalSpending int64
						for catID, spending := range categorySpending {
							totalSpending += spending
							totalBudgeted += budgetedPerCategory[catID]
						}

						// Calculate unbudgeted debt: spending that's NOT covered by expense allocations
						unbudgetedDebt := totalSpending - totalBudgeted

						// Underfunded = unbudgeted debt minus what's already in payment category
						if unbudgetedDebt > available {
							shortfall := unbudgetedDebt - available
							underfunded = &shortfall

							// Update underfundedCategories to only show truly underfunded
							// (categories where spending exceeds allocation)
							for catID, spending := range categorySpending {
								if spending > budgetedPerCategory[catID] {
									// This category has unbudgeted spending
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

	// Calculate total allocations through this period
	// After removing auto-sync, payment category allocations are now manual user actions
	// They represent money the user has intentionally set aside to pay CC bills
	// These must be included in total allocations to reduce Ready to Assign
	var totalAllocations int64
	for _, alloc := range allAllocations {
		if alloc.Period <= period {
			totalAllocations += alloc.Amount
		}
	}

	// Ready to Assign = Total Inflows - Total Allocated
	// This can be negative if you over-allocated!
	// When you categorize unbudgeted spending, categories go negative (overspent).
	// You must then allocate money to cover the overspending, reducing RTA.
	readyToAssign := totalInflows - totalAllocations

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

// CoverUnderfundedPayment creates an allocation to cover underfunded credit card spending
// This is a manual, user-initiated action to allocate money to a payment category
func (s *AllocationService) CoverUnderfundedPayment(ctx context.Context, paymentCategoryID string, period string) (*domain.Allocation, error) {
	// 1. Verify this is a payment category
	category, err := s.categoryRepo.GetByID(ctx, paymentCategoryID)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	if category.PaymentForAccountID == nil || *category.PaymentForAccountID == "" {
		return nil, fmt.Errorf("category is not a payment category")
	}

	// 2. Get underfunded amount from allocation summary
	summaries, err := s.GetAllocationSummary(ctx, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get allocation summary: %w", err)
	}

	var underfundedAmount int64
	var found bool
	for _, summary := range summaries {
		if summary.Category.ID == paymentCategoryID {
			if summary.Underfunded != nil && *summary.Underfunded > 0 {
				underfundedAmount = *summary.Underfunded
				found = true
			}
			break
		}
	}

	if !found || underfundedAmount == 0 {
		return nil, fmt.Errorf("payment category is not underfunded")
	}

	// 3. Check Ready to Assign
	readyToAssign, err := s.CalculateReadyToAssignForPeriod(ctx, period)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate ready to assign: %w", err)
	}

	if readyToAssign < underfundedAmount {
		return nil, fmt.Errorf("insufficient funds: need %d cents but only %d cents available", underfundedAmount, readyToAssign)
	}

	// 4. Get existing allocation for this period (if any) and add to it
	// The underfundedAmount is the SHORTFALL (additional amount needed)
	// We need to ADD this to any existing allocation for the period
	existingAllocation, err := s.allocationRepo.GetByCategoryAndPeriod(ctx, paymentCategoryID, period)
	existingAmount := int64(0)
	if err == nil && existingAllocation != nil {
		existingAmount = existingAllocation.Amount
	}

	// New total = existing + shortfall
	newTotalAmount := existingAmount + underfundedAmount

	// Create/update allocation with new total
	allocation, err := s.CreateAllocation(ctx, paymentCategoryID, newTotalAmount, period, "Cover underfunded credit card spending")
	if err != nil {
		return nil, fmt.Errorf("failed to create allocation: %w", err)
	}

	return allocation, nil
}
