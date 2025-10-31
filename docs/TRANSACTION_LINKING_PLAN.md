# Transaction Linking - Comprehensive Implementation Plan

**Version:** 2.0
**Updated:** 2025-10-31
**Status:** Planning Phase

---

## 1. Executive Summary

This plan outlines the implementation of automatic transaction linking for imported transactions. The system will detect when transactions across multiple accounts represent the same money transfer (e.g., $1,000 leaving Chase Checking and arriving at OnPoint Checking), and allow users to link them as transfers rather than categorizing them as separate income/expense events.

### Key Features
- **Automatic detection** of potential transfer matches during import
- **User confirmation** workflow (semi-automatic, not fully automated)
- **Credit card payment** special handling with payment category application
- **Manual linking** for existing transactions
- **Cross-month support** with proper budget impact handling

---

## 2. Problem Statement

### Current Behavior
- **Manual Transfers**: Work perfectly - create two linked transactions using `transfer_to_account_id`
- **Imported Transactions**: All come in as `type='normal'`, `category_id=null`, no linking detected

### The Problem
When importing transactions from multiple accounts at different times:
```
Chase Checking:     -$1,000  "Transfer to OnPoint"     (needs linking)
OnPoint Checking:   +$1,000  "Transfer from Chase"     (needs linking)
OnPoint Checking:   -$1,000  "Mortgage Payment"        (separate, needs category)
```

Without linking:
- Both transfer transactions show as uncategorized
- Clutter the "Needs Categorization" list
- May affect budget calculations if manually categorized incorrectly
- Require manual cleanup

---

## 3. Current System Analysis

### Transaction Types
```go
type TransactionType string

const (
    TransactionTypeNormal   TransactionType = "normal"
    TransactionTypeTransfer TransactionType = "transfer"
)
```

### Transaction Behaviors

| Type | Amount Sign | Category | RTA Impact | Budget Impact | Account Balance |
|------|-------------|----------|------------|---------------|----------------|
| **Inflow** | Positive | Optional | ✅ Increases | None | ✅ Increases |
| **Outflow** (debit) | Negative | Required | None | ✅ Category spending | ✅ Decreases |
| **Outflow** (credit card) | Negative | Required | None | ✅ Category spending | ✅ More negative (debt↑) |
| **Transfer** (accounts) | Both +/- | None | None | None | ✅ Both sides |
| **Transfer** (CC payment) | Both +/- | Payment cat* | None | ✅ Payment category | ✅ Both sides |

\* Payment category applied only to outbound side (from checking), only if within budget.

### Current Transfer Implementation
**File**: `internal/application/transaction_service.go:200-327`

Manual transfers create TWO linked transactions:
```go
// Outbound transaction
outboundTxn := &domain.Transaction{
    Type: domain.TransactionTypeTransfer,
    AccountID: fromAccountID,
    TransferToAccountID: &toAccountID,  // Link to destination
    Amount: -amount,
    CategoryID: outboundCategoryID,     // Payment category for CC payments
}

// Inbound transaction
inboundTxn := &domain.Transaction{
    Type: domain.TransactionTypeTransfer,
    AccountID: toAccountID,
    TransferToAccountID: &fromAccountID,  // Link back to source
    Amount: amount,
    CategoryID: nil,
}
```

### Import System
**File**: `internal/application/import_service.go:47-158`

1. Parses OFX file
2. Uses FitID for duplicate detection
3. Creates transactions as `type='normal'`, `category_id=null`
4. Updates account balance to ledger balance from OFX
5. Adjusts Ready to Assign by balance delta

**Key limitation**: No matching logic during or after import.

---

## 4. Solution Architecture

### Three-Phase Approach

#### Phase 1: Detection & Suggestion (Automatic)
During or immediately after import, automatically detect potential transfer pairs and create suggestions.

#### Phase 2: User Confirmation (Semi-Automatic)
Present suggested matches to the user for approval before linking.

#### Phase 3: Manual Linking (User-Initiated)
Allow users to manually link any two transactions they know are transfers.

---

## 5. Matching Algorithm

### 5.1 Core Matching Criteria

**Required conditions** (all must be true):
1. **Amount Match**: `|amount_a| == |amount_b|` (absolute values equal)
2. **Opposite Signs**: One negative, one positive
3. **Date Proximity**: Within ±3 days (configurable)
4. **Account Ownership**: Both accounts belong to same user
5. **Type Check**: Both are `type='normal'` (not already transfers)
6. **Not Already Linked**: Neither has `transfer_to_account_id` set
7. **Different Accounts**: Must be in different accounts

