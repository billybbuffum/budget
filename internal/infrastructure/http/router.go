package http

import (
	"net/http"

	"github.com/billybbuffum/budget/internal/infrastructure/http/handlers"
)

// NewRouter creates and configures the HTTP router
func NewRouter(
	accountHandler *handlers.AccountHandler,
	categoryHandler *handlers.CategoryHandler,
	categoryGroupHandler *handlers.CategoryGroupHandler,
	transactionHandler *handlers.TransactionHandler,
	allocationHandler *handlers.AllocationHandler,
	importHandler *handlers.ImportHandler,
) *http.ServeMux {
	mux := http.NewServeMux()

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/", fs)

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Account routes
	mux.HandleFunc("POST /api/accounts", accountHandler.CreateAccount)
	mux.HandleFunc("GET /api/accounts", accountHandler.ListAccounts)
	mux.HandleFunc("GET /api/accounts/summary", accountHandler.GetSummary)
	mux.HandleFunc("GET /api/accounts/{id}", accountHandler.GetAccount)
	mux.HandleFunc("PUT /api/accounts/{id}", accountHandler.UpdateAccount)
	mux.HandleFunc("DELETE /api/accounts/{id}", accountHandler.DeleteAccount)

	// Category routes
	mux.HandleFunc("POST /api/categories", categoryHandler.CreateCategory)
	mux.HandleFunc("GET /api/categories", categoryHandler.ListCategories)
	mux.HandleFunc("GET /api/categories/{id}", categoryHandler.GetCategory)
	mux.HandleFunc("PUT /api/categories/{id}", categoryHandler.UpdateCategory)
	mux.HandleFunc("DELETE /api/categories/{id}", categoryHandler.DeleteCategory)
	mux.HandleFunc("POST /api/categories/{id}/restore", categoryHandler.RestoreCategory)

	// Category Group routes
	mux.HandleFunc("POST /api/category-groups", categoryGroupHandler.CreateCategoryGroup)
	mux.HandleFunc("GET /api/category-groups", categoryGroupHandler.ListCategoryGroups)
	mux.HandleFunc("GET /api/category-groups/{id}", categoryGroupHandler.GetCategoryGroup)
	mux.HandleFunc("PUT /api/category-groups/{id}", categoryGroupHandler.UpdateCategoryGroup)
	mux.HandleFunc("DELETE /api/category-groups/{id}", categoryGroupHandler.DeleteCategoryGroup)
	mux.HandleFunc("POST /api/category-groups/assign", categoryGroupHandler.AssignCategoryToGroup)
	mux.HandleFunc("POST /api/category-groups/unassign/{id}", categoryGroupHandler.UnassignCategoryFromGroup)

	// Transaction routes
	mux.HandleFunc("POST /api/transactions", transactionHandler.CreateTransaction)
	mux.HandleFunc("POST /api/transactions/transfer", transactionHandler.CreateTransfer)
	mux.HandleFunc("GET /api/transactions", transactionHandler.ListTransactions)
	mux.HandleFunc("GET /api/transactions/{id}", transactionHandler.GetTransaction)
	mux.HandleFunc("PUT /api/transactions/{id}", transactionHandler.UpdateTransaction)
	mux.HandleFunc("DELETE /api/transactions/{id}", transactionHandler.DeleteTransaction)
	mux.HandleFunc("POST /api/transactions/bulk-categorize", transactionHandler.BulkCategorizeTransactions)

	// Import routes
	mux.HandleFunc("POST /api/transactions/import", importHandler.ImportTransactions)

	// Allocation routes
	mux.HandleFunc("POST /api/allocations", allocationHandler.CreateAllocation)
	mux.HandleFunc("GET /api/allocations", allocationHandler.ListAllocations)
	mux.HandleFunc("GET /api/allocations/summary", allocationHandler.GetAllocationSummary)
	mux.HandleFunc("GET /api/allocations/ready-to-assign", allocationHandler.GetReadyToAssign)
	mux.HandleFunc("GET /api/allocations/{id}", allocationHandler.GetAllocation)
	mux.HandleFunc("DELETE /api/allocations/{id}", allocationHandler.DeleteAllocation)

	return mux
}
