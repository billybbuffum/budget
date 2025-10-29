package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/billybbuffum/budget/config"
	"github.com/billybbuffum/budget/internal/application"
	"github.com/billybbuffum/budget/internal/infrastructure/database"
	"github.com/billybbuffum/budget/internal/infrastructure/http"
	"github.com/billybbuffum/budget/internal/infrastructure/http/handlers"
	"github.com/billybbuffum/budget/internal/infrastructure/repository"
)

func main() {
	// Load configuration
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Initialize database
	db, err := database.NewSQLiteDB(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Println("Database initialized successfully")

	// Initialize repositories
	accountRepo := repository.NewAccountRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	allocationRepo := repository.NewAllocationRepository(db)
	budgetStateRepo := repository.NewBudgetStateRepository(db)

	// Initialize services
	accountService := application.NewAccountService(accountRepo, budgetStateRepo)
	categoryService := application.NewCategoryService(categoryRepo)
	transactionService := application.NewTransactionService(transactionRepo, accountRepo, categoryRepo, budgetStateRepo)
	allocationService := application.NewAllocationService(allocationRepo, categoryRepo, transactionRepo, budgetStateRepo)

	// Initialize handlers
	accountHandler := handlers.NewAccountHandler(accountService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	allocationHandler := handlers.NewAllocationHandler(allocationService)

	// Setup router
	router := http.NewRouter(accountHandler, categoryHandler, transactionHandler, allocationHandler)

	// Create server
	server := http.NewServer(fmt.Sprintf(":%s", cfg.Server.Port), router)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Received shutdown signal")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