### 5.2 Scoring Heuristics

For ranking multiple potential matches:

| Condition | Points |
|-----------|--------|
| Same date | +10 |
| 1 day apart | +5 |
| 2-3 days apart | +2 |
| Description similarity* | +5 |
| Round amount** | +3 |
| Credit card payment*** | +5 |

\* Contains keywords: "transfer", "xfer", "from", "to", "payment"
\*\* Ends in .00 (e.g., $1000.00 vs $1,237.48)
\*\*\* One account is credit card type

### 5.3 Confidence Levels

- **High** (≥15 points): Same date, similar description, round amount
- **Medium** (10-14 points): Close dates, amount match only
- **Low** (<10 points): Weak signals, may be coincidence

### 5.4 Algorithm Pseudocode

```
FUNCTION findTransferMatches(accountID, importedTransactionIDs):
    suggestions = []

    FOR EACH txnID IN importedTransactionIDs:
        txn = getTransaction(txnID)

        IF txn.type != 'normal' OR txn.transfer_to_account_id != null:
            CONTINUE  // Already a transfer

        // Search for matching transaction in OTHER accounts (same user)
        candidates = findCandidates(
            userID = txn.user_id,
            excludeAccountID = txn.account_id,
            amount = -txn.amount,  // Opposite sign
            dateMin = txn.date - 3 days,
            dateMax = txn.date + 3 days,
            type = 'normal',
            unlinked = true
        )

        FOR EACH candidate IN candidates:
            score = calculateMatchScore(txn, candidate)
            confidence = classifyConfidence(score)

            IF score >= 10:  // Minimum threshold
                suggestions.APPEND({
                    transaction_a_id: txn.id,
                    transaction_b_id: candidate.id,
                    score: score,
                    confidence: confidence,
                    is_credit_payment: isOneAccountCredit(txn, candidate)
                })

    RETURN suggestions
```

### 5.5 Asynchronous Import Support

**Works when files uploaded at different times:**

When OnPoint import runs (Day 5), it searches ALL existing Chase transactions:
```sql
SELECT * FROM transactions
WHERE account_id IN (user's_other_accounts)
  AND type = 'normal'
  AND transfer_to_account_id IS NULL
  AND amount = -1000  -- Opposite sign
  AND date >= '2025-10-25'  -- txn.date minus 3 days
  AND date <= '2025-10-31'  -- txn.date plus 3 days
```

**Import order doesn't matter:**
- ✅ Chase first, OnPoint later → Match found when OnPoint imports
- ✅ OnPoint first, Chase later → Match found when Chase imports
- ✅ Different date ranges → Matches found where dates overlap
- ✅ Overlapping ranges → Duplicates filtered by FitID

---

## 6. Credit Card Payment Handling

### 6.1 The Special Case

Credit card payments are transfers but require payment category tracking:

```
Manual Payment (current):
  Checking: type='transfer', category='Chase CC Payment'
  Credit:   type='transfer', category=null

Imported Payment (needs special handling):
  Checking: type='normal', category=null  (imported)
  Credit:   type='normal', category=null  (imported)

  After linking (must match manual behavior):
  Checking: type='transfer', category='Chase CC Payment' ✅
  Credit:   type='transfer', category=null
```

### 6.2 Detection Enhancement

Add to scoring:
```go
if (accountA.Type == AccountTypeCredit || accountB.Type == AccountTypeCredit) {
    score += 5  // Boost for CC payments

    if containsPaymentKeywords(description) {
        score += 3  // "payment", "autopay", etc.
    }
}
```

### 6.3 Linking Logic for CC Payments

```go
func linkTransactions(txnA, txnB) {
    // Determine which is credit card
    var checkingSide, creditSide *Transaction
    if accountA.Type == AccountTypeCredit {
        creditSide = txnA
        checkingSide = txnB
    } else if accountB.Type == AccountTypeCredit {
        creditSide = txnB
        checkingSide = txnA
    }

    // Apply payment category if CC payment
    var checkingCategoryID *string
    if creditSide != nil {
        paymentCat := getPaymentCategoryByAccountID(creditSide.AccountID)
        if paymentCat != nil {
            available := calculateAvailableInCategory(paymentCat.ID)
            if available >= abs(checkingSide.Amount) {
                checkingCategoryID = &paymentCat.ID  // Apply payment category
            }
        }
    }

    // Convert both to transfers
    checkingSide.Type = TransactionTypeTransfer
    checkingSide.TransferToAccountID = &creditSide.AccountID
    checkingSide.CategoryID = checkingCategoryID

    creditSide.Type = TransactionTypeTransfer
    creditSide.TransferToAccountID = &checkingSide.AccountID
    creditSide.CategoryID = nil
}
```

