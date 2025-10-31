#!/bin/bash

set -e

API_URL="http://localhost:8080/api"

echo "=== Transaction Linking Feature Test ==="
echo ""

# Helper function to make API calls
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3

    if [ -z "$data" ]; then
        curl -s -X "$method" "$API_URL$endpoint"
    else
        curl -s -X "$method" "$API_URL$endpoint" -H "Content-Type: application/json" -d "$data"
    fi
}

# Helper function to upload OFX file
upload_ofx() {
    local account_id=$1
    local file_path=$2

    curl -s -X POST "$API_URL/transactions/import" \
        -F "account_id=$account_id" \
        -F "file=@$file_path"
}

echo "Step 1: Creating test accounts..."
echo ""

# Create Chase Checking account (balance in cents: $10,000.00 = 1000000)
CHASE_RESPONSE=$(api_call POST /accounts '{
    "name": "Chase Checking",
    "type": "checking",
    "balance": 1000000
}')
CHASE_ID=$(echo "$CHASE_RESPONSE" | jq -r '.id')
echo "✓ Created Chase Checking: $CHASE_ID"

# Create OnPoint Checking account (balance in cents: $7,000.00 = 700000)
ONPOINT_RESPONSE=$(api_call POST /accounts '{
    "name": "OnPoint Checking",
    "type": "checking",
    "balance": 700000
}')
ONPOINT_ID=$(echo "$ONPOINT_RESPONSE" | jq -r '.id')
echo "✓ Created OnPoint Checking: $ONPOINT_ID"

# Create Savings account (balance in cents: $3,000.00 = 300000)
SAVINGS_RESPONSE=$(api_call POST /accounts '{
    "name": "Savings Account",
    "type": "savings",
    "balance": 300000
}')
SAVINGS_ID=$(echo "$SAVINGS_RESPONSE" | jq -r '.id')
echo "✓ Created Savings Account: $SAVINGS_ID"

# Create Visa Credit Card (balance in cents: -$1,000.00 = -100000)
VISA_RESPONSE=$(api_call POST /accounts '{
    "name": "Visa Credit Card",
    "type": "credit",
    "balance": -100000
}')
VISA_ID=$(echo "$VISA_RESPONSE" | jq -r '.id')
echo "✓ Created Visa Credit Card: $VISA_ID"

echo ""
echo "Step 2: Testing Scenario 1 - Same Day Transfer Detection"
echo ""

# Import Chase transfer out
echo "Importing Chase transfer out ($1000)..."
IMPORT1=$(upload_ofx "$CHASE_ID" "test_data/chase_transfer_out.ofx")
echo "$IMPORT1" | jq '.'

# Import OnPoint transfer in
echo "Importing OnPoint transfer in ($1000)..."
IMPORT2=$(upload_ofx "$ONPOINT_ID" "test_data/onpoint_transfer_in.ofx")
echo "$IMPORT2" | jq '.'

# Check for suggestions
echo ""
echo "Checking for transfer suggestions..."
SUGGESTIONS=$(api_call GET /transfer-suggestions)
echo "$SUGGESTIONS" | jq '.'

NUM_SUGGESTIONS=$(echo "$SUGGESTIONS" | jq '.suggestions | length')
echo ""
echo "Found $NUM_SUGGESTIONS suggestion(s)"

if [ "$NUM_SUGGESTIONS" -gt 0 ]; then
    echo "✓ PASS: Same-day transfer detected!"

    # Check confidence level
    CONFIDENCE=$(echo "$SUGGESTIONS" | jq -r '.suggestions[0].confidence')
    SCORE=$(echo "$SUGGESTIONS" | jq -r '.suggestions[0].score')
    echo "  Confidence: $CONFIDENCE (Score: $SCORE)"

    # Check if it's NOT marked as credit payment
    IS_CC=$(echo "$SUGGESTIONS" | jq -r '.suggestions[0].is_credit_payment')
    if [ "$IS_CC" = "false" ]; then
        echo "✓ Correctly identified as regular transfer (not CC payment)"
    else
        echo "✗ FAIL: Incorrectly marked as credit card payment"
    fi

    # Test accepting the suggestion
    SUGGESTION_ID=$(echo "$SUGGESTIONS" | jq -r '.suggestions[0].id')
    echo ""
    echo "Testing suggestion acceptance..."
    ACCEPT_RESULT=$(api_call POST "/transfer-suggestions/$SUGGESTION_ID/accept" "")
    echo "$ACCEPT_RESULT" | jq '.'

    # Verify transactions are now linked
    LINKED_TXNS=$(echo "$ACCEPT_RESULT" | jq '.linked_transactions')
    if [ "$LINKED_TXNS" != "null" ]; then
        echo "✓ PASS: Transactions successfully linked!"
    else
        echo "✗ FAIL: Transactions not linked"
    fi
else
    echo "✗ FAIL: No suggestions created for same-day transfer"
fi

echo ""
echo "========================================="
echo ""
echo "Step 3: Testing Scenario 2 - Delayed Transfer (1 day apart)"
echo ""

# Import Chase delayed transfer out
echo "Importing Chase delayed transfer out ($500, Oct 26)..."
IMPORT3=$(upload_ofx "$CHASE_ID" "test_data/chase_delayed_out.ofx")
echo "$IMPORT3" | jq '.'

# Import Savings delayed transfer in
echo "Importing Savings delayed transfer in ($500, Oct 27)..."
IMPORT4=$(upload_ofx "$SAVINGS_ID" "test_data/savings_delayed_in.ofx")
echo "$IMPORT4" | jq '.'

