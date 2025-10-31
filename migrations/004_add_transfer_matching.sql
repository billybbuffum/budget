-- Migration to add transaction linking and transfer matching support
-- Adds transfer_match_suggestions table and performance indexes for matching algorithm

-- Step 1: Add fitid column to transactions if it doesn't exist (for OFX duplicate detection)
-- SQLite doesn't support ADD COLUMN IF NOT EXISTS, so we check first
-- This is safe to run even if the column already exists (will fail silently in that case)
ALTER TABLE transactions ADD COLUMN fitid TEXT;

-- Step 2: Create transfer_match_suggestions table
CREATE TABLE IF NOT EXISTS transfer_match_suggestions (
    id TEXT PRIMARY KEY,
    transaction_a_id TEXT NOT NULL,
    transaction_b_id TEXT NOT NULL,
    confidence TEXT NOT NULL CHECK(confidence IN ('high', 'medium', 'low')),
    score INTEGER NOT NULL,
    is_credit_payment BOOLEAN NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'accepted', 'rejected')),
    created_at DATETIME NOT NULL,
    reviewed_at DATETIME,

    FOREIGN KEY (transaction_a_id) REFERENCES transactions(id) ON DELETE CASCADE,
    FOREIGN KEY (transaction_b_id) REFERENCES transactions(id) ON DELETE CASCADE,

    -- Prevent duplicate suggestions (same pair shouldn't be suggested twice)
    UNIQUE(transaction_a_id, transaction_b_id)
);

-- Step 3: Create indexes for transfer_match_suggestions table
CREATE INDEX IF NOT EXISTS idx_transfer_suggestions_status ON transfer_match_suggestions(status);
CREATE INDEX IF NOT EXISTS idx_transfer_suggestions_confidence ON transfer_match_suggestions(confidence);
CREATE INDEX IF NOT EXISTS idx_transfer_suggestions_credit ON transfer_match_suggestions(is_credit_payment);
CREATE INDEX IF NOT EXISTS idx_transfer_suggestions_txn_a ON transfer_match_suggestions(transaction_a_id);
CREATE INDEX IF NOT EXISTS idx_transfer_suggestions_txn_b ON transfer_match_suggestions(transaction_b_id);

-- Step 4: Create performance index for matching algorithm
-- This index speeds up the candidate search query (finding opposite amounts within date window)
CREATE INDEX IF NOT EXISTS idx_transactions_matching ON transactions(type, amount, date)
    WHERE transfer_to_account_id IS NULL;

-- Step 5: Create index for FitID lookups (OFX duplicate detection)
CREATE INDEX IF NOT EXISTS idx_transactions_fitid ON transactions(account_id, fitid)
    WHERE fitid IS NOT NULL;

-- Verify the result
SELECT 'Migration Complete' as status,
       (SELECT COUNT(*) FROM transfer_match_suggestions) as suggestion_count,
       'Indexes created for matching performance' as note;