### 6.4 CC Payment Edge Cases

**Overpayment** (paying more than budgeted):
```
Budget: $500 allocated to CC Payment
Already paid: $300
Available: $200
Import: $250 payment

Action: Link as transfer, but don't categorize (exceeds budget)
Result: Transfer works, budget shows $200 available (correct)
```

**Multiple payments in one month:**
```
Oct 5:  Pay $200
Oct 15: Pay $300
Budget: $500 total

Linking:
  First:  $200 categorized (500 - 0 = 500 available) ✅
  Second: $300 categorized (500 - 200 = 300 available) ✅
```

---

## 7. Database Schema Changes

### 7.1 New Table: `transfer_match_suggestions`

```sql
CREATE TABLE transfer_match_suggestions (
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

    -- Prevent duplicate suggestions
    UNIQUE(transaction_a_id, transaction_b_id)
);

CREATE INDEX idx_transfer_suggestions_status ON transfer_match_suggestions(status);
CREATE INDEX idx_transfer_suggestions_confidence ON transfer_match_suggestions(confidence);
CREATE INDEX idx_transfer_suggestions_credit ON transfer_match_suggestions(is_credit_payment);
```

### 7.2 Performance Index

```sql
-- Speed up candidate search
CREATE INDEX idx_transactions_matching ON transactions(type, amount, date)
    WHERE transfer_to_account_id IS NULL;
```

---

## 8. Implementation Roadmap

### Sprint 1: Foundation (1-2 weeks)
- [ ] Create migration 004 for `transfer_match_suggestions` table
- [ ] Add performance indexes
- [ ] Implement `TransferSuggestion` domain model
- [ ] Implement `TransferSuggestionRepository`
- [ ] Write repository unit tests

### Sprint 2: Matching Algorithm (1-2 weeks)
- [ ] Implement `TransferMatcherService`
- [ ] Implement scoring algorithm with all heuristics
- [ ] Add credit card payment detection
- [ ] Add candidate search with date window
- [ ] Write comprehensive unit tests (20+ test cases)
- [ ] Performance test with 1000+ transactions

