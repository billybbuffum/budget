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

		// RETROACTIVE PAYMENT CATEGORY ALLOCATION
		// Sync payment categories when updating allocation
		if err := s.syncPaymentCategoryAllocations(ctx, categoryID); err != nil {
			fmt.Printf("Warning: failed to sync payment category allocations: %v\n", err)
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

	// RETROACTIVE PAYMENT CATEGORY ALLOCATION
	// When allocating to an expense category, check for CC spending and allocate to payment categories
	if err := s.syncPaymentCategoryAllocations(ctx, categoryID); err != nil {
		// Log error but don't fail the allocation
		// The payment category allocation can be synced later
		fmt.Printf("Warning: failed to sync payment category allocations: %v\n", err)
	}

	return allocation, nil
}

// syncPaymentCategoryAllocations checks for credit card spending on a category
// and allocates to payment categories retroactively
func (s *AllocationService) syncPaymentCategoryAllocations(ctx context.Context, categoryID string) error {
	// Get all credit card accounts
	allAccounts, err := s.accountRepo.List(ctx)
	if err != nil {
		return err
	}

	for _, account := range allAccounts {
		if account.Type != domain.AccountTypeCredit {
			continue
		}

		// Get payment category for this CC
		paymentCategory, err := s.categoryRepo.GetPaymentCategoryByAccountID(ctx, account.ID)
		if err != nil || paymentCategory == nil {
			continue
		}

		// Get all transactions on this CC for the given expense category
		ccTransactions, err := s.transactionRepo.ListByAccount(ctx, account.ID)
		if err != nil {
			continue
		}

		var totalCCSpending int64
		for _, txn := range ccTransactions {
			if txn.CategoryID != nil && *txn.CategoryID == categoryID && txn.Amount < 0 {
				totalCCSpending += -txn.Amount // Convert to positive
			}
		}

		if totalCCSpending == 0 {
			continue
		}

		// Get the expense category's total allocation
		allAllocations, err := s.allocationRepo.List(ctx)
		if err != nil {
			continue
		}

		var expenseCategoryAllocated int64
		for _, alloc := range allAllocations {
			if alloc.CategoryID == categoryID {
				expenseCategoryAllocated += alloc.Amount
			}
		}

		// Calculate how much to allocate to payment category
		// Move the minimum of: CC spending or expense category allocation
		amountToAllocate := totalCCSpending
		if expenseCategoryAllocated < totalCCSpending {
			amountToAllocate = expenseCategoryAllocated
		}

		if amountToAllocate == 0 {
			continue
		}

		// Update payment category allocation by adding this category's contribution
		// Find the allocation for the current period (or create one)
		period := time.Now().Format("2006-01")
		paymentAlloc, err := s.allocationRepo.GetByCategoryAndPeriod(ctx, paymentCategory.ID, period)

		if err != nil {
			// Create new payment category allocation
			newAlloc := &domain.Allocation{
				ID:         uuid.New().String(),
				CategoryID: paymentCategory.ID,
				Amount:     amountToAllocate,
				Period:     period,
				Notes:      "Auto-allocated from retroactive CC spending",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			if err := s.allocationRepo.Create(ctx, newAlloc); err != nil {
				return err
			}
		} else {
			// Update existing allocation
			// Calculate how much of the payment allocation comes from THIS expense category
			// by checking previous allocations. We need to recalculate the entire payment allocation.

			// Get ALL expense categories' CC spending and allocations
			var totalShouldBeAllocated int64
			ccTxns, _ := s.transactionRepo.ListByAccount(ctx, account.ID)
			categoryContributions := make(map[string]int64)

			for _, txn := range ccTxns {
				if txn.CategoryID != nil && *txn.CategoryID != "" && txn.Amount < 0 {
					categoryContributions[*txn.CategoryID] += -txn.Amount
				}
			}

			for catID, spending := range categoryContributions {
				var catAlloc int64
				for _, alloc := range allAllocations {
					if alloc.CategoryID == catID {
						catAlloc += alloc.Amount
					}
				}

				contribution := spending
				if catAlloc < spending {
					contribution = catAlloc
				}
				totalShouldBeAllocated += contribution
			}

			// Update payment category to match total
			if totalShouldBeAllocated != paymentAlloc.Amount {
				paymentAlloc.Amount = totalShouldBeAllocated
				paymentAlloc.UpdatedAt = time.Now()

				if err := s.allocationRepo.Update(ctx, paymentAlloc); err != nil {
					return err
				}
			}
		}
	}

	return nil
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
	// Get total account balance (all money available)
	allAccounts, err := s.accountRepo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list accounts: %w", err)
	}

	var totalAccountBalance int64
	for _, account := range allAccounts {
		totalAccountBalance += account.Balance
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

	// Get all transactions through this period
	allTransactions, err := s.transactionRepo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list transactions: %w", err)
	}

	// Calculate total spent through this period (negative transactions = spending)
	// IMPORTANT: Exclude transfers - they move money between accounts but don't "free up" allocated funds
	var totalSpent int64
	for _, txn := range allTransactions {
		// Extract period from transaction date (YYYY-MM)
		txnPeriod := txn.Date.Format("2006-01")

		// Only count actual spending (negative amounts) through this period
		// Skip transfers - they just move money between accounts
		if txn.Amount < 0 && txnPeriod <= period && txn.Type != "transfer" {
			totalSpent += -txn.Amount // Convert to positive
		}
	}

	// Ready to Assign = Total Account Balance - (Allocated - Spent)
	// This can be negative if over-allocated!
	readyToAssign := totalAccountBalance - (totalAllocations - totalSpent)

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