# Check for new suggestions
echo ""
echo "Checking for transfer suggestions..."
SUGGESTIONS2=$(api_call GET /transfer-suggestions?status=pending)
echo "$SUGGESTIONS2" | jq '.'

NUM_SUGGESTIONS2=$(echo "$SUGGESTIONS2" | jq '.suggestions | length')
echo ""
echo "Found $NUM_SUGGESTIONS2 pending suggestion(s)"

if [ "$NUM_SUGGESTIONS2" -gt 0 ]; then
    echo "✓ PASS: Delayed transfer detected!"

    CONFIDENCE2=$(echo "$SUGGESTIONS2" | jq -r '.suggestions[0].confidence')
    SCORE2=$(echo "$SUGGESTIONS2" | jq -r '.suggestions[0].score')
    echo "  Confidence: $CONFIDENCE2 (Score: $SCORE2)"

    # Test rejecting this suggestion
    SUGGESTION_ID2=$(echo "$SUGGESTIONS2" | jq -r '.suggestions[0].id')
    echo ""
    echo "Testing suggestion rejection..."
    REJECT_RESULT=$(api_call POST "/transfer-suggestions/$SUGGESTION_ID2/reject" "")
    echo "$REJECT_RESULT" | jq '.'

    if echo "$REJECT_RESULT" | jq -e '.success' > /dev/null; then
        echo "✓ PASS: Suggestion rejected successfully!"
    else
        echo "✗ FAIL: Failed to reject suggestion"
    fi
else
    echo "✗ FAIL: No suggestions for delayed transfer"
fi

echo ""
echo "========================================="
echo ""
echo "Step 4: Testing Scenario 5 - Credit Card Payment Detection"
echo ""

# Import Chase CC payment
echo "Importing Chase CC payment ($750)..."
IMPORT5=$(upload_ofx "$CHASE_ID" "test_data/chase_cc_payment.ofx")
echo "$IMPORT5" | jq '.'

# Import Visa payment received
echo "Importing Visa payment received ($750)..."
IMPORT6=$(upload_ofx "$VISA_ID" "test_data/visa_cc_payment_received.ofx")
echo "$IMPORT6" | jq '.'

# Check for new suggestions
echo ""
echo "Checking for transfer suggestions..."
SUGGESTIONS3=$(api_call GET /transfer-suggestions?status=pending)
echo "$SUGGESTIONS3" | jq '.'

NUM_SUGGESTIONS3=$(echo "$SUGGESTIONS3" | jq '.suggestions | length')
echo ""
echo "Found $NUM_SUGGESTIONS3 pending suggestion(s)"

if [ "$NUM_SUGGESTIONS3" -gt 0 ]; then
    echo "✓ PASS: Credit card payment detected!"

    # Check if marked as credit payment
    IS_CC3=$(echo "$SUGGESTIONS3" | jq -r '.suggestions[0].is_credit_payment')
    if [ "$IS_CC3" = "true" ]; then
        echo "✓ PASS: Correctly marked as credit card payment!"
    else
        echo "✗ FAIL: Not marked as credit card payment"
    fi

    CONFIDENCE3=$(echo "$SUGGESTIONS3" | jq -r '.suggestions[0].confidence')
    SCORE3=$(echo "$SUGGESTIONS3" | jq -r '.suggestions[0].score')
    echo "  Confidence: $CONFIDENCE3 (Score: $SCORE3)"
else
    echo "✗ FAIL: No suggestions for credit card payment"
fi

echo ""
echo "========================================="
echo ""
echo "Step 5: Testing Manual Linking"
echo ""

# Get all transactions
echo "Fetching all transactions..."
ALL_TXNS=$(api_call GET /transactions)

# Get two unlinked transactions (if any exist)
TXN_A_ID=$(echo "$ALL_TXNS" | jq -r '.transactions[0].id // empty')
TXN_B_ID=$(echo "$ALL_TXNS" | jq -r '.transactions[1].id // empty')

if [ -n "$TXN_A_ID" ] && [ -n "$TXN_B_ID" ]; then
    echo "Testing manual link with transactions: $TXN_A_ID and $TXN_B_ID"

    MANUAL_LINK=$(api_call POST /transactions/link "{
        \"transaction_a_id\": \"$TXN_A_ID\",
        \"transaction_b_id\": \"$TXN_B_ID\"
    }")

    echo "$MANUAL_LINK" | jq '.'

    if echo "$MANUAL_LINK" | jq -e '.success' > /dev/null 2>&1; then
        echo "✓ PASS: Manual linking works!"
    else
        echo "Note: Manual linking may have failed due to validation (expected for non-matching transactions)"
    fi
else
    echo "Skipping manual link test - not enough transactions"
fi

echo ""
echo "========================================="
echo ""
echo "Step 6: Final Summary"
echo ""

# Get final counts
FINAL_SUGGESTIONS=$(api_call GET /transfer-suggestions)
PENDING_COUNT=$(echo "$FINAL_SUGGESTIONS" | jq '.suggestions | map(select(.status == "pending")) | length')
TOTAL_COUNT=$(echo "$FINAL_SUGGESTIONS" | jq '.suggestions | length')

echo "Total suggestions created: $TOTAL_COUNT"
echo "Pending suggestions: $PENDING_COUNT"

FINAL_TXNS=$(api_call GET /transactions)
TOTAL_TXNS=$(echo "$FINAL_TXNS" | jq '.transactions | length')
LINKED_TXNS=$(echo "$FINAL_TXNS" | jq '.transactions | map(select(.type == "transfer")) | length')

echo "Total transactions: $TOTAL_TXNS"
echo "Linked transactions: $LINKED_TXNS"

echo ""
echo "=== Test Complete ==="
