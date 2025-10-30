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
		Version:     "003_deprecate_ready_to_assign",
		Description: "Deprecate ready_to_assign field - now calculated per period instead of global singleton",
		Up:          migrateDeprecateReadyToAssign,
		Down:        rollbackDeprecateReadyToAssign,
	},
	{
		Version:     "004_add_credit_card_support",
		Description: "Add type and transfer_to_account_id columns for credit card and transfer support",
		Up:          migrateAddCreditCardSupport,
		Down:        rollbackAddCreditCardSupport,
	},
	{
		Version:     "005_add_category_groups",
		Description: "Add category_groups table and group_id to categories for organizing categories into groups",
		Up:          migrateAddCategoryGroups,
		Down:        rollbackAddCategoryGroups,
	},
	{
		Version:     "006_simplify_category_groups",
		Description: "Remove type field from category_groups - groups are for budget organization only",
		Up:          migrateSimplifyGroups,
		Down:        rollbackSimplifyGroups,
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

// migrateDeprecateReadyToAssign sets ready_to_assign to 0 as it's now calculated per-period
func migrateDeprecateReadyToAssign(db *sql.DB) error {
	// Just reset the value to 0 - the field will remain for backward compatibility
	// but won't be used. Ready to Assign is now calculated per period.
	_, err := db.Exec(`
		UPDATE budget_state
		SET ready_to_assign = 0, updated_at = datetime('now')
		WHERE id = 'singleton'
	`)
	return err
}

// rollbackDeprecateReadyToAssign would need to recalculate ready_to_assign from data
func rollbackDeprecateReadyToAssign(db *sql.DB) error {
	// Recalculate ready_to_assign as: Total Account Balance - Total Allocated + Total Spent
	_, err := db.Exec(`
		UPDATE budget_state
		SET ready_to_assign = (
			SELECT COALESCE(SUM(balance), 0) FROM accounts
		) - (
			SELECT COALESCE(
				(SELECT COALESCE(SUM(amount), 0) FROM allocations) -
				(SELECT COALESCE(SUM(ABS(amount)), 0) FROM transactions WHERE amount < 0),
			0)
		),
		updated_at = datetime('now')
		WHERE id = 'singleton'
	`)
	return err
}

// migrateAddCategoryGroups creates the category_groups table and adds group_id to categories
// Note: This migration uses ALTER TABLE pattern like migration 004
func migrateAddCategoryGroups(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Step 1: Create category_groups table
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

	// Step 2: Add group_id column to categories (if it doesn't exist)
	var columnExists int
	err = tx.QueryRow("SELECT COUNT(*) FROM pragma_table_info('categories') WHERE name='group_id'").Scan(&columnExists)
	if err != nil {
		return fmt.Errorf("failed to check for group_id column: %w", err)
	}

	if columnExists == 0 {
		_, err = tx.Exec("ALTER TABLE categories ADD COLUMN group_id TEXT REFERENCES category_groups(id) ON DELETE SET NULL")
		if err != nil {
			return fmt.Errorf("failed to add group_id column: %w", err)
		}

		// Create index for group_id
		_, err = tx.Exec("CREATE INDEX IF NOT EXISTS idx_categories_group_id ON categories(group_id)")
		if err != nil {
			return fmt.Errorf("failed to create index on group_id: %w", err)
		}
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

	// Check if group_id column exists
	var columnExists int
	err = tx.QueryRow("SELECT COUNT(*) FROM pragma_table_info('categories') WHERE name='group_id'").Scan(&columnExists)
	if err != nil {
		return fmt.Errorf("failed to check for group_id column: %w", err)
	}

	// If group_id exists, we need to recreate the table without it
	// (SQLite doesn't support DROP COLUMN until version 3.35.0)
	if columnExists > 0 {
		// Create categories table without group_id column
		_, err = tx.Exec(`
			CREATE TABLE categories_new (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				description TEXT,
				color TEXT,
				payment_for_account_id TEXT,
				created_at DATETIME NOT NULL,
				updated_at DATETIME NOT NULL,
				FOREIGN KEY (payment_for_account_id) REFERENCES accounts(id) ON DELETE SET NULL
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create new categories table: %w", err)
		}

		// Copy all data from old table to new table (group_id column is dropped)
		_, err = tx.Exec(`
			INSERT INTO categories_new (id, name, description, color, payment_for_account_id, created_at, updated_at)
			SELECT id, name, description, color, payment_for_account_id, created_at, updated_at
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

// migrateAddCreditCardSupport adds type and transfer_to_account_id columns
func migrateAddCreditCardSupport(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Step 1: Add payment_for_account_id column to categories (if it doesn't exist)
	// Check if column exists first
	var columnExists int
	err = tx.QueryRow("SELECT COUNT(*) FROM pragma_table_info('categories') WHERE name='payment_for_account_id'").Scan(&columnExists)
	if err != nil {
		return fmt.Errorf("failed to check if payment_for_account_id exists: %w", err)
	}

	if columnExists == 0 {
		_, err = tx.Exec("ALTER TABLE categories ADD COLUMN payment_for_account_id TEXT REFERENCES accounts(id) ON DELETE CASCADE")
		if err != nil {
			return fmt.Errorf("failed to add payment_for_account_id column: %w", err)
		}
	}

	// Step 2: Create new transactions table with type and transfer support
	_, err = tx.Exec(`
		CREATE TABLE transactions_new (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL DEFAULT 'normal' CHECK(type IN ('normal', 'transfer')),
			account_id TEXT NOT NULL,
			transfer_to_account_id TEXT,
			category_id TEXT,
			amount INTEGER NOT NULL,
			description TEXT,
			date DATETIME NOT NULL,
			fitid TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
			FOREIGN KEY (transfer_to_account_id) REFERENCES accounts(id) ON DELETE CASCADE,
			FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create new transactions table: %w", err)
	}

	// Step 3: Copy existing transactions (all as 'normal' type)
	_, err = tx.Exec(`
		INSERT INTO transactions_new (id, type, account_id, transfer_to_account_id, category_id, amount, description, date, fitid, created_at, updated_at)
		SELECT id, 'normal', account_id, NULL, category_id, amount, description, date, fitid, created_at, updated_at
		FROM transactions
	`)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// Step 4: Drop old table and rename
	_, err = tx.Exec("DROP TABLE transactions")
	if err != nil {
		return fmt.Errorf("failed to drop old table: %w", err)
	}

	_, err = tx.Exec("ALTER TABLE transactions_new RENAME TO transactions")
	if err != nil {
		return fmt.Errorf("failed to rename table: %w", err)
	}

	// Step 5: Recreate indexes
	_, err = tx.Exec(`
		CREATE INDEX idx_transactions_account_id ON transactions(account_id);
		CREATE INDEX idx_transactions_category_id ON transactions(category_id);
		CREATE INDEX idx_transactions_date ON transactions(date);
		CREATE INDEX idx_transactions_fitid ON transactions(fitid);
		CREATE INDEX idx_transactions_transfer_to_account_id ON transactions(transfer_to_account_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to recreate indexes: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// rollbackAddCreditCardSupport removes type and transfer_to_account_id columns
func rollbackAddCreditCardSupport(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if there are any transfer transactions
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM transactions WHERE type = 'transfer'").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for transfer transactions: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("cannot rollback: %d transfer transactions exist", count)
	}

	// Create new transactions table without type and transfer_to_account_id
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

	// Copy all data
	_, err = tx.Exec(`
		INSERT INTO transactions_new (id, account_id, category_id, amount, description, date, fitid, created_at, updated_at)
		SELECT id, account_id, category_id, amount, description, date, fitid, created_at, updated_at
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

// migrateSimplifyGroups removes the type field from category_groups table
func migrateSimplifyGroups(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if type column exists
	var columnExists int
	err = tx.QueryRow("SELECT COUNT(*) FROM pragma_table_info('category_groups') WHERE name='type'").Scan(&columnExists)
	if err != nil {
		return fmt.Errorf("failed to check for type column: %w", err)
	}

	// If type column exists, recreate table without it
	if columnExists > 0 {
		// Create new table without type field
		_, err = tx.Exec(`
			CREATE TABLE category_groups_new (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				description TEXT,
				display_order INTEGER NOT NULL DEFAULT 0,
				created_at DATETIME NOT NULL,
				updated_at DATETIME NOT NULL
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create new category_groups table: %w", err)
		}

		// Copy data (type field is dropped)
		_, err = tx.Exec(`
			INSERT INTO category_groups_new (id, name, description, display_order, created_at, updated_at)
			SELECT id, name, description, display_order, created_at, updated_at
			FROM category_groups
		`)
		if err != nil {
			return fmt.Errorf("failed to copy data: %w", err)
		}

		// Drop old table
		_, err = tx.Exec("DROP TABLE category_groups")
		if err != nil {
			return fmt.Errorf("failed to drop old table: %w", err)
		}

		// Rename new table
		_, err = tx.Exec("ALTER TABLE category_groups_new RENAME TO category_groups")
		if err != nil {
			return fmt.Errorf("failed to rename table: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// rollbackSimplifyGroups adds back the type field to category_groups
func rollbackSimplifyGroups(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if type column exists
	var columnExists int
	err = tx.QueryRow("SELECT COUNT(*) FROM pragma_table_info('category_groups') WHERE name='type'").Scan(&columnExists)
	if err != nil {
		return fmt.Errorf("failed to check for type column: %w", err)
	}

	// If type doesn't exist, add it back
	if columnExists == 0 {
		_, err = tx.Exec("ALTER TABLE category_groups ADD COLUMN type TEXT NOT NULL DEFAULT 'expense' CHECK(type IN ('income', 'expense'))")
		if err != nil {
			return fmt.Errorf("failed to add type column: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
