# OFX/QFX Transaction Import Implementation Guide

## Project Context

This is a personal budget management application built with:
- **Backend**: Go 1.23, SQLite database
- **Frontend**: Vanilla JavaScript, HTML, Tailwind CSS
- **Architecture**: Clean Architecture (Domain → Application → Infrastructure)

Current state: Users manually create transactions. We need to add the ability to import transactions from bank/credit card OFX/QFX files.

## Problem Statement

The user has accounts at 3 financial institutions (9 total accounts):
- OnPoint Community Credit Union (checking, savings, premium savings, credit card)
- JP Morgan Chase (checking, savings, 2 credit cards)
- Wells Fargo (checking)

All these institutions export OFX/QFX format, which is a standardized XML-based financial data format. We need to:
1. Parse OFX/QFX files
2. Import transactions into the database
3. Handle uncategorized transactions (current model requires category_id)
4. Prevent duplicate imports
5. Provide UI for categorizing imported transactions

## Current System Overview

### Database Schema (SQLite)
Location: `/home/user/budget/internal/infrastructure/database/sqlite.go`

```sql
CREATE TABLE transactions (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    category_id TEXT NOT NULL,  -- ⚠️ This needs to become nullable
    amount INTEGER NOT NULL,     -- In cents (positive=inflow, negative=outflow)
    description TEXT,
    date DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);
```

### Domain Model
Location: `/home/user/budget/internal/domain/transaction.go`

```go
type Transaction struct {
    ID          string    `json:"id"`
    AccountID   string    `json:"account_id"`
    CategoryID  string    `json:"category_id"`  // ⚠️ This needs to become *string
    Amount      int64     `json:"amount"`
    Description string    `json:"description"`
    Date        time.Time `json:"date"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### Repository Interface
Location: `/home/user/budget/internal/domain/repository.go`

```go
type TransactionRepository interface {
    Create(ctx context.Context, transaction *Transaction) error
    GetByID(ctx context.Context, id string) (*Transaction, error)
    List(ctx context.Context) ([]*Transaction, error)
    ListByAccount(ctx context.Context, accountID string) ([]*Transaction, error)
    ListByCategory(ctx context.Context, categoryID string) ([]*Transaction, error)
    ListByPeriod(ctx context.Context, startDate, endDate string) ([]*Transaction, error)
    GetCategoryActivity(ctx context.Context, categoryID, period string) (int64, error)
    Update(ctx context.Context, transaction *Transaction) error
    Delete(ctx context.Context, id string) error
}
```

## Implementation Requirements

### Phase 1: Schema Migration (Make category_id Nullable)

#### 1.1 Create Migration System
Since the app currently uses `initSchema()` for initialization, we need to add a migration system.

**Create**: `internal/infrastructure/database/migrations.go`

```go
package database

import (
    "database/sql"
    "fmt"
)

// Migration represents a database migration
type Migration struct {
    Version int
    Name    string
    Up      string
}

var migrations = []Migration{
    {
        Version: 1,
        Name:    "make_category_id_nullable",
        Up: `
            -- Create new table with nullable category_id
            CREATE TABLE transactions_new (
                id TEXT PRIMARY KEY,
                account_id TEXT NOT NULL,
                category_id TEXT,  -- Now nullable
                amount INTEGER NOT NULL,
                description TEXT,
                date DATETIME NOT NULL,
                created_at DATETIME NOT NULL,
                updated_at DATETIME NOT NULL,
                FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
                FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
            );

            -- Copy existing data
            INSERT INTO transactions_new
            SELECT id, account_id, category_id, amount, description, date, created_at, updated_at
            FROM transactions;

            -- Drop old table
            DROP TABLE transactions;

            -- Rename new table
            ALTER TABLE transactions_new RENAME TO transactions;

            -- Recreate indexes
            CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions(account_id);
            CREATE INDEX IF NOT EXISTS idx_transactions_category_id ON transactions(category_id);
            CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(date);
        `,
    },
}

// runMigrations applies all pending migrations
func runMigrations(db *sql.DB) error {
    // Create migrations table if it doesn't exist
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version INTEGER PRIMARY KEY,
            name TEXT NOT NULL,
            applied_at DATETIME NOT NULL
        )
    `)
    if err != nil {
        return fmt.Errorf("failed to create migrations table: %w", err)
    }

    // Get current version
    var currentVersion int
    err = db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&currentVersion)
    if err != nil {
        return fmt.Errorf("failed to get current version: %w", err)
    }

    // Apply pending migrations
    for _, migration := range migrations {
        if migration.Version <= currentVersion {
            continue
        }

        tx, err := db.Begin()
        if err != nil {
            return fmt.Errorf("failed to begin transaction: %w", err)
        }

        // Run migration
        if _, err := tx.Exec(migration.Up); err != nil {
            tx.Rollback()
            return fmt.Errorf("failed to run migration %d (%s): %w", migration.Version, migration.Name, err)
        }

        // Record migration
        _, err = tx.Exec(
            "INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, datetime('now'))",
            migration.Version, migration.Name,
        )
        if err != nil {
            tx.Rollback()
            return fmt.Errorf("failed to record migration: %w", err)
        }

        if err := tx.Commit(); err != nil {
            return fmt.Errorf("failed to commit migration: %w", err)
        }

        fmt.Printf("Applied migration %d: %s\n", migration.Version, migration.Name)
    }

    return nil
}
```

#### 1.2 Update sqlite.go to Use Migrations

**Update**: `internal/infrastructure/database/sqlite.go`

Add this call after `initSchema()`:

```go
func NewSQLiteDB(dbPath string) (*sql.DB, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
        return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
    }

    if err := initSchema(db); err != nil {
        return nil, fmt.Errorf("failed to initialize schema: %w", err)
    }

    // Run migrations
    if err := runMigrations(db); err != nil {
        return nil, fmt.Errorf("failed to run migrations: %w", err)
    }

    return db, nil
}
```

### Phase 2: Update Domain Model

#### 2.1 Update Transaction Model

**Update**: `internal/domain/transaction.go`

```go
package domain

