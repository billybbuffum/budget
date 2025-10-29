package http

import (
	"net/http"

	"github.com/billybbuffum/budget/internal/infrastructure/http/handlers"
)

// NewRouter creates and configures the HTTP router
func NewRouter(
	userHandler *handlers.UserHandler,
	categoryHandler *handlers.CategoryHandler,
	transactionHandler *handlers.TransactionHandler,
	budgetHandler *handlers.BudgetHandler,
) *http.ServeMux {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// User routes
	mux.HandleFunc("POST /api/users", userHandler.CreateUser)
	mux.HandleFunc("GET /api/users", userHandler.ListUsers)
	mux.HandleFunc("GET /api/users/{id}", userHandler.GetUser)
	mux.HandleFunc("PUT /api/users/{id}", userHandler.UpdateUser)
	mux.HandleFunc("DELETE /api/users/{id}", userHandler.DeleteUser)

	// Category routes
	mux.HandleFunc("POST /api/categories", categoryHandler.CreateCategory)
	mux.HandleFunc("GET /api/categories", categoryHandler.ListCategories)
	mux.HandleFunc("GET /api/categories/{id}", categoryHandler.GetCategory)
	mux.HandleFunc("PUT /api/categories/{id}", categoryHandler.UpdateCategory)
	mux.HandleFunc("DELETE /api/categories/{id}", categoryHandler.DeleteCategory)

	// Transaction routes
	mux.HandleFunc("POST /api/transactions", transactionHandler.CreateTransaction)
	mux.HandleFunc("GET /api/transactions", transactionHandler.ListTransactions)
	mux.HandleFunc("GET /api/transactions/{id}", transactionHandler.GetTransaction)
	mux.HandleFunc("PUT /api/transactions/{id}", transactionHandler.UpdateTransaction)
	mux.HandleFunc("DELETE /api/transactions/{id}", transactionHandler.DeleteTransaction)

	// Budget routes
	mux.HandleFunc("POST /api/budgets", budgetHandler.CreateBudget)
	mux.HandleFunc("GET /api/budgets", budgetHandler.ListBudgets)
	mux.HandleFunc("GET /api/budgets/summary", budgetHandler.GetBudgetSummary)
	mux.HandleFunc("GET /api/budgets/{id}", budgetHandler.GetBudget)
	mux.HandleFunc("PUT /api/budgets/{id}", budgetHandler.UpdateBudget)
	mux.HandleFunc("DELETE /api/budgets/{id}", budgetHandler.DeleteBudget)

	return mux
}
