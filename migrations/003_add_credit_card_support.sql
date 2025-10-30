-- Migration to add credit card and transfer support
-- Adds credit card account type, transaction types, and payment categories

-- Step 1: Add payment_for_account_id column to categories
ALTER TABLE categories ADD COLUMN payment_for_account_id TEXT REFERENCES accounts(id) ON DELETE CASCADE;

-- Step 2: Create new transactions table with type and transfer support
CREATE TABLE IF NOT EXISTS transactions_new (
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

-- Step 3: Copy existing transactions (all as 'normal' type)
INSERT INTO transactions_new (id, type, account_id, transfer_to_account_id, category_id, amount, description, date, created_at, updated_at)
SELECT id, 'normal', account_id, NULL, category_id, amount, description, date, created_at, updated_at
FROM transactions;

-- Step 4: Drop old table and rename
DROP TABLE transactions;
ALTER TABLE transactions_new RENAME TO transactions;

-- Step 5: Recreate indexes
CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions(account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_category_id ON transactions(category_id);
CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(date);
CREATE INDEX IF NOT EXISTS idx_transactions_transfer_to_account_id ON transactions(transfer_to_account_id);

-- Verify the result
SELECT 'Migration Complete' as status,
       (SELECT COUNT(*) FROM transactions) as transaction_count,
       (SELECT COUNT(*) FROM transactions WHERE type = 'normal') as normal_count,
       (SELECT COUNT(*) FROM transactions WHERE type = 'transfer') as transfer_count;
