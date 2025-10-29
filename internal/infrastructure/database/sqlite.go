package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// NewSQLiteDB creates a new SQLite database connection
func NewSQLiteDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Initialize schema
	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

// initSchema creates all necessary tables
func initSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS accounts (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		balance INTEGER NOT NULL,
		type TEXT NOT NULL CHECK(type IN ('checking', 'savings', 'cash', 'credit')),
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS categories (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		color TEXT,
		payment_for_account_id TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (payment_for_account_id) REFERENCES accounts(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS transactions (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL DEFAULT 'normal' CHECK(type IN ('normal', 'transfer')),
		account_id TEXT NOT NULL,
		transfer_to_account_id TEXT,
		category_id TEXT,
		amount INTEGER NOT NULL,
		description TEXT,
		date DATETIME NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
		FOREIGN KEY (transfer_to_account_id) REFERENCES accounts(id) ON DELETE CASCADE,
		FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS allocations (
		id TEXT PRIMARY KEY,
		category_id TEXT NOT NULL,
		amount INTEGER NOT NULL,
		period TEXT NOT NULL,
		notes TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE,
		UNIQUE(category_id, period)
	);

	CREATE TABLE IF NOT EXISTS budget_state (
		id TEXT PRIMARY KEY,
		ready_to_assign INTEGER NOT NULL DEFAULT 0,
		updated_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions(account_id);
	CREATE INDEX IF NOT EXISTS idx_transactions_category_id ON transactions(category_id);
	CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(date);
	CREATE INDEX IF NOT EXISTS idx_allocations_period ON allocations(period);
	CREATE INDEX IF NOT EXISTS idx_allocations_category_id ON allocations(category_id);

	-- Insert default budget state if it doesn't exist
	INSERT OR IGNORE INTO budget_state (id, ready_to_assign, updated_at)
	VALUES ('singleton', 0, datetime('now'));
	`

	_, err := db.Exec(schema)
	return err
}
