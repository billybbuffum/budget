package database

import (
	"database/sql"
	"fmt"
	"sort"
	"time"
)

// Migration represents a database migration
type Migration struct {
	Version     string
	Description string
	Up          func(*sql.DB) error
	Down        func(*sql.DB) error
}

// migrations holds all registered migrations in order
var migrations = []Migration{
	{
		Version:     "001_make_category_id_nullable",
		Description: "Make category_id nullable in transactions table to support imported transactions",
		Up:          migrateCategoryIDNullable,
		Down:        rollbackCategoryIDNullable,
	},
	{
		Version:     "002_add_fitid_to_transactions",
		Description: "Add fitid column to transactions table for OFX duplicate detection",
		Up:          migrateAddFitID,
		Down:        rollbackAddFitID,
	},
	{
		Version:     "003_add_category_groups",
		Description: "Add category_groups table and group_id to categories for organizing categories into groups",
		Up:          migrateAddCategoryGroups,
		Down:        rollbackAddCategoryGroups,
	},
}

// migrateCategoryIDNullable makes the category_id column nullable in transactions table
func migrateCategoryIDNullable(db *sql.DB) error {
	// SQLite doesn't support ALTER COLUMN, so we need to:
	// 1. Create a new table with the updated schema
	// 2. Copy data from old table to new table
	// 3. Drop old table
	// 4. Rename new table to old name
	// 5. Recreate indexes

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create new transactions table with nullable category_id
	_, err = tx.Exec(`
		CREATE TABLE transactions_new (
			id TEXT PRIMARY KEY,
			account_id TEXT NOT NULL,
			category_id TEXT,
			amount INTEGER NOT NULL,
			description TEXT,
			date DATETIME NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
			FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create new transactions table: %w", err)
	}

	// Copy all data from old table to new table
	_, err = tx.Exec(`
		INSERT INTO transactions_new (id, account_id, category_id, amount, description, date, created_at, updated_at)
		SELECT id, account_id, category_id, amount, description, date, created_at, updated_at
		FROM transactions
	`)
	if err != nil {
		return fmt.Errorf("failed to copy data to new transactions table: %w", err)
	}

	// Drop old table
	_, err = tx.Exec("DROP TABLE transactions")
	if err != nil {
		return fmt.Errorf("failed to drop old transactions table: %w", err)
	}

	// Rename new table to original name
	_, err = tx.Exec("ALTER TABLE transactions_new RENAME TO transactions")
	if err != nil {
		return fmt.Errorf("failed to rename new transactions table: %w", err)
	}

	// Recreate indexes
	_, err = tx.Exec(`
		CREATE INDEX idx_transactions_account_id ON transactions(account_id);
		CREATE INDEX idx_transactions_category_id ON transactions(category_id);
		CREATE INDEX idx_transactions_date ON transactions(date);
	`)
	if err != nil {
		return fmt.Errorf("failed to recreate indexes: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// rollbackCategoryIDNullable reverts the category_id nullable migration
func rollbackCategoryIDNullable(db *sql.DB) error {
	// For rollback, we need to ensure all category_ids are non-null first
	// This is a safeguard - in practice we may not rollback this migration
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if there are any transactions with null category_id
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM transactions WHERE category_id IS NULL").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for null category_ids: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("cannot rollback: %d transactions have null category_id", count)
	}

	// Create new transactions table with NOT NULL category_id
	_, err = tx.Exec(`
		CREATE TABLE transactions_new (
			id TEXT PRIMARY KEY,
			account_id TEXT NOT NULL,
			category_id TEXT NOT NULL,
			amount INTEGER NOT NULL,
			description TEXT,
			date DATETIME NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
			FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create new transactions table: %w", err)
	}

	// Copy all data
	_, err = tx.Exec(`
		INSERT INTO transactions_new (id, account_id, category_id, amount, description, date, created_at, updated_at)
		SELECT id, account_id, category_id, amount, description, date, created_at, updated_at
		FROM transactions
	`)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// Drop old table
	_, err = tx.Exec("DROP TABLE transactions")
	if err != nil {
		return fmt.Errorf("failed to drop old table: %w", err)
	}

	// Rename new table
	_, err = tx.Exec("ALTER TABLE transactions_new RENAME TO transactions")
	if err != nil {
		return fmt.Errorf("failed to rename table: %w", err)
	}

	// Recreate indexes
	_, err = tx.Exec(`
		CREATE INDEX idx_transactions_account_id ON transactions(account_id);
		CREATE INDEX idx_transactions_category_id ON transactions(category_id);
		CREATE INDEX idx_transactions_date ON transactions(date);
	`)
	if err != nil {
		return fmt.Errorf("failed to recreate indexes: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// migrateAddFitID adds the fitid column to transactions table
func migrateAddFitID(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create new transactions table with fitid column
	_, err = tx.Exec(`
		CREATE TABLE transactions_new (
			id TEXT PRIMARY KEY,
			account_id TEXT NOT NULL,
			category_id TEXT,
			amount INTEGER NOT NULL,
			description TEXT,
			date DATETIME NOT NULL,
			fitid TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
			FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create new transactions table: %w", err)
	}

	// Copy all data from old table to new table
	_, err = tx.Exec(`
		INSERT INTO transactions_new (id, account_id, category_id, amount, description, date, fitid, created_at, updated_at)
		SELECT id, account_id, category_id, amount, description, date, NULL, created_at, updated_at
		FROM transactions
	`)
	if err != nil {
		return fmt.Errorf("failed to copy data to new transactions table: %w", err)
	}

	// Drop old table
	_, err = tx.Exec("DROP TABLE transactions")
	if err != nil {
		return fmt.Errorf("failed to drop old transactions table: %w", err)
	}

	// Rename new table to original name
	_, err = tx.Exec("ALTER TABLE transactions_new RENAME TO transactions")
	if err != nil {
		return fmt.Errorf("failed to rename new transactions table: %w", err)
	}

	// Recreate indexes and add index for fitid
	_, err = tx.Exec(`
		CREATE INDEX idx_transactions_account_id ON transactions(account_id);
		CREATE INDEX idx_transactions_category_id ON transactions(category_id);
		CREATE INDEX idx_transactions_date ON transactions(date);
		CREATE INDEX idx_transactions_fitid ON transactions(fitid);
	`)
	if err != nil {
		return fmt.Errorf("failed to recreate indexes: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// rollbackAddFitID removes the fitid column from transactions table
func rollbackAddFitID(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create new transactions table without fitid column
	_, err = tx.Exec(`
		CREATE TABLE transactions_new (
			id TEXT PRIMARY KEY,
			account_id TEXT NOT NULL,
			category_id TEXT,
			amount INTEGER NOT NULL,
			description TEXT,
			date DATETIME NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
			FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create new transactions table: %w", err)
	}

	// Copy all data from old table to new table (fitid column is dropped)
	_, err = tx.Exec(`
		INSERT INTO transactions_new (id, account_id, category_id, amount, description, date, created_at, updated_at)
		SELECT id, account_id, category_id, amount, description, date, created_at, updated_at
		FROM transactions
	`)
	if err != nil {
		return fmt.Errorf("failed to copy data to new transactions table: %w", err)
	}

	// Drop old table
	_, err = tx.Exec("DROP TABLE transactions")
	if err != nil {
		return fmt.Errorf("failed to drop old transactions table: %w", err)
	}

	// Rename new table to original name
	_, err = tx.Exec("ALTER TABLE transactions_new RENAME TO transactions")
	if err != nil {
		return fmt.Errorf("failed to rename new transactions table: %w", err)
	}

	// Recreate indexes
	_, err = tx.Exec(`
		CREATE INDEX idx_transactions_account_id ON transactions(account_id);
		CREATE INDEX idx_transactions_category_id ON transactions(category_id);
		CREATE INDEX idx_transactions_date ON transactions(date);
	`)
	if err != nil {
		return fmt.Errorf("failed to recreate indexes: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// migrateAddCategoryGroups creates the category_groups table and adds group_id to categories
func migrateAddCategoryGroups(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create category_groups table
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS category_groups (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL CHECK(type IN ('income', 'expense')),
			description TEXT,
			display_order INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create category_groups table: %w", err)
	}

	// Create new categories table with group_id column
	_, err = tx.Exec(`
		CREATE TABLE categories_new (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL CHECK(type IN ('income', 'expense')),
			description TEXT,
			color TEXT,
			group_id TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (group_id) REFERENCES category_groups(id) ON DELETE SET NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create new categories table: %w", err)
	}

	// Copy all data from old table to new table (group_id will be NULL initially)
	_, err = tx.Exec(`
		INSERT INTO categories_new (id, name, type, description, color, group_id, created_at, updated_at)
		SELECT id, name, type, description, color, NULL, created_at, updated_at
		FROM categories
	`)
	if err != nil {
		return fmt.Errorf("failed to copy data to new categories table: %w", err)
	}

	// Drop old table
	_, err = tx.Exec("DROP TABLE categories")
	if err != nil {
		return fmt.Errorf("failed to drop old categories table: %w", err)
	}

	// Rename new table to original name
	_, err = tx.Exec("ALTER TABLE categories_new RENAME TO categories")
	if err != nil {
		return fmt.Errorf("failed to rename new categories table: %w", err)
	}

	// Create index for group_id
	_, err = tx.Exec("CREATE INDEX idx_categories_group_id ON categories(group_id)")
	if err != nil {
		return fmt.Errorf("failed to create index on group_id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// rollbackAddCategoryGroups removes the category_groups table and group_id from categories
func rollbackAddCategoryGroups(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create categories table without group_id column
	_, err = tx.Exec(`
		CREATE TABLE categories_new (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL CHECK(type IN ('income', 'expense')),
			description TEXT,
			color TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create new categories table: %w", err)
	}

	// Copy all data from old table to new table (group_id column is dropped)
	_, err = tx.Exec(`
		INSERT INTO categories_new (id, name, type, description, color, created_at, updated_at)
		SELECT id, name, type, description, color, created_at, updated_at
		FROM categories
	`)
	if err != nil {
		return fmt.Errorf("failed to copy data to new categories table: %w", err)
	}

	// Drop old table
	_, err = tx.Exec("DROP TABLE categories")
	if err != nil {
		return fmt.Errorf("failed to drop old categories table: %w", err)
	}

	// Rename new table to original name
	_, err = tx.Exec("ALTER TABLE categories_new RENAME TO categories")
	if err != nil {
		return fmt.Errorf("failed to rename new categories table: %w", err)
	}

	// Drop category_groups table
	_, err = tx.Exec("DROP TABLE IF EXISTS category_groups")
	if err != nil {
		return fmt.Errorf("failed to drop category_groups table: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// initMigrationTable creates the schema_migrations table if it doesn't exist
func initMigrationTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME NOT NULL
		)
	`)
	return err
}

// getAppliedMigrations returns a set of already applied migration versions
func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// recordMigration records a migration as applied
func recordMigration(db *sql.DB, version string) error {
	_, err := db.Exec(
		"INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)",
		version,
		time.Now(),
	)
	return err
}

// RunMigrations runs all pending migrations
func RunMigrations(db *sql.DB) error {
	// Create migration tracking table
	if err := initMigrationTable(db); err != nil {
		return fmt.Errorf("failed to initialize migration table: %w", err)
	}

	// Get already applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Sort migrations by version to ensure consistent order
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	// Run pending migrations
	for _, migration := range migrations {
		if applied[migration.Version] {
			continue
		}

		fmt.Printf("Running migration: %s - %s\n", migration.Version, migration.Description)
		if err := migration.Up(db); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.Version, err)
		}

		if err := recordMigration(db, migration.Version); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
		}

		fmt.Printf("Migration %s completed successfully\n", migration.Version)
	}

	return nil
}