### Sprint 3: Import Integration (1 week)
- [ ] Modify `ImportFromOFX` to call matcher
- [ ] Handle matcher errors gracefully (don't fail import)
- [ ] Add logging for debugging
- [ ] Test with real OFX files from Chase, OnPoint, etc.

### Sprint 4: API Endpoints (1 week)
- [ ] `GET /api/transfer-suggestions` - List pending suggestions
- [ ] `POST /api/transfer-suggestions/{id}/accept` - Accept and link
- [ ] `POST /api/transfer-suggestions/{id}/reject` - Reject suggestion
- [ ] `POST /api/transactions/link` - Manual linking
- [ ] API integration tests

### Sprint 5: UI Implementation (1-2 weeks)
- [ ] Suggestion list page with filtering
- [ ] Individual suggestion card (accept/reject)
- [ ] Credit card payment badge display
- [ ] Notification badge for pending suggestions
- [ ] Manual link mode in transaction list
- [ ] Bulk actions (accept/reject multiple)
- [ ] Loading states and error handling

### Sprint 6: Polish & Testing (1 week)
- [ ] Handle amount tolerance for fees (±2% or $5)
- [ ] Confirmation dialogs for edge cases
- [ ] Cross-month transfer testing
- [ ] User acceptance testing
- [ ] Documentation updates

---

## 9. API Specification

### 9.1 List Suggestions

```
GET /api/transfer-suggestions
Query Parameters:
  - confidence?: "high" | "medium" | "low"
  - status?: "pending" | "accepted" | "rejected"
  - credit_only?: boolean

Response:
{
  "suggestions": [
    {
      "id": "suggestion-uuid",
      "transaction_a": { /* full transaction */ },
      "transaction_b": { /* full transaction */ },
      "account_a": { /* account details */ },
      "account_b": { /* account details */ },
      "confidence": "high",
      "score": 18,
      "is_credit_payment": true,
      "created_at": "2025-10-31T10:00:00Z"
    }
  ]
}
```

### 9.2 Accept Suggestion

```
POST /api/transfer-suggestions/{id}/accept

Response:
{
  "success": true,
  "linked_transactions": [
    { /* updated transaction A */ },
    { /* updated transaction B */ }
  ]
}

Error Responses:
- 404: Suggestion not found
- 400: Suggestion already accepted/rejected
- 409: Transactions already linked
```

### 9.3 Reject Suggestion

```
POST /api/transfer-suggestions/{id}/reject

Response:
{
  "success": true
}
```

### 9.4 Manual Link

```
POST /api/transactions/link
Body:
{
  "transaction_a_id": "uuid",
  "transaction_b_id": "uuid"
}

Validations:
- Both transactions exist and belong to user
- Different accounts
- Opposite signs (or at least different accounts)
- Neither already linked
- Both type='normal'

Response:
{
  "success": true,
  "linked_transactions": [
    { /* updated transaction A */ },
    { /* updated transaction B */ }
  ]
}
```

---

## 10. Edge Cases & Solutions

### 10.1 Cross-Month Transfers

**Scenario:**
```
Sept 30: Chase -$1,000 (sent)
Oct 2:   OnPoint +$1,000 (received)
```

**Budget Impact:**
- September: -$1,000 from total cash (correct - money left)
- October: +$1,000 to total cash (correct - money arrived)
- Ready to Assign: Temporarily lower in September (accurate - in transit)
- Categories: No impact (transfers excluded)

**Solution:** Display paired transaction date in UI:
```
Sept 30 • Transfer: Chase → OnPoint
ℹ️  Received in OnPoint on Oct 2
```

### 10.2 Amount Mismatches (Fees)

**Scenario:**
```
Chase:   -$1,000.00
OnPoint: +$998.00  ($2 wire fee)
```

**Solution (Future Enhancement):**
- Allow configurable tolerance (±2% or $5, whichever smaller)
- Show amount discrepancy in UI
- Option to create third transaction for fee

**Current:** Only exact matches (Phase 1)

### 10.3 Already Categorized

**Scenario:** User already manually categorized import as "Misc"

**Solution:**
- Still show suggestion
- When linking, warn: "This will change category from 'Misc' to 'Transfer' (no category). Continue?"
- For CC payments: "This will change to 'CC Payment' category. Continue?"

### 10.4 Same-Day, Same-Amount Non-Transfers

**Scenario:**
```
Chase:   -$50 (grocery store)
OnPoint: +$50 (paycheck reimbursement)
Same day, same amount, but NOT a transfer
```

**Solution:**
- Scoring will be medium/low (no description match)
- User rejects suggestion
- Marked as rejected, won't appear again

### 10.5 Split Transfers (Future)

**Scenario:**
```
Chase:   -$1,000
OnPoint: +$700
Savings: +$300
```

**Solution:** Phase 4 enhancement - one-to-many linking

---

## 11. Testing Strategy

### 11.1 Unit Tests

**Matching Algorithm:**
- Same date, opposite amounts → High score
- Different dates (within window) → Medium score
- Outside date window → No match
- Same user, different accounts → Match
- Different users → No match
- Already linked → Excluded
- Round amounts → Score boost
- Description keywords → Score boost
- Credit card payment → Score boost + flag set

**Repository:**
- Create suggestion
- List pending suggestions
- Accept/reject suggestion
- Delete by transaction ID
- Uniqueness constraint

### 11.2 Integration Tests

**Import Flow:**
1. Import Chase transactions
2. Import OnPoint transactions
3. Assert suggestions created
4. Assert correct confidence levels
5. Accept suggestion
6. Assert both transactions converted to transfers
7. Assert cross-references set

**CC Payment Flow:**
1. Import checking payment
2. Import credit card receipt
3. Assert credit payment suggestion created
4. Accept suggestion
5. Assert payment category applied to checking side
6. Assert no category on credit side
7. Assert budget reflects payment

**Cross-Month Flow:**
1. Import Sept 30 transaction
2. Import Oct 2 transaction
3. Assert match found (within 3 days)
4. Accept link
5. Assert both transactions remain on original dates
6. Assert balances correct for both months

### 11.3 User Acceptance Testing

**Scenarios:**
1. Import two accounts with obvious transfer → High confidence suggestion
2. Import two accounts with no transfers → No false positives
3. Import with 3-day timing delay → Match found
4. Credit card payment → Payment category applied
5. Reject suggestion → Transactions stay normal
6. Manual link two old transactions → Link created
7. Cross-month transfer → Budget correct in both months

---

## 12. Configuration

```go
type TransferMatchConfig struct {
    Enabled              bool    // Feature flag
    MaxDateDiffDays      int     // Default: 3
    MinConfidenceScore   int     // Default: 10
    AmountTolerancePct   float64 // Default: 0.0 (exact match only in Phase 1)
    AmountToleranceCents int64   // Default: 0
    AutoLinkHighConf     bool    // Default: false (always require confirmation)
}
```

**Recommended Phase 1 settings:**
- Max date diff: 3 days
- Min score: 10 (medium confidence threshold)
- Amount tolerance: 0 (exact match only)
- Auto-link: false (always require user confirmation)

---

## 13. Performance Considerations

### 13.1 Query Optimization

**Challenge:** Finding candidates across all user accounts is expensive.

**Solutions:**
1. **Index**: `(type, amount, date)` for fast candidate lookup
2. **Time window**: Only search last 90 days
3. **Batch processing**: Search all candidates once, match in-memory
4. **Limit scope**: Only run on newly imported transactions

**Expected performance:**
- Small import (10 transactions): <100ms
- Large import (500 transactions): <5s
- Very large import (1000+ transactions): Consider async job

### 13.2 Scaling Strategy

**Phase 1:** Synchronous during import (blocking)
**Phase 4:** Async job queue for large imports

---

## 14. Success Metrics

**Adoption:**
- % of users who review suggestions
- % of imports that generate suggestions

**Accuracy:**
- Acceptance rate by confidence level (target: >70% for high)
- Rejection rate (target: <30% overall)

**User Impact:**
- Reduction in uncategorized transactions
- Time saved vs manual categorization
- User satisfaction (surveys)

**Technical:**
- Average suggestions per import
- Average matching time
- False positive rate

---

## 15. Future Enhancements (Phase 4+)

### 15.1 Machine Learning
- Track user accept/reject patterns
- Learn description patterns per bank
- Adjust scoring weights based on user behavior

### 15.2 Recurring Transfers
- Detect monthly patterns (e.g., automatic savings)
- Auto-link recurring transfers with high confidence
- "Always link these accounts" preference

### 15.3 Split Transactions
- One-to-many linking (one source, multiple destinations)
- Sum matching algorithm
- UI to select multiple transactions

### 15.4 Transfer Rules
- User-defined auto-link rules
- "Always link Chase→OnPoint transfers"
- Safer than full automation

### 15.5 Amount Tolerance
- ±2% or $5 tolerance for fees
- Suggest fee transaction creation
- Three-transaction linking (transfer + fee)

---

## 16. Security & Privacy

**User Isolation:**
- All queries filtered by user ID
- No cross-user matching (privacy protection)
- Suggestions scoped to user's accounts only

**Data Protection:**
- Suggestions are user-specific
- Deleted when user deleted (CASCADE)
- No sensitive data in suggestion table (just IDs)

**Audit Trail:**
- Log all accept/reject actions
- Track who linked what and when
- Enable support debugging

---

## 17. Documentation Requirements

**User Documentation:**
- How transaction linking works
- How to review suggestions
- How to manually link transactions
- Understanding credit card payments
- Cross-month transfer explanation

**Developer Documentation:**
- Matching algorithm details
- Database schema
- API documentation
- Testing guide
- Troubleshooting guide

---

## 18. Rollout Strategy

### Phase 1 (MVP): Basic Linking
- Exact amount matching only
- 3-day window
- User confirmation required
- Credit card payment support

### Phase 2: Enhanced Matching
- Description similarity
- Better scoring
- Performance optimization

### Phase 3: User Features
- Manual linking
- Bulk actions
- Better UI/UX

### Phase 4: Advanced Features
- Amount tolerance
- Split transactions
- Recurring detection
- ML improvements

---

## 19. Open Questions

1. Should we support unlinking of transfers? (Undo)
2. Should suggestions expire after X days?
3. Should we notify users of new suggestions?
4. Should we support "never suggest these two accounts" rules?
5. What happens to suggestions when transactions are deleted?

---

## 20. Dependencies

**Technical:**
- Database migration system
- OFX parser (existing)
- Transaction repository (existing)
- Category repository (existing)

**Business:**
- User acceptance of semi-automatic system
- Testing with real bank data
- UI/UX design for suggestion review

---

## Next Steps

1. ✅ Review and approve this plan
2. Begin Sprint 1: Database migration and foundation
3. Build matching algorithm with comprehensive tests
4. Integrate with import flow
5. Create API endpoints
6. Build UI for suggestion review
7. User acceptance testing
8. Deploy to production with feature flag

---

**Plan Status:** Ready for review and approval
**Estimated Total Time:** 8-12 weeks
**Risk Level:** Low (additive feature, doesn't break existing functionality)
