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
	"github.com/billybbuffum/budget/internal/infrastructure/ofx"
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
	categoryGroupRepo := repository.NewCategoryGroupRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	allocationRepo := repository.NewAllocationRepository(db)
	budgetStateRepo := repository.NewBudgetStateRepository(db)
	transferSuggestionRepo := repository.NewTransferSuggestionRepository(db)

	// Initialize default data
	bootstrapService := application.NewBootstrapService(categoryGroupRepo, categoryRepo)
	ctx := context.Background()
	if err := bootstrapService.InitializeDefaultData(ctx); err != nil {
		log.Fatalf("Failed to initialize default data: %v", err)
	}
	log.Println("Default data initialized successfully")

	// Initialize OFX parser
	ofxParser := ofx.NewParser()

	// Initialize services
	categoryService := application.NewCategoryService(categoryRepo)
	categoryGroupService := application.NewCategoryGroupService(categoryGroupRepo, categoryRepo)
	accountService := application.NewAccountService(accountRepo, categoryRepo, budgetStateRepo, transactionRepo, categoryGroupService)
	transactionService := application.NewTransactionService(transactionRepo, accountRepo, categoryRepo, allocationRepo, budgetStateRepo)
	allocationService := application.NewAllocationService(allocationRepo, categoryRepo, transactionRepo, budgetStateRepo, accountRepo)

	// Initialize transfer services
	transferMatcherService := application.NewTransferMatcherService(transactionRepo, accountRepo, transferSuggestionRepo)
	transferLinkService := application.NewTransferLinkService(transactionRepo, accountRepo, categoryRepo, allocationRepo, transferSuggestionRepo)
	importService := application.NewImportService(transactionRepo, accountRepo, budgetStateRepo, ofxParser, transferMatcherService)

	// Initialize handlers
	accountHandler := handlers.NewAccountHandler(accountService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	categoryGroupHandler := handlers.NewCategoryGroupHandler(categoryGroupService)
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	allocationHandler := handlers.NewAllocationHandler(allocationService)
	importHandler := handlers.NewImportHandler(importService)
	transferSuggestionHandler := handlers.NewTransferSuggestionHandler(transferLinkService, transferSuggestionRepo, transactionRepo)

	// Setup router
	router := http.NewRouter(accountHandler, categoryHandler, categoryGroupHandler, transactionHandler, allocationHandler, importHandler, transferSuggestionHandler)

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
