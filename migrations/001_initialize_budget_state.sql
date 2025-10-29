-- Migration to initialize budget_state based on existing data
-- This should be run once when upgrading to the new system

-- Calculate the correct Ready to Assign value from existing data
-- Formula: Total Account Balance - Total Available (across all categories)
--
-- Where Available = Allocated - Spent for each category
--
-- This works because:
-- - We base it on actual account balances (which are real)
-- - We don't try to reconstruct income history (which may be incomplete)
-- - Available = money that's been allocated but not yet spent

UPDATE budget_state
SET ready_to_assign = (
    -- Total in all accounts
    SELECT COALESCE(SUM(balance), 0)
    FROM accounts
) - (
    -- Total Available across all categories
    -- Available = Allocated - Spent
    SELECT COALESCE(
        -- Total Allocated
        (SELECT COALESCE(SUM(amount), 0) FROM allocations) -
        -- Minus Total Spent (negative transactions)
        (SELECT COALESCE(SUM(ABS(amount)), 0) FROM transactions WHERE amount < 0),
    0)
),
updated_at = datetime('now')
WHERE id = 'singleton';

-- Verify the result
SELECT
    'Migration Complete' as status,
    ready_to_assign as new_ready_to_assign,
    (SELECT SUM(balance) FROM accounts) as total_account_balance,
    (SELECT SUM(amount) FROM allocations) as total_allocated,
    (SELECT SUM(ABS(amount)) FROM transactions WHERE amount < 0) as total_spent
FROM budget_state
WHERE id = 'singleton';
