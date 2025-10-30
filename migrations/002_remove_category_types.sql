-- Migration to remove category types and make transaction category_id optional
-- This simplifies the system: categories are for budgeting expenses only,
-- and income transactions don't need to be categorized

-- Step 1: Create new categories table without type column
CREATE TABLE IF NOT EXISTS categories_new (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    color TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

-- Step 2: Copy data from old categories table (only expense categories)
-- We drop income categories since they're no longer needed
INSERT INTO categories_new (id, name, description, color, created_at, updated_at)
SELECT id, name, description, color, created_at, updated_at
FROM categories
WHERE type = 'expense';

-- Step 3: Create new transactions table with nullable category_id
CREATE TABLE IF NOT EXISTS transactions_new (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    category_id TEXT,  -- Now nullable
    amount INTEGER NOT NULL,
    description TEXT,
    date DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories_new(id) ON DELETE CASCADE
);

-- Step 4: Copy transactions
-- For income transactions (amount > 0), set category_id to NULL if category was income type
-- For expense transactions (amount < 0), keep category_id
INSERT INTO transactions_new (id, account_id, category_id, amount, description, date, created_at, updated_at)
SELECT
    t.id,
    t.account_id,
    CASE
        WHEN t.amount > 0 THEN NULL  -- Income transactions don't need category
        ELSE t.category_id           -- Expense transactions keep their category
    END as category_id,
    t.amount,
    t.description,
    t.date,
    t.created_at,
    t.updated_at
FROM transactions t;

-- Step 5: Drop old tables and rename new ones
DROP TABLE transactions;
DROP TABLE categories;
ALTER TABLE categories_new RENAME TO categories;
ALTER TABLE transactions_new RENAME TO transactions;

-- Step 6: Recreate indexes
CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions(account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_category_id ON transactions(category_id);
CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(date);

-- Verify the result
SELECT 'Migration Complete' as status,
       (SELECT COUNT(*) FROM categories) as category_count,
       (SELECT COUNT(*) FROM transactions) as transaction_count,
       (SELECT COUNT(*) FROM transactions WHERE category_id IS NULL) as uncategorized_count;