import "time"

// Transaction represents a single income or expense transaction
// Positive amounts = Inflows (money coming in)
// Negative amounts = Outflows (money going out/expenses)
type Transaction struct {
    ID          string    `json:"id"`
    AccountID   string    `json:"account_id"`
    CategoryID  *string   `json:"category_id,omitempty"`  // Nullable - nil means uncategorized
    Amount      int64     `json:"amount"`
    Description string    `json:"description"`
    Date        time.Time `json:"date"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// IsUncategorized returns true if the transaction has no category assigned
func (t *Transaction) IsUncategorized() bool {
    return t.CategoryID == nil
}

// SetCategory assigns a category to the transaction
func (t *Transaction) SetCategory(categoryID string) {
    t.CategoryID = &categoryID
}

// ClearCategory removes the category assignment
func (t *Transaction) ClearCategory() {
    t.CategoryID = nil
}
```

#### 2.2 Update Repository Interface

**Update**: `internal/domain/repository.go`

Add new methods to the `TransactionRepository` interface:

```go
type TransactionRepository interface {
    // Existing methods...
    Create(ctx context.Context, transaction *Transaction) error
    GetByID(ctx context.Context, id string) (*Transaction, error)
    List(ctx context.Context) ([]*Transaction, error)
    ListByAccount(ctx context.Context, accountID string) ([]*Transaction, error)
    ListByCategory(ctx context.Context, categoryID string) ([]*Transaction, error)
    ListByPeriod(ctx context.Context, startDate, endDate string) ([]*Transaction, error)
    GetCategoryActivity(ctx context.Context, categoryID, period string) (int64, error)
    Update(ctx context.Context, transaction *Transaction) error
    Delete(ctx context.Context, id string) error

    // New methods for import functionality
    ListUncategorized(ctx context.Context) ([]*Transaction, error)
    FindByAccountAndDate(ctx context.Context, accountID string, date time.Time, amount int64, description string) (*Transaction, error)
    BulkCreate(ctx context.Context, transactions []*Transaction) error
}
```

#### 2.3 Update Repository Implementation

**Update**: `internal/infrastructure/repository/transaction_repository.go`

Update all SQL queries to handle nullable category_id and add new methods:

```go
package repository

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    "github.com/billybbuffum/budget/internal/domain"
)

type transactionRepository struct {
    db *sql.DB
}

func NewTransactionRepository(db *sql.DB) domain.TransactionRepository {
    return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(ctx context.Context, transaction *domain.Transaction) error {
    query := `
        INSERT INTO transactions (id, account_id, category_id, amount, description, date, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `
    _, err := r.db.ExecContext(ctx, query,
        transaction.ID, transaction.AccountID, transaction.CategoryID,
        transaction.Amount, transaction.Description, transaction.Date,
        transaction.CreatedAt, transaction.UpdatedAt)
    if err != nil {
        return fmt.Errorf("failed to create transaction: %w", err)
    }
    return nil
}

func (r *transactionRepository) GetByID(ctx context.Context, id string) (*domain.Transaction, error) {
    query := `
        SELECT id, account_id, category_id, amount, description, date, created_at, updated_at
        FROM transactions
        WHERE id = ?
    `
    transaction := &domain.Transaction{}
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &transaction.ID, &transaction.AccountID, &transaction.CategoryID,
        &transaction.Amount, &transaction.Description, &transaction.Date,
        &transaction.CreatedAt, &transaction.UpdatedAt)
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("transaction not found")
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get transaction: %w", err)
    }
    return transaction, nil
}

func (r *transactionRepository) List(ctx context.Context) ([]*domain.Transaction, error) {
    query := `
        SELECT id, account_id, category_id, amount, description, date, created_at, updated_at
        FROM transactions
        ORDER BY date DESC
    `
    rows, err := r.db.QueryContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to list transactions: %w", err)
    }
    defer rows.Close()

    return r.scanTransactions(rows)
}

func (r *transactionRepository) ListByAccount(ctx context.Context, accountID string) ([]*domain.Transaction, error) {
    query := `
        SELECT id, account_id, category_id, amount, description, date, created_at, updated_at
        FROM transactions
        WHERE account_id = ?
        ORDER BY date DESC
    `
    rows, err := r.db.QueryContext(ctx, query, accountID)
    if err != nil {
        return nil, fmt.Errorf("failed to list transactions by account: %w", err)
    }
    defer rows.Close()

    return r.scanTransactions(rows)
}

func (r *transactionRepository) ListByCategory(ctx context.Context, categoryID string) ([]*domain.Transaction, error) {
    query := `
        SELECT id, account_id, category_id, amount, description, date, created_at, updated_at
        FROM transactions
        WHERE category_id = ?
        ORDER BY date DESC
    `
    rows, err := r.db.QueryContext(ctx, query, categoryID)
    if err != nil {
        return nil, fmt.Errorf("failed to list transactions by category: %w", err)
    }
    defer rows.Close()

    return r.scanTransactions(rows)
}

func (r *transactionRepository) ListByPeriod(ctx context.Context, startDate, endDate string) ([]*domain.Transaction, error) {
    query := `
        SELECT id, account_id, category_id, amount, description, date, created_at, updated_at
        FROM transactions
        WHERE date >= ? AND date <= ?
        ORDER BY date DESC
    `
    rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
    if err != nil {
        return nil, fmt.Errorf("failed to list transactions by period: %w", err)
    }
    defer rows.Close()

    return r.scanTransactions(rows)
}

func (r *transactionRepository) GetCategoryActivity(ctx context.Context, categoryID, period string) (int64, error) {
    t, err := time.Parse("2006-01", period)
    if err != nil {
        return 0, fmt.Errorf("invalid period format: %w", err)
    }

    t = t.UTC()
    startDate := t.Add(-time.Second).Format(time.RFC3339)
    endDate := t.AddDate(0, 1, 0).Add(-time.Second).Format(time.RFC3339)

    query := `
        SELECT COALESCE(SUM(amount), 0)
        FROM transactions
        WHERE category_id = ? AND date >= ? AND date <= ?
    `
    var activity int64
    err = r.db.QueryRowContext(ctx, query, categoryID, startDate, endDate).Scan(&activity)
    if err != nil {
        return 0, fmt.Errorf("failed to get category activity: %w", err)
    }
    return activity, nil
}

func (r *transactionRepository) Update(ctx context.Context, transaction *domain.Transaction) error {
    query := `
        UPDATE transactions
        SET account_id = ?, category_id = ?, amount = ?, description = ?, date = ?, updated_at = ?
        WHERE id = ?
    `
    result, err := r.db.ExecContext(ctx, query,
        transaction.AccountID, transaction.CategoryID, transaction.Amount,
        transaction.Description, transaction.Date, transaction.UpdatedAt, transaction.ID)
    if err != nil {
        return fmt.Errorf("failed to update transaction: %w", err)
    }
    rows, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    if rows == 0 {
        return fmt.Errorf("transaction not found")
    }
    return nil
}

func (r *transactionRepository) Delete(ctx context.Context, id string) error {
    query := `DELETE FROM transactions WHERE id = ?`
    result, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return fmt.Errorf("failed to delete transaction: %w", err)
    }
    rows, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    if rows == 0 {
        return fmt.Errorf("transaction not found")
    }
    return nil
}

// ListUncategorized returns all transactions without a category
func (r *transactionRepository) ListUncategorized(ctx context.Context) ([]*domain.Transaction, error) {
    query := `
        SELECT id, account_id, category_id, amount, description, date, created_at, updated_at
        FROM transactions
        WHERE category_id IS NULL
        ORDER BY date DESC
    `
    rows, err := r.db.QueryContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to list uncategorized transactions: %w", err)
    }
    defer rows.Close()

    return r.scanTransactions(rows)
}

// FindByAccountAndDate finds a transaction by account, date, amount, and description
// Used for duplicate detection during import
func (r *transactionRepository) FindByAccountAndDate(ctx context.Context, accountID string, date time.Time, amount int64, description string) (*domain.Transaction, error) {
    query := `
        SELECT id, account_id, category_id, amount, description, date, created_at, updated_at
        FROM transactions
        WHERE account_id = ?
          AND date = ?
          AND amount = ?
          AND description = ?
        LIMIT 1
    `
    transaction := &domain.Transaction{}
    err := r.db.QueryRowContext(ctx, query, accountID, date, amount, description).Scan(
        &transaction.ID, &transaction.AccountID, &transaction.CategoryID,
        &transaction.Amount, &transaction.Description, &transaction.Date,
        &transaction.CreatedAt, &transaction.UpdatedAt)
    if err == sql.ErrNoRows {
        return nil, nil // Not found is not an error
    }
    if err != nil {
        return nil, fmt.Errorf("failed to find transaction: %w", err)
    }
    return transaction, nil
}

// BulkCreate inserts multiple transactions in a single transaction
func (r *transactionRepository) BulkCreate(ctx context.Context, transactions []*domain.Transaction) error {
    if len(transactions) == 0 {
        return nil
    }

    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO transactions (id, account_id, category_id, amount, description, date, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `)
    if err != nil {
        return fmt.Errorf("failed to prepare statement: %w", err)
    }
    defer stmt.Close()

    for _, transaction := range transactions {
        _, err := stmt.ExecContext(ctx,
            transaction.ID, transaction.AccountID, transaction.CategoryID,
            transaction.Amount, transaction.Description, transaction.Date,
            transaction.CreatedAt, transaction.UpdatedAt)
        if err != nil {
            return fmt.Errorf("failed to insert transaction %s: %w", transaction.ID, err)
        }
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}

func (r *transactionRepository) scanTransactions(rows *sql.Rows) ([]*domain.Transaction, error) {
    var transactions []*domain.Transaction
    for rows.Next() {
        transaction := &domain.Transaction{}
        if err := rows.Scan(&transaction.ID, &transaction.AccountID, &transaction.CategoryID,
            &transaction.Amount, &transaction.Description, &transaction.Date,
            &transaction.CreatedAt, &transaction.UpdatedAt); err != nil {
            return nil, fmt.Errorf("failed to scan transaction: %w", err)
        }
        transactions = append(transactions, transaction)
    }
    return transactions, nil
}
```

### Phase 3: OFX Parser Implementation

#### 3.1 Add Dependency

**Update**: `go.mod`

```bash
go get github.com/aclindsa/ofxgo
```

#### 3.2 Create OFX Parser

**Create**: `internal/infrastructure/importer/ofx_parser.go`

```go
package importer

import (
    "fmt"
    "io"
    "time"

    "github.com/aclindsa/ofxgo"
)

// OFXTransaction represents a parsed transaction from an OFX file
type OFXTransaction struct {
    FitID       string    // Financial Institution Transaction ID
    Type        string    // DEBIT, CREDIT, etc.
    Date        time.Time
    Amount      float64
    Name        string    // Payee/merchant name
    Memo        string    // Additional description
}

// OFXParser parses OFX/QFX files
type OFXParser struct{}

// NewOFXParser creates a new OFX parser
func NewOFXParser() *OFXParser {
    return &OFXParser{}
}

// Parse reads an OFX file and extracts transactions
func (p *OFXParser) Parse(reader io.Reader) ([]OFXTransaction, error) {
    // Parse OFX file
    response, err := ofxgo.ParseResponse(reader)
    if err != nil {
        return nil, fmt.Errorf("failed to parse OFX file: %w", err)
    }

    var transactions []OFXTransaction

    // Extract bank transactions
    if len(response.Bank) > 0 {
        for _, stmt := range response.Bank {
            if stmt.BankTranList != nil {
                for _, txn := range stmt.BankTranList.Transactions {
                    transactions = append(transactions, p.convertBankTransaction(txn))
                }
            }
        }
    }

    // Extract credit card transactions
    if len(response.CreditCard) > 0 {
        for _, stmt := range response.CreditCard {
            if stmt.BankTranList != nil {
                for _, txn := range stmt.BankTranList.Transactions {
                    transactions = append(transactions, p.convertBankTransaction(txn))
                }
            }
        }
    }

    return transactions, nil
}

// convertBankTransaction converts an OFX transaction to our internal format
func (p *OFXParser) convertBankTransaction(txn ofxgo.Transaction) OFXTransaction {
    return OFXTransaction{
        FitID:  txn.FiTID.String(),
        Type:   txn.TrnType.String(),
        Date:   txn.DtPosted.Time,
        Amount: txn.TrnAmt.Float(),
        Name:   txn.Name.String(),
        Memo:   txn.Memo.String(),
    }
}
```

#### 3.3 Create Import Service

**Create**: `internal/application/import_service.go`

```go
package application

import (
    "context"
    "fmt"
    "io"
    "math"
    "time"

    "github.com/billybbuffum/budget/internal/domain"
    "github.com/billybbuffum/budget/internal/infrastructure/importer"
    "github.com/google/uuid"
)

// ImportService handles importing transactions from external files
type ImportService struct {
    transactionRepo domain.TransactionRepository
    accountRepo     domain.AccountRepository
    ofxParser       *importer.OFXParser
}

// NewImportService creates a new import service
func NewImportService(transactionRepo domain.TransactionRepository, accountRepo domain.AccountRepository) *ImportService {
    return &ImportService{
        transactionRepo: transactionRepo,
        accountRepo:     accountRepo,
        ofxParser:       importer.NewOFXParser(),
    }
}

// ImportResult contains the results of an import operation
type ImportResult struct {
    TotalTransactions int      `json:"total_transactions"`
    ImportedCount     int      `json:"imported_count"`
    SkippedCount      int      `json:"skipped_count"`
    ErrorCount        int      `json:"error_count"`
    Errors            []string `json:"errors,omitempty"`
}

// ImportOFX imports transactions from an OFX/QFX file
func (s *ImportService) ImportOFX(ctx context.Context, accountID string, reader io.Reader) (*ImportResult, error) {
    // Verify account exists
    account, err := s.accountRepo.GetByID(ctx, accountID)
    if err != nil {
        return nil, fmt.Errorf("account not found: %w", err)
    }

    // Parse OFX file
    ofxTransactions, err := s.ofxParser.Parse(reader)
    if err != nil {
        return nil, fmt.Errorf("failed to parse OFX file: %w", err)
    }

    result := &ImportResult{
        TotalTransactions: len(ofxTransactions),
        Errors:            []string{},
    }

    now := time.Now()
    var transactionsToImport []*domain.Transaction

    for _, ofxTxn := range ofxTransactions {
        // Convert amount from dollars to cents
        amountCents := int64(math.Round(ofxTxn.Amount * 100))

        // Build description from available fields
        description := ofxTxn.Name
        if ofxTxn.Memo != "" && ofxTxn.Memo != ofxTxn.Name {
            description = fmt.Sprintf("%s - %s", ofxTxn.Name, ofxTxn.Memo)
        }

        // Check for duplicate (same account, date, amount, description)
        existing, err := s.transactionRepo.FindByAccountAndDate(ctx, accountID, ofxTxn.Date, amountCents, description)
        if err != nil {
            result.ErrorCount++
            result.Errors = append(result.Errors, fmt.Sprintf("Error checking duplicate for %s: %v", description, err))
            continue
        }

        if existing != nil {
            result.SkippedCount++
            continue
        }

        // Create transaction (without category - will be categorized later)
        transaction := &domain.Transaction{
            ID:          uuid.New().String(),
            AccountID:   accountID,
            CategoryID:  nil, // Uncategorized
            Amount:      amountCents,
            Description: description,
            Date:        ofxTxn.Date,
            CreatedAt:   now,
            UpdatedAt:   now,
        }

        transactionsToImport = append(transactionsToImport, transaction)
    }

    // Bulk insert transactions
    if len(transactionsToImport) > 0 {
        if err := s.transactionRepo.BulkCreate(ctx, transactionsToImport); err != nil {
            return nil, fmt.Errorf("failed to import transactions: %w", err)
        }
        result.ImportedCount = len(transactionsToImport)

        // Update account balance
        var totalChange int64
        for _, txn := range transactionsToImport {
            totalChange += txn.Amount
        }
        account.Balance += totalChange
        account.UpdatedAt = now
        if err := s.accountRepo.Update(ctx, account); err != nil {
            return nil, fmt.Errorf("failed to update account balance: %w", err)
        }
    }

    return result, nil
}
```

### Phase 4: HTTP Handlers

#### 4.1 Create Import Handler

**Create**: `internal/infrastructure/http/handlers/import_handler.go`

```go
package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/billybbuffum/budget/internal/application"
)

