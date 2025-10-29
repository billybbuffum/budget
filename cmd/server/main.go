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
	userRepo := repository.NewUserRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	budgetRepo := repository.NewBudgetRepository(db)

	// Initialize services
	userService := application.NewUserService(userRepo)
	categoryService := application.NewCategoryService(categoryRepo)
	transactionService := application.NewTransactionService(transactionRepo, userRepo, categoryRepo)
	budgetService := application.NewBudgetService(budgetRepo, categoryRepo, transactionRepo)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	budgetHandler := handlers.NewBudgetHandler(budgetService)

	// Setup router
	router := http.NewRouter(userHandler, categoryHandler, transactionHandler, budgetHandler)

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
