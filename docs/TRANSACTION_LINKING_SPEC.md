# Transaction Linking - Technical Specification

**Version:** 1.0
**Date:** 2025-10-31

---

## Table of Contents
1. [System Overview](#system-overview)
2. [All Scenarios](#all-scenarios)
3. [Test Cases](#test-cases)
4. [High-Level System Design](#high-level-system-design)

---

## System Overview

### Purpose
Automatically detect and link imported transactions that represent the same money transfer between user accounts, eliminating the need for manual categorization of internal transfers.

### Core Concept
When money moves between user accounts:
- **Without Linking**: Two uncategorized transactions requiring manual cleanup
- **With Linking**: Two linked transfer transactions, automatically excluded from budget

### Transaction Model
```
Normal Transaction:   type='normal', amount=±X, category_id=required (outflows)
Transfer Transaction: type='transfer', amount=±X, transfer_to_account_id=uuid, category_id=null*

* Exception: Credit card payments have payment category on checking side
```

### Three-Phase Process
1. **Detection**: Automatically find potential matches after import
2. **Confirmation**: User reviews and approves/rejects suggestions
3. **Linking**: Convert approved pairs to transfer transactions

---

## All Scenarios

### Scenario 1: Basic Transfer (Same Day)
**Description:** Transfer between checking accounts on the same day

**Input:**
```
Chase Checking:    Oct 28, -$1,000.00, "Transfer to OnPoint"
OnPoint Checking:  Oct 28, +$1,000.00, "Transfer from Chase"
```

**Expected:**
- Match detected: High confidence (same date, opposite amounts)
- Score: 15+ points
- User approves
- Both converted to type='transfer' with cross-references
- No category on either transaction
- No budget impact
- Account balances updated correctly

**Budget Impact:**
- Ready to Assign: No change (money moved, not gained/lost)
- Categories: No change (transfers excluded)

---

### Scenario 2: Delayed Transfer (Cross-Day)
**Description:** Transfer takes 2 business days to process

**Input:**
```
Chase Checking:    Oct 28, -$1,000.00, "ACH Transfer"
OnPoint Checking:  Oct 30, +$1,000.00, "ACH Deposit"
```

**Expected:**
- Match detected: High confidence (within 3-day window)
- Score: 12-14 points (date proximity penalty)
- User approves
- Both converted to transfers
- Dates remain unchanged (Oct 28 and Oct 30)

**Budget Impact:**
- Oct 28: RTA decreases by $1,000 (money left)
- Oct 30: RTA increases by $1,000 (money arrived)
- Net: $0 over time
- In-transit period (Oct 28-30): Money temporarily unavailable

---

### Scenario 3: Cross-Month Transfer
**Description:** Transfer sent last day of month, received first day of next month

**Input:**
```
Chase Checking:    Sept 30, -$1,000.00, "Transfer"
OnPoint Checking:  Oct 2, +$1,000.00, "Deposit"
```

**Expected:**
- Match detected: High confidence (within 3-day window)
- Both converted to transfers
- Dates remain on original months

**Budget Impact:**
- September: -$1,000 from total cash
- October: +$1,000 to total cash
- Ready to Assign: Lower in Sept, higher in Oct
- Monthly reports show split (correct behavior)

**UI Enhancement:** Show paired transaction date
```
Sept 30 • Transfer: Chase → OnPoint
ℹ️  Received in OnPoint on Oct 2
```

---

### Scenario 4: Credit Card Payment (Same Day)
**Description:** Paying credit card from checking account

**Input:**
```
Chase Checking:    Oct 28, -$500.00, "Payment to Chase Credit"
Chase Credit Card: Oct 28, +$500.00, "Payment Received"
```

**Expected:**
- Match detected: High confidence (same date, CC flag set)
- Score: 20+ points (CC payment boost)
- User approves
- Both converted to transfers
- **Special:** Checking side gets payment category applied
- Credit side has no category

**Result:**
```
Chase Checking:    type='transfer', category='Chase CC Payment', amount=-500
Chase Credit Card: type='transfer', category=null, amount=+500
```

**Budget Impact:**
- Ready to Assign: No change
- Chase CC Payment category: Activity +$500, Available -$500
- Credit card balance: -$500 → $0 (debt reduced)

---

### Scenario 5: CC Payment (Delayed, Cross-Month)
**Description:** Credit card payment initiated Sept 30, posted Oct 2

**Input:**
```
Chase Checking:    Sept 30, -$500.00, "CC Payment"
Chase Credit Card: Oct 2, +$500.00, "Payment"
```

**Expected:**
- Match detected: High confidence
- Both converted to transfers with payment category
- Dates remain on original months

**Budget Impact:**
- September: Chase CC Payment activity +$500
- October: Credit card balance improves by $500
- RTA: No change overall

---

### Scenario 6: CC Payment (Overpayment)
**Description:** Paying more than budgeted to credit card

**Input:**
```
Budgeted to CC Payment category: $500
Already paid this month: $300
Available: $200

Import:
  Chase Checking:    Oct 28, -$250.00, "Payment"
  Chase Credit Card: Oct 28, +$250.00, "Payment"
```

**Expected:**
- Match detected: High confidence
- Both converted to transfers
- **Special:** No payment category applied (exceeds available budget)
- User must manually categorize or leave uncategorized

**Result:**
```
Chase Checking:    type='transfer', category=null, amount=-250
Chase Credit Card: type='transfer', category=null, amount=+250
```

**Budget Impact:**
- CC Payment category: Still shows $200 available (correct)
- Overpayment doesn't show as negative available

---

### Scenario 7: CC Spending (Not a Transfer)
**Description:** Normal credit card purchase - should NOT match

**Input:**
```
Chase Credit Card: Oct 28, -$50.00, "Restaurant"
```

**Expected:**
- No match detected (only one transaction, no opposite pair)
- Remains as normal transaction
- Requires categorization

---

### Scenario 8: Coincidental Same-Amount (False Positive)
**Description:** Two unrelated transactions with same amount on same day

**Input:**
```
Chase Checking:    Oct 28, -$50.00, "Grocery Store"
OnPoint Checking:  Oct 28, +$50.00, "Employer Reimbursement"
```

**Expected:**
- Match detected: Medium/Low confidence (amount match, no description similarity)
- Score: 10-12 points
- User **rejects** suggestion
- Both remain as normal transactions
- User categorizes separately

**Outcome:**
- Suggestion marked as rejected
- Won't appear again
- Demonstrates importance of user confirmation

---

### Scenario 9: Multiple CC Payments in One Month
**Description:** Making partial credit card payments

**Input:**
```
Budgeted: $500 to CC Payment

Payment 1:
  Oct 5: Chase Checking -$200, Chase CC +$200

Payment 2:
  Oct 15: Chase Checking -$300, Chase CC +$300
```

**Expected:**
- Both matches detected
- Both converted to transfers with payment category

**Budget Tracking:**
```
After Payment 1:
  Available: $500 - $200 = $300
  Payment categorized ✅

After Payment 2:
  Available: $300 - $300 = $0
  Payment categorized ✅
```

---

### Scenario 10: Asynchronous Imports (Different Times)
**Description:** Importing accounts days apart

**Timeline:**
```
Monday (Day 1):    Import Chase (Oct 1-31)
Friday (Day 5):    Import OnPoint (Oct 1-31)
```

**Chase Import (Monday):**
- Creates: Chase -$1,000 (Oct 28)
- Matcher runs: Searches all OnPoint transactions
- No match found (OnPoint not imported yet)
- No suggestion created

**OnPoint Import (Friday):**
- Creates: OnPoint +$1,000 (Oct 28)
- Matcher runs: Searches all Chase transactions
- **Finds Chase -$1,000 from Monday!**
- Creates high confidence suggestion

**Expected:**
- Match detected despite 4-day import gap
- User approves
- Both linked correctly

**Key:** Transaction dates matter, not import dates

---

### Scenario 11: Overlapping Date Range Imports
**Description:** Importing overlapping periods

**Input:**
```
Import 1: Chase Oct 1-15
Import 2: Chase Oct 10-31 (overlaps by 5 days)
```

**Expected:**
- Duplicate transactions (Oct 10-15) detected by FitID
- Skipped automatically
- No duplicate suggestions
- Matching works for new transactions only

---

### Scenario 12: Manual Linking (Old Transactions)
**Description:** User wants to link transactions imported before feature existed

**Input:**
```
Existing transactions (both type='normal'):
  Chase:   Sept 1, -$500
  OnPoint: Sept 1, +$500
```

**User Action:**
1. Selects Chase transaction
2. Clicks "Link Transaction"
3. Selects OnPoint transaction
4. Clicks "Complete Link"

**System Validates:**
- ✅ Different accounts
- ✅ Same user
- ✅ Neither already linked
- ✅ Both type='normal'

**Expected:**
- Both converted to transfers
- Cross-references added
- If credit card involved: Payment category applied

---

### Scenario 13: Already Categorized Transaction
**Description:** User already categorized import before reviewing suggestions

**Input:**
```
Chase:   Oct 28, -$1,000, category='Misc'
OnPoint: Oct 28, +$1,000, category=null
```

**Expected:**
- Match detected
- User sees suggestion with warning:
  "This will change Chase transaction from 'Misc' to 'Transfer' (no category)"
- User approves
- Chase category removed
- Both linked as transfers

---

### Scenario 14: Outside Date Window
**Description:** Transfer takes 5 days (outside 3-day window)

**Input:**
```
Chase:   Oct 1, -$1,000
OnPoint: Oct 6, +$1,000
```

**Expected:**
- No automatic match (5 days > 3-day window)
- User must manually link
- Suggestion: Add config option to increase window to 7 days if needed

---

### Scenario 15: Round Amount Boost
**Description:** Round amounts are more likely transfers

**Input:**
```
Transaction A: Oct 28, -$1,000.00 (round)
Transaction B: Oct 28, +$1,000.00 (round)
```

**Expected:**
- Score boost: +3 points for round amount
- Higher confidence than non-round amounts

**Comparison:**
```
$1,000.00 → Round, score boost
$1,237.48 → Not round, no boost
```

---

### Scenario 16: Description Similarity Boost
**Description:** Transfer keywords in description

**Input:**
```
Chase:   "Transfer to OnPoint Checking"   (contains "transfer")
OnPoint: "Transfer from Chase Bank"       (contains "transfer")
```

**Expected:**
- Score boost: +5 points for description similarity
- Higher confidence than generic descriptions

**Keywords:** "transfer", "xfer", "from", "to", "payment"

---

### Scenario 17: Income vs Transfer Disambiguation
**Description:** Positive transaction that is NOT a transfer

**Input:**
```
Chase Checking: Oct 28, +$1,000, "Paycheck"
```

**Expected:**
- No match detected (no opposite transaction)
- Remains as normal inflow
- Increases Ready to Assign
- No categorization required

---

### Scenario 18: Expense vs Transfer Disambiguation
**Description:** Negative transaction that is NOT a transfer

**Input:**
```
Chase Checking: Oct 28, -$1,000, "Mortgage Payment"
```

**Expected:**
- No match detected (no opposite transaction)
- Remains as normal outflow
- Requires categorization
- Affects budget category when categorized

---

### Scenario 19: Transfer Between Savings Accounts
**Description:** Both accounts are savings (not checking)

**Input:**
```
OnPoint Savings:  Oct 28, -$500, "Transfer"
Ally Savings:     Oct 28, +$500, "Deposit"
```

**Expected:**
- Match detected (account type doesn't matter)
- Both converted to transfers
- No special handling (not CC payment)

---

### Scenario 20: Three-Way Confusion
**Description:** Three transactions with same amount on same day

**Input:**
```
Chase:    Oct 28, -$100
OnPoint:  Oct 28, +$100
Savings:  Oct 28, +$100
```

**Expected:**
- Two suggestions created:
  1. Chase ↔ OnPoint (score X)
  2. Chase ↔ Savings (score Y)
- User chooses correct match
- Other suggestion auto-rejected (Chase already linked)

---

## Test Cases

### Unit Tests: Matching Algorithm

#### Test 1: Exact Match (Same Day)
```go
func TestMatchScore_ExactMatch(t *testing.T) {
    txnA := Transaction{Date: oct28, Amount: -1000, AccountID: "chase"}
    txnB := Transaction{Date: oct28, Amount: 1000, AccountID: "onpoint"}

    score := calculateMatchScore(txnA, txnB)

    assert.GreaterOrEqual(t, score, 10)  // At least medium confidence
    assert.Equal(t, "high", classifyConfidence(score))
}
```

#### Test 2: Different Users (Should Not Match)
```go
func TestMatch_DifferentUsers_NoMatch(t *testing.T) {
    txnA := Transaction{UserID: "user1", Amount: -1000}
    txnB := Transaction{UserID: "user2", Amount: 1000}

    candidates := findCandidates(txnA)

    assert.NotContains(t, candidates, txnB)  // Privacy protection
}
```

#### Test 3: Same Account (Should Not Match)
```go
func TestMatch_SameAccount_NoMatch(t *testing.T) {
    txnA := Transaction{AccountID: "chase", Amount: -1000}
    txnB := Transaction{AccountID: "chase", Amount: 1000}

    candidates := findCandidates(txnA)

    assert.NotContains(t, candidates, txnB)
}
```

#### Test 4: Already Linked (Should Be Excluded)
```go
func TestMatch_AlreadyLinked_Excluded(t *testing.T) {
    txnA := Transaction{Type: "transfer", TransferToAccountID: "xyz"}

    candidates := findCandidates(txnA)

    assert.Empty(t, candidates)
}
```

#### Test 5: Date Window (Within 3 Days)
```go
func TestMatch_WithinDateWindow(t *testing.T) {
    txnA := Transaction{Date: oct28, Amount: -1000}
    txnB := Transaction{Date: oct30, Amount: 1000}  // 2 days later

    candidates := findCandidates(txnA)

    assert.Contains(t, candidates, txnB)
}
```

#### Test 6: Date Window (Outside 3 Days)
```go
func TestMatch_OutsideDateWindow(t *testing.T) {
    txnA := Transaction{Date: oct28, Amount: -1000}
    txnB := Transaction{Date: nov2, Amount: 1000}  // 5 days later

    candidates := findCandidates(txnA)

    assert.NotContains(t, candidates, txnB)
}
```

#### Test 7: Credit Card Payment Detection
```go
func TestMatch_CreditCardPayment_Detected(t *testing.T) {
    checking := Account{Type: "checking"}
    credit := Account{Type: "credit"}
    txnA := Transaction{AccountID: checking.ID, Amount: -500}
    txnB := Transaction{AccountID: credit.ID, Amount: 500}

    suggestion := createSuggestion(txnA, txnB)

    assert.True(t, suggestion.IsCreditPayment)
}
```

#### Test 8: Round Amount Scoring
```go
func TestMatchScore_RoundAmount_Boost(t *testing.T) {
    txnA := Transaction{Amount: -100000}  // $1,000.00
    txnB := Transaction{Amount: 100000}

    txnC := Transaction{Amount: -123748}  // $1,237.48
    txnD := Transaction{Amount: 123748}

    scoreRound := calculateMatchScore(txnA, txnB)
    scoreNotRound := calculateMatchScore(txnC, txnD)

    assert.Greater(t, scoreRound, scoreNotRound)
}
```

#### Test 9: Description Similarity
```go
func TestMatchScore_DescriptionMatch_Boost(t *testing.T) {
    txnA := Transaction{Description: "Transfer to OnPoint"}
    txnB := Transaction{Description: "Transfer from Chase"}

    score := calculateMatchScore(txnA, txnB)

    // Should get +5 for "transfer" keyword
    assert.GreaterOrEqual(t, score, 15)
}
```

#### Test 10: Confidence Classification
```go
func TestConfidence_Levels(t *testing.T) {
    assert.Equal(t, "high", classifyConfidence(15))
    assert.Equal(t, "medium", classifyConfidence(12))
    assert.Equal(t, "low", classifyConfidence(9))
}
```

---

### Integration Tests: Linking Flow

#### Test 11: Accept Suggestion (Normal Transfer)
```go
func TestAcceptSuggestion_NormalTransfer(t *testing.T) {
    // Setup: Create two normal transactions
    txnA := createTransaction(Type: "normal", Amount: -1000, Account: chase)
    txnB := createTransaction(Type: "normal", Amount: 1000, Account: onpoint)
    suggestion := createSuggestion(txnA.ID, txnB.ID)

    // Action: Accept suggestion
    err := acceptSuggestion(suggestion.ID)
    assert.NoError(t, err)

    // Verify: Both converted to transfers
    txnA = getTransaction(txnA.ID)
    txnB = getTransaction(txnB.ID)
    assert.Equal(t, "transfer", txnA.Type)
    assert.Equal(t, "transfer", txnB.Type)
    assert.Equal(t, txnB.AccountID, *txnA.TransferToAccountID)
    assert.Equal(t, txnA.AccountID, *txnB.TransferToAccountID)
    assert.Nil(t, txnA.CategoryID)
    assert.Nil(t, txnB.CategoryID)

    // Verify: Suggestion marked accepted
    suggestion = getSuggestion(suggestion.ID)
    assert.Equal(t, "accepted", suggestion.Status)
}
```

#### Test 12: Accept Suggestion (CC Payment)
```go
func TestAcceptSuggestion_CCPayment(t *testing.T) {
    // Setup
    checking := createAccount(Type: "checking")
    credit := createAccount(Type: "credit")
    paymentCat := createPaymentCategory(AccountID: credit.ID)
    allocateBudget(paymentCat.ID, 50000)  // $500

    txnA := createTransaction(Account: checking, Amount: -50000)
    txnB := createTransaction(Account: credit, Amount: 50000)
    suggestion := createSuggestion(txnA.ID, txnB.ID)

    // Action
    err := acceptSuggestion(suggestion.ID)
    assert.NoError(t, err)

    // Verify: Payment category applied to checking side
    txnA = getTransaction(txnA.ID)
    txnB = getTransaction(txnB.ID)
    assert.Equal(t, paymentCat.ID, *txnA.CategoryID)
    assert.Nil(t, txnB.CategoryID)
}
```

#### Test 13: Accept Suggestion (CC Overpayment)
```go
func TestAcceptSuggestion_CCOverpayment(t *testing.T) {
    // Setup: Only $200 available, but paying $250
    checking := createAccount(Type: "checking")
    credit := createAccount(Type: "credit")
    paymentCat := createPaymentCategory(AccountID: credit.ID)
    allocateBudget(paymentCat.ID, 50000)  // $500
    createPayment(paymentCat.ID, 30000)   // Already spent $300
    // Available: $200

    txnA := createTransaction(Account: checking, Amount: -25000)  // $250
    txnB := createTransaction(Account: credit, Amount: 25000)
    suggestion := createSuggestion(txnA.ID, txnB.ID)

    // Action
    err := acceptSuggestion(suggestion.ID)
    assert.NoError(t, err)

    // Verify: No category applied (overpayment)
    txnA = getTransaction(txnA.ID)
    assert.Nil(t, txnA.CategoryID)
}
```

#### Test 14: Reject Suggestion
```go
func TestRejectSuggestion(t *testing.T) {
    // Setup
    suggestion := createSuggestion(txnAID, txnBID)

    // Action
    err := rejectSuggestion(suggestion.ID)
    assert.NoError(t, err)

    // Verify: Suggestion marked rejected
    suggestion = getSuggestion(suggestion.ID)
    assert.Equal(t, "rejected", suggestion.Status)

    // Verify: Transactions unchanged
    txnA := getTransaction(txnAID)
    txnB := getTransaction(txnBID)
    assert.Equal(t, "normal", txnA.Type)
    assert.Equal(t, "normal", txnB.Type)
}
```

#### Test 15: Manual Link
```go
func TestManualLink(t *testing.T) {
    // Setup: Two old normal transactions
    txnA := createTransaction(Type: "normal", Amount: -1000)
    txnB := createTransaction(Type: "normal", Amount: 1000)

    // Action
    err := manualLink(txnA.ID, txnB.ID)
    assert.NoError(t, err)

    // Verify
    txnA = getTransaction(txnA.ID)
    txnB = getTransaction(txnB.ID)
    assert.Equal(t, "transfer", txnA.Type)
    assert.Equal(t, "transfer", txnB.Type)
}
```

#### Test 16: Manual Link Validation (Same Account)
```go
func TestManualLink_SameAccount_Error(t *testing.T) {
    txnA := createTransaction(AccountID: "chase", Amount: -1000)
    txnB := createTransaction(AccountID: "chase", Amount: 1000)

    err := manualLink(txnA.ID, txnB.ID)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "same account")
}
```

#### Test 17: Manual Link Validation (Already Linked)
```go
func TestManualLink_AlreadyLinked_Error(t *testing.T) {
    txnA := createTransaction(Type: "transfer", TransferToAccountID: "xyz")
    txnB := createTransaction(Type: "normal")

    err := manualLink(txnA.ID, txnB.ID)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "already linked")
}
```

---

### Integration Tests: Import Flow

#### Test 18: Import Creates Suggestions
```go
func TestImport_CreatesSuggestions(t *testing.T) {
    // Setup: Import Chase first
    chaseOFX := loadOFXFile("chase_oct.ofx")
    importResult := importOFX(chaseAccount.ID, chaseOFX)
    assert.Equal(t, 10, importResult.ImportedCount)

    // No suggestions yet (no matching account data)
    suggestions := listSuggestions()
    assert.Empty(t, suggestions)

    // Action: Import OnPoint
    onpointOFX := loadOFXFile("onpoint_oct.ofx")
    importResult = importOFX(onpointAccount.ID, onpointOFX)
    assert.Equal(t, 8, importResult.ImportedCount)

    // Verify: Suggestions created
    suggestions = listSuggestions()
    assert.NotEmpty(t, suggestions)
    assert.Equal(t, "high", suggestions[0].Confidence)
}
```

#### Test 19: Asynchronous Import (Days Apart)
```go
func TestImport_AsyncImport_StillMatches(t *testing.T) {
    // Day 1: Import Chase
    time.Sleep(0)  // Simulate Monday
    importOFX(chaseAccount.ID, chaseOFX)
    suggestions := listSuggestions()
    assert.Empty(t, suggestions)  // No OnPoint data yet

    // Day 5: Import OnPoint
    time.Sleep(4 * 24 * time.Hour)  // Simulate Friday
    importOFX(onpointAccount.ID, onpointOFX)

    // Verify: Match found despite import time gap
    suggestions = listSuggestions()
    assert.NotEmpty(t, suggestions)
}
```

#### Test 20: Duplicate Detection (Overlapping Imports)
```go
func TestImport_Duplicates_Skipped(t *testing.T) {
    // Import Oct 1-31
    ofx1 := loadOFXFile("chase_oct_full.ofx")
    result1 := importOFX(chaseAccount.ID, ofx1)
    assert.Equal(t, 20, result1.ImportedCount)

    // Import Oct 15-31 again (overlaps)
    ofx2 := loadOFXFile("chase_oct_partial.ofx")
    result2 := importOFX(chaseAccount.ID, ofx2)
    assert.Equal(t, 10, result2.ImportedCount)  // Only new transactions
    assert.Equal(t, 10, result2.SkippedDuplicates)
}
```

---

### User Acceptance Tests

#### Test 21: End-to-End Transfer Linking
```
1. User imports Chase transactions (Oct 1-31)
   → 15 transactions created

2. User imports OnPoint transactions (Oct 1-31)
   → 12 transactions created
   → Notification: "3 potential transfer matches found"

3. User clicks "Review Matches"
   → Sees 3 suggestions
   → 2 high confidence, 1 medium confidence

4. User reviews first suggestion (high confidence)
   → Chase -$1,000 ↔ OnPoint +$1,000
   → Same date, description match
   → Clicks "Link as Transfer"

5. System links transactions
   → Both converted to transfers
   → Success message shown
   → Removed from "Uncategorized" list

6. User checks budget
   → No impact on categories (correct)
   → Account balances correct
```

#### Test 22: Credit Card Payment Workflow
```
1. User imports checking transactions
   → $500 payment to credit card

2. User imports credit card transactions
   → $500 payment received
   → Notification: "1 credit card payment match found"

3. User reviews suggestion
   → Badge: "Credit Card Payment"
   → Info: "Will be categorized under CC Payment"
   → Clicks "Link as Payment"

4. System links with payment category
   → Checking: category='Chase CC Payment'
   → Credit: category=null

5. User checks budget
   → CC Payment category shows $500 activity
   → Available reduced by $500
   → Credit card balance reduced
```

#### Test 23: False Positive Rejection
```
1. User imports both accounts
   → Suggestion: Chase -$50 ↔ OnPoint +$50
   → Medium confidence (amount match only)

2. User recognizes these are NOT the same transfer
   → Chase: Grocery store
   → OnPoint: Employer reimbursement

3. User clicks "Not a Match"
   → Suggestion dismissed
   → Both remain as normal transactions

4. User categorizes separately
   → Chase: Groceries category
   → OnPoint: Income (optional)
```

---

## High-Level System Design

### Architecture Components

```
┌─────────────────────────────────────────────────────────────┐
│                         USER INTERFACE                       │
│  - Import Page (file upload)                                │
│  - Suggestions Review Page (accept/reject)                  │
│  - Transaction List (manual link mode)                      │
│  - Notification Badge (pending count)                       │
└────────────────────────┬────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────┐
│                      API ENDPOINTS                           │
│  POST /api/transactions/import                              │
│  GET  /api/transfer-suggestions                             │
│  POST /api/transfer-suggestions/{id}/accept                 │
│  POST /api/transfer-suggestions/{id}/reject                 │
│  POST /api/transactions/link                                │
└────────────────────────┬────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────┐
│                   APPLICATION SERVICES                       │
│                                                              │
│  ┌──────────────────────────────────────────────┐          │
│  │  ImportService                               │          │
│  │  - Parse OFX                                 │          │
│  │  - Deduplicate (FitID)                       │          │
│  │  - Create transactions                       │          │
│  │  - Call TransferMatcherService               │          │
│  └──────────────────────────────────────────────┘          │
│                         │                                    │
│                         v                                    │
│  ┌──────────────────────────────────────────────┐          │
│  │  TransferMatcherService                      │          │
│  │  - Find candidates                           │          │
│  │  - Score matches                             │          │
│  │  - Classify confidence                       │          │
│  │  - Create suggestions                        │          │
│  └──────────────────────────────────────────────┘          │
│                         │                                    │
│                         v                                    │
│  ┌──────────────────────────────────────────────┐          │
│  │  TransferLinkService                         │          │
│  │  - Accept suggestion                         │          │
│  │  - Reject suggestion                         │          │
│  │  - Manual link                               │          │
│  │  - Apply CC payment category                 │          │
│  │  - Convert to transfers                      │          │
│  └──────────────────────────────────────────────┘          │
│                                                              │
└────────────────────────┬────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────┐
│                   DOMAIN REPOSITORIES                        │
│  - TransactionRepository                                    │
│  - TransferSuggestionRepository                             │
│  - AccountRepository                                        │
│  - CategoryRepository                                       │
└────────────────────────┬────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────┐
│                       DATABASE                               │
│  Tables:                                                     │
│  - transactions                                             │
│  - transfer_match_suggestions                               │
│  - accounts                                                 │
│  - categories                                               │
└─────────────────────────────────────────────────────────────┘
```

---

### Data Flow: Import to Linking

```
┌─────────────────────────────────────────────────────────────┐
│ 1. USER UPLOADS OFX FILE                                     │
└────────────────────────┬────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────┐
│ 2. IMPORT SERVICE                                            │
│    - Parse OFX → extract transactions + ledger balance      │
│    - Check FitID duplicates → skip existing                 │
│    - Create new transactions (type='normal', no category)   │
│    - Update account balance                                 │
│    - Adjust Ready to Assign                                 │
│    Result: [txn_id_1, txn_id_2, ..., txn_id_n]             │
└────────────────────────┬────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────┐
│ 3. TRANSFER MATCHER SERVICE (auto-triggered)                │
│    For each imported transaction:                           │
│      - Query candidates (opposite amount, ±3 days)          │
│      - Score each candidate                                 │
│      - Filter: score >= 10 (minimum threshold)              │
│      - Create suggestion if match found                     │
│    Result: [suggestion_1, suggestion_2, ...]                │
└────────────────────────┬────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────┐
│ 4. SUGGESTIONS STORED IN DATABASE                            │
│    transfer_match_suggestions table                         │
│    Status: 'pending'                                        │
└────────────────────────┬────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────┐
│ 5. USER REVIEWS SUGGESTIONS                                  │
│    UI shows:                                                │
│    - Transaction A details                                  │
│    - Transaction B details                                  │
│    - Confidence level                                       │
│    - Account names (A → B)                                  │
│    - [Accept] [Reject] buttons                              │
└────────────────────────┬────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────┐
│ 6. USER ACCEPTS SUGGESTION                                   │
└────────────────────────┬────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────┐
│ 7. TRANSFER LINK SERVICE                                     │
│    Begin transaction:                                       │
│      - Load both transactions                               │
│      - Check if one is credit card                          │
│      - If CC: Apply payment category logic                  │
│      - Update txn A: type='transfer', add cross-ref         │
│      - Update txn B: type='transfer', add cross-ref         │
│      - Mark suggestion as 'accepted'                        │
│      - Delete other suggestions involving these txns        │
│    Commit transaction                                       │
└────────────────────────┬────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────┐
│ 8. RESULT                                                    │
│    - Both transactions now type='transfer'                  │
│    - Cross-references set                                   │
│    - Category applied (if CC payment)                       │
│    - Removed from uncategorized list                        │
│    - Budget reflects changes                                │
└─────────────────────────────────────────────────────────────┘
```

---

### Database Schema

#### Existing: `transactions` Table
```sql
CREATE TABLE transactions (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL DEFAULT 'normal' CHECK(type IN ('normal', 'transfer')),
    account_id TEXT NOT NULL,
    transfer_to_account_id TEXT,              -- Links transfers
    category_id TEXT,
    amount INTEGER NOT NULL,                  -- In cents
    description TEXT,
    date DATETIME NOT NULL,
    fitid TEXT,                               -- OFX duplicate detection
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    FOREIGN KEY (transfer_to_account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);

CREATE INDEX idx_transactions_account ON transactions(account_id);
CREATE INDEX idx_transactions_date ON transactions(date);
CREATE INDEX idx_transactions_transfer ON transactions(transfer_to_account_id);
```

#### New: `transfer_match_suggestions` Table
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

    UNIQUE(transaction_a_id, transaction_b_id)
);

CREATE INDEX idx_suggestions_status ON transfer_match_suggestions(status);
CREATE INDEX idx_suggestions_confidence ON transfer_match_suggestions(confidence);
CREATE INDEX idx_suggestions_credit ON transfer_match_suggestions(is_credit_payment);
```

#### New: Performance Index
```sql
-- Optimize candidate searches
CREATE INDEX idx_transactions_matching ON transactions(type, amount, date)
    WHERE transfer_to_account_id IS NULL;
```

---

### Matching Algorithm (Detailed)

#### Query: Find Candidates
```sql
SELECT t.*
FROM transactions t
JOIN accounts a ON t.account_id = a.id
WHERE a.user_id = ?                          -- Same user (privacy)
  AND t.account_id != ?                      -- Different account
  AND t.type = 'normal'                      -- Not already a transfer
  AND t.transfer_to_account_id IS NULL       -- Not already linked
  AND t.amount = ?                           -- Opposite amount (-X vs +X)
  AND t.date >= ?                            -- Date window start (txn.date - 3 days)
  AND t.date <= ?                            -- Date window end (txn.date + 3 days)
LIMIT 10;                                    -- Limit to prevent too many suggestions
```

#### Scoring Function
```go
func calculateMatchScore(txnA, txnB Transaction) int {
    score := 0

    // Date proximity
    daysDiff := abs(txnA.Date.Sub(txnB.Date).Hours() / 24)
    if daysDiff == 0 {
        score += 10  // Same day
    } else if daysDiff == 1 {
        score += 5   // 1 day apart
    } else if daysDiff <= 3 {
        score += 2   // 2-3 days apart
    }

    // Description similarity
    if containsTransferKeywords(txnA.Description) && containsTransferKeywords(txnB.Description) {
        score += 5
    }

    // Round amount
    if isRoundAmount(txnA.Amount) {
        score += 3
    }

    // Credit card payment
    accountA := getAccount(txnA.AccountID)
    accountB := getAccount(txnB.AccountID)
    if accountA.Type == "credit" || accountB.Type == "credit" {
        score += 5
        if containsPaymentKeywords(txnA.Description) || containsPaymentKeywords(txnB.Description) {
            score += 3
        }
    }

    return score
}

func containsTransferKeywords(desc string) bool {
    keywords := []string{"transfer", "xfer", "from", "to"}
    descLower := strings.ToLower(desc)
    for _, kw := range keywords {
        if strings.Contains(descLower, kw) {
            return true
        }
    }
    return false
}

func containsPaymentKeywords(desc string) bool {
    keywords := []string{"payment", "autopay", "pay"}
    descLower := strings.ToLower(desc)
    for _, kw := range keywords {
        if strings.Contains(descLower, kw) {
            return true
        }
    }
    return false
}

func isRoundAmount(amount int64) bool {
    return amount%100 == 0  // Ends in .00
}
```

#### Confidence Classification
```go
func classifyConfidence(score int) string {
    if score >= 15 {
        return "high"
    } else if score >= 10 {
        return "medium"
    } else {
        return "low"
    }
}
```

---

### Credit Card Payment Logic

#### Apply Payment Category
```go
func applyCCPaymentCategory(checkingTxn, creditTxn *Transaction) *string {
    // Get payment category for this credit card
    paymentCat, err := categoryRepo.GetPaymentCategoryByAccountID(creditTxn.AccountID)
    if err != nil || paymentCat == nil {
        return nil  // No payment category exists
    }

    // Calculate available budget in payment category
    available := calculateAvailableInCategory(paymentCat.ID)

    // Only categorize if payment <= available
    paymentAmount := abs(checkingTxn.Amount)
    if available >= paymentAmount {
        return &paymentCat.ID
    }

    return nil  // Overpayment - don't categorize
}

func calculateAvailableInCategory(categoryID string) int64 {
    // Sum all allocations to this category
    allocations := allocationRepo.ListByCategory(categoryID)
    totalAllocated := sum(allocations)

    // Sum all spending in this category (negative amounts)
    transactions := transactionRepo.ListByCategory(categoryID)
    totalSpent := 0
    for _, txn := range transactions {
        if txn.Amount < 0 {
            totalSpent += abs(txn.Amount)
        }
    }

    return totalAllocated - totalSpent
}
```

---

### Performance Considerations

#### Query Optimization
- **Index on (type, amount, date)**: Speeds up candidate search
- **WHERE clause filtering**: Reduces result set before sorting
- **LIMIT**: Prevents excessive suggestions

#### Scalability
- **Small import** (10 txns): <100ms
- **Medium import** (100 txns): <1s
- **Large import** (500 txns): <5s
- **Very large import** (1000+ txns): Consider async job

#### Optimization Strategies
1. **Batch processing**: Find all candidates once, match in-memory
2. **Time window**: Only search last 90 days (configurable)
3. **Async processing** (Phase 4): Background job for large imports
4. **Caching**: Cache account/user data during matching

---

### Configuration

```go
type TransferMatchConfig struct {
    Enabled              bool    `json:"enabled"`                // Feature flag
    MaxDateDiffDays      int     `json:"max_date_diff_days"`     // Default: 3
    MinConfidenceScore   int     `json:"min_confidence_score"`   // Default: 10
    AmountTolerancePct   float64 `json:"amount_tolerance_pct"`   // Default: 0.0 (Phase 1)
    AmountToleranceCents int64   `json:"amount_tolerance_cents"` // Default: 0 (Phase 1)
    AutoLinkHighConf     bool    `json:"auto_link_high_conf"`    // Default: false
    MaxSuggestionsPerImport int  `json:"max_suggestions"`        // Default: 20
}
```

**Phase 1 Defaults:**
```json
{
  "enabled": true,
  "max_date_diff_days": 3,
  "min_confidence_score": 10,
  "amount_tolerance_pct": 0.0,
  "amount_tolerance_cents": 0,
  "auto_link_high_conf": false,
  "max_suggestions": 20
}
```

---

## Summary

### What We're Building
A **semi-automatic transaction linking system** that:
1. Detects potential transfer matches during import
2. Presents suggestions to users for confirmation
3. Links approved transactions as transfers
4. Handles credit card payments with special category logic
5. Supports manual linking for missed matches

### Key Principles
- **User confirmation required** (safety first)
- **Credit card payments are special** (payment category tracking)
- **Works asynchronously** (import order doesn't matter)
- **Exact matches only in Phase 1** (no tolerance for fees yet)
- **Privacy protected** (user-scoped queries only)

### Success Criteria
- ✅ High confidence suggestions accepted >70% of time
- ✅ False positive rate <30%
- ✅ Performance: <1s for typical imports
- ✅ No budget calculation errors
- ✅ User satisfaction with feature

---

**End of Specification**