type ImportHandler struct {
    importService *application.ImportService
}

func NewImportHandler(importService *application.ImportService) *ImportHandler {
    return &ImportHandler{importService: importService}
}

// ImportOFXRequest represents the form data for OFX import
type ImportOFXRequest struct {
    AccountID string `json:"account_id"`
}

// HandleImportOFX handles POST /api/transactions/import
func (h *ImportHandler) HandleImportOFX(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Parse multipart form (max 10MB file)
    if err := r.ParseMultipartForm(10 << 20); err != nil {
        http.Error(w, "Failed to parse form", http.StatusBadRequest)
        return
    }

    // Get account ID
    accountID := r.FormValue("account_id")
    if accountID == "" {
        http.Error(w, "account_id is required", http.StatusBadRequest)
        return
    }

    // Get uploaded file
    file, header, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Failed to get file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Validate file extension
    if header.Filename[len(header.Filename)-4:] != ".ofx" &&
       header.Filename[len(header.Filename)-4:] != ".qfx" &&
       header.Filename[len(header.Filename)-4:] != ".OFX" &&
       header.Filename[len(header.Filename)-4:] != ".QFX" {
        http.Error(w, "Only .ofx and .qfx files are supported", http.StatusBadRequest)
        return
    }

    // Import transactions
    result, err := h.importService.ImportOFX(r.Context(), accountID, file)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

#### 4.2 Update Transaction Handler

**Update**: `internal/infrastructure/http/handlers/transaction_handler.go`

Add methods to list uncategorized transactions and bulk categorize:

```go
// Add to existing TransactionHandler

// HandleListUncategorized handles GET /api/transactions?uncategorized=true
func (h *TransactionHandler) HandleListUncategorized(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    transactions, err := h.service.ListUncategorized(r.Context())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(transactions)
}

// BulkCategorizeRequest represents the request to bulk categorize transactions
type BulkCategorizeRequest struct {
    TransactionIDs []string `json:"transaction_ids"`
    CategoryID     string   `json:"category_id"`
}

// HandleBulkCategorize handles PUT /api/transactions/bulk-categorize
func (h *TransactionHandler) HandleBulkCategorize(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPut {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req BulkCategorizeRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    if len(req.TransactionIDs) == 0 {
        http.Error(w, "transaction_ids is required", http.StatusBadRequest)
        return
    }

    if req.CategoryID == "" {
        http.Error(w, "category_id is required", http.StatusBadRequest)
        return
    }

    // Update each transaction
    for _, txnID := range req.TransactionIDs {
        if err := h.service.UpdateCategory(r.Context(), txnID, req.CategoryID); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    }

    w.WriteHeader(http.StatusNoContent)
}
```

#### 4.3 Add Service Method

**Update**: `internal/application/transaction_service.go`

```go
// Add to TransactionService

// ListUncategorized returns all transactions without a category
func (s *TransactionService) ListUncategorized(ctx context.Context) ([]*domain.Transaction, error) {
    return s.repo.ListUncategorized(ctx)
}

// UpdateCategory updates only the category of a transaction
func (s *TransactionService) UpdateCategory(ctx context.Context, id, categoryID string) error {
    transaction, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return err
    }

    transaction.SetCategory(categoryID)
    transaction.UpdatedAt = time.Now()

    return s.repo.Update(ctx, transaction)
}
```

#### 4.4 Register Routes

**Update**: `internal/infrastructure/http/router.go`

```go
// Add import handler initialization in NewRouter()

importService := application.NewImportService(transactionRepo, accountRepo)
importHandler := handlers.NewImportHandler(importService)

// Register routes
http.HandleFunc("/api/transactions/import", importHandler.HandleImportOFX)
http.HandleFunc("/api/transactions/uncategorized", transactionHandler.HandleListUncategorized)
http.HandleFunc("/api/transactions/bulk-categorize", transactionHandler.HandleBulkCategorize)
```

### Phase 5: Frontend Implementation

#### 5.1 Add Import UI

**Update**: `static/index.html`

Add this section after the transactions section:

```html
<!-- Import Transactions Section -->
<div class="bg-white rounded-lg shadow p-6">
    <h2 class="text-xl font-semibold mb-4">Import Transactions</h2>
    <form id="import-form" class="space-y-4">
        <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">Select Account</label>
            <select id="import-account-select" class="w-full p-2 border rounded focus:ring-2 focus:ring-blue-500" required>
                <option value="">Choose an account...</option>
            </select>
        </div>
        <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">OFX/QFX File</label>
            <input type="file" id="import-file" accept=".ofx,.qfx"
                   class="w-full p-2 border rounded focus:ring-2 focus:ring-blue-500" required>
        </div>
        <button type="submit" class="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">
            Import Transactions
        </button>
    </form>
    <div id="import-result" class="mt-4 hidden"></div>
</div>

<!-- Uncategorized Transactions Section -->
<div class="bg-white rounded-lg shadow p-6">
    <div class="flex justify-between items-center mb-4">
        <h2 class="text-xl font-semibold">Uncategorized Transactions</h2>
        <button id="refresh-uncategorized-btn" class="text-blue-600 hover:text-blue-700">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
        </button>
    </div>
    <div id="uncategorized-list" class="space-y-2"></div>
    <div id="uncategorized-empty" class="text-gray-500 text-center py-8 hidden">
        No uncategorized transactions
    </div>
</div>
```

#### 5.2 Add JavaScript Functions

**Update**: `static/app.js`

Add these functions:

```javascript
// Import transactions
async function importTransactions(accountId, file) {
    const formData = new FormData();
    formData.append('account_id', accountId);
    formData.append('file', file);

    const response = await fetch('/api/transactions/import', {
        method: 'POST',
        body: formData
    });

    if (!response.ok) {
        const error = await response.text();
        throw new Error(error);
    }

    return await response.json();
}

// Get uncategorized transactions
async function getUncategorizedTransactions() {
    const response = await fetch('/api/transactions/uncategorized');
    if (!response.ok) throw new Error('Failed to fetch uncategorized transactions');
    return await response.json();
}

// Bulk categorize transactions
async function bulkCategorizeTransactions(transactionIds, categoryId) {
    const response = await fetch('/api/transactions/bulk-categorize', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            transaction_ids: transactionIds,
            category_id: categoryId
        })
    });

    if (!response.ok) throw new Error('Failed to categorize transactions');
}

// Render uncategorized transactions
function renderUncategorizedTransactions(transactions) {
    const container = document.getElementById('uncategorized-list');
    const emptyState = document.getElementById('uncategorized-empty');

    if (transactions.length === 0) {
        container.classList.add('hidden');
        emptyState.classList.remove('hidden');
        return;
    }

    container.classList.remove('hidden');
    emptyState.classList.add('hidden');

    container.innerHTML = transactions.map(txn => {
        const account = state.accounts.find(a => a.id === txn.account_id);
        const amount = (txn.amount / 100).toFixed(2);
        const amountClass = txn.amount >= 0 ? 'text-green-600' : 'text-red-600';

        return `
            <div class="border rounded p-4 flex items-center justify-between">
                <div class="flex-1">
                    <div class="font-medium">${txn.description}</div>
                    <div class="text-sm text-gray-500">
                        ${account?.name || 'Unknown'} • ${new Date(txn.date).toLocaleDateString()}
                    </div>
                </div>
                <div class="flex items-center gap-4">
                    <div class="${amountClass} font-semibold">
                        $${amount}
                    </div>
                    <select class="categorize-select border rounded p-2" data-transaction-id="${txn.id}">
                        <option value="">Choose category...</option>
                        ${state.categories.map(cat =>
                            `<option value="${cat.id}">${cat.name}</option>`
                        ).join('')}
                    </select>
                </div>
            </div>
        `;
    }).join('');

    // Add event listeners for category selects
    container.querySelectorAll('.categorize-select').forEach(select => {
        select.addEventListener('change', async (e) => {
            const transactionId = e.target.dataset.transactionId;
            const categoryId = e.target.value;

            if (categoryId) {
                try {
                    await bulkCategorizeTransactions([transactionId], categoryId);
                    await loadUncategorizedTransactions();
                    await loadTransactions();
                } catch (error) {
                    alert('Failed to categorize transaction: ' + error.message);
                }
            }
        });
    });
}

// Load uncategorized transactions
async function loadUncategorizedTransactions() {
    try {
        const transactions = await getUncategorizedTransactions();
        renderUncategorizedTransactions(transactions);
    } catch (error) {
        console.error('Failed to load uncategorized transactions:', error);
    }
}

// Initialize import functionality
function initializeImport() {
    const form = document.getElementById('import-form');
    const accountSelect = document.getElementById('import-account-select');
    const resultDiv = document.getElementById('import-result');

    // Populate account select
    state.accounts.forEach(account => {
        const option = document.createElement('option');
        option.value = account.id;
        option.textContent = account.name;
        accountSelect.appendChild(option);
    });

    // Handle form submission
    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        const accountId = accountSelect.value;
        const fileInput = document.getElementById('import-file');
        const file = fileInput.files[0];

        if (!file) {
            alert('Please select a file');
            return;
        }

        try {
            resultDiv.classList.add('hidden');
            const result = await importTransactions(accountId, file);

            // Show results
            resultDiv.innerHTML = `
                <div class="p-4 bg-blue-50 border border-blue-200 rounded">
                    <h3 class="font-semibold mb-2">Import Complete</h3>
                    <ul class="text-sm space-y-1">
                        <li>Total transactions: ${result.total_transactions}</li>
                        <li>Imported: ${result.imported_count}</li>
                        <li>Skipped (duplicates): ${result.skipped_count}</li>
                        ${result.error_count > 0 ? `<li class="text-red-600">Errors: ${result.error_count}</li>` : ''}
                    </ul>
                </div>
            `;
            resultDiv.classList.remove('hidden');

            // Refresh data
            await loadAccounts();
            await loadTransactions();
            await loadUncategorizedTransactions();

            // Reset form
            form.reset();
        } catch (error) {
            resultDiv.innerHTML = `
                <div class="p-4 bg-red-50 border border-red-200 rounded text-red-700">
                    <strong>Import Failed:</strong> ${error.message}
                </div>
            `;
            resultDiv.classList.remove('hidden');
        }
    });

    // Refresh button
    document.getElementById('refresh-uncategorized-btn').addEventListener('click', loadUncategorizedTransactions);
}

// Update initialization
async function init() {
    await loadAccounts();
    await loadCategories();
    await loadTransactions();
    await loadAllocations();
    await loadUncategorizedTransactions();

    initializeImport();
    // ... rest of existing initialization
}

// Call init on page load
document.addEventListener('DOMContentLoaded', init);
```

## Testing Requirements

### Manual Testing Checklist

1. **Schema Migration**
   - [ ] Start app with existing data
   - [ ] Verify migration runs successfully
   - [ ] Verify existing transactions still load correctly
   - [ ] Verify can create transactions with and without categories

2. **OFX Import**
   - [ ] Download OFX file from each bank (OnPoint, Chase, Wells Fargo)
   - [ ] Import each file through the UI
   - [ ] Verify correct number of transactions imported
   - [ ] Verify duplicate detection works (re-import same file)
   - [ ] Verify account balances update correctly
   - [ ] Verify transactions appear as uncategorized

3. **Categorization**
   - [ ] List uncategorized transactions
   - [ ] Assign category to single transaction
   - [ ] Verify transaction moves out of uncategorized list
   - [ ] Test bulk categorization (future feature)

4. **Error Handling**
   - [ ] Upload invalid file (not OFX)
   - [ ] Upload to non-existent account
   - [ ] Upload corrupted OFX file
   - [ ] Verify error messages are helpful

### Unit Tests (Optional but Recommended)

Create test files:
- `internal/infrastructure/importer/ofx_parser_test.go`
- `internal/application/import_service_test.go`

Example test for OFX parser:

```go
package importer

import (
    "strings"
    "testing"
)

func TestOFXParser_Parse(t *testing.T) {
    ofxData := `
    <?xml version="1.0"?>
    <OFX>
        <SIGNONMSGSRSV1>...</SIGNONMSGSRSV1>
        <BANKMSGSRSV1>
            <STMTTRNRS>
                <STMTRS>
                    <BANKTRANLIST>
                        <STMTTRN>
                            <TRNTYPE>DEBIT</TRNTYPE>
                            <DTPOSTED>20250115</DTPOSTED>
                            <TRNAMT>-45.23</TRNAMT>
                            <FITID>12345</FITID>
                            <NAME>SAFEWAY</NAME>
                        </STMTTRN>
                    </BANKTRANLIST>
                </STMTRS>
            </STMTTRNRS>
        </BANKMSGSRSV1>
    </OFX>
    `

    parser := NewOFXParser()
    transactions, err := parser.Parse(strings.NewReader(ofxData))

    if err != nil {
        t.Fatalf("Parse failed: %v", err)
    }

    if len(transactions) != 1 {
        t.Fatalf("Expected 1 transaction, got %d", len(transactions))
    }

    txn := transactions[0]
    if txn.Amount != -45.23 {
        t.Errorf("Expected amount -45.23, got %f", txn.Amount)
    }
}
```

## Acceptance Criteria

### Must Have
- ✅ Schema migration runs automatically on startup
- ✅ CategoryID is nullable in database and domain model
- ✅ Can import OFX/QFX files from all 3 banks (OnPoint, Chase, Wells Fargo)
- ✅ Duplicate detection prevents re-importing same transactions
- ✅ Account balances update correctly after import
- ✅ Can list uncategorized transactions
- ✅ Can assign categories to individual transactions
- ✅ UI shows import results (imported, skipped, errors)

### Should Have
- ✅ File upload validation (OFX/QFX only)
- ✅ Clear error messages for import failures
- ✅ Transaction description combines Name and Memo from OFX
- ✅ Import results show helpful summary

### Nice to Have (Future Enhancements)
- Bulk categorization UI (select multiple, assign one category)
- Auto-categorization rules (e.g., "SAFEWAY" → "Groceries")
- Import history tracking
- CSV import support
- Transaction matching/reconciliation

## Common Pitfalls to Avoid

1. **Float Precision**: Always convert dollars to cents (int64) immediately after parsing
2. **Timezone Handling**: OFX dates may be in different timezones; normalize to UTC
3. **NULL Handling**: Go's `sql.NullString` vs `*string` - we chose `*string` for JSON compatibility
4. **File Size**: Limit upload size to 10MB to prevent DoS
5. **Transaction Ordering**: Sort by date DESC for better UX
6. **Foreign Keys**: Remember category_id can now be NULL, so ON DELETE CASCADE still applies only when set

## Resources

- OFX Specification: https://www.ofx.net/
- ofxgo Library: https://github.com/aclindsa/ofxgo
- Go SQLite Driver: https://github.com/mattn/go-sqlite3
- UUID Library: https://github.com/google/uuid

## Questions?

If you encounter issues or need clarification:
1. Check the OFX file structure (open in text editor)
2. Verify database schema matches expected structure
3. Test with small OFX sample before full import
4. Check server logs for detailed error messages

Good luck with the implementation!
