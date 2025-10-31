#!/bin/bash

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

echo "=== Step 1: Creating test accounts ==="
echo ""

# Create Chase Checking account (balance in cents: $10,000.00 = 1000000)
CHASE_RESPONSE=$(api_call POST /accounts '{"name": "Chase Checking", "type": "checking", "balance": 1000000}')
CHASE_ID=$(echo "$CHASE_RESPONSE" | jq -r '.id')
echo "✓ Created Chase Checking: $CHASE_ID"

# Create OnPoint Checking account (balance in cents: $7,000.00 = 700000)
ONPOINT_RESPONSE=$(api_call POST /accounts '{"name": "OnPoint Checking", "type": "checking", "balance": 700000}')
ONPOINT_ID=$(echo "$ONPOINT_RESPONSE" | jq -r '.id')
echo "✓ Created OnPoint Checking: $ONPOINT_ID"

# Create Savings account (balance in cents: $3,000.00 = 300000)
SAVINGS_RESPONSE=$(api_call POST /accounts '{"name": "Savings Account", "type": "savings", "balance": 300000}')
SAVINGS_ID=$(echo "$SAVINGS_RESPONSE" | jq -r '.id')
echo "✓ Created Savings Account: $SAVINGS_ID"

# Create Visa Credit Card (balance in cents: -$1,000.00 = -100000)
VISA_RESPONSE=$(api_call POST /accounts '{"name": "Visa Credit Card", "type": "credit", "balance": -100000}')
VISA_ID=$(echo "$VISA_RESPONSE" | jq -r '.id')
echo "✓ Created Visa Credit Card: $VISA_ID"

echo ""
echo "=== Step 2: Testing Scenario 1 - Same Day Transfer Detection ==="
echo ""

# Import Chase transfer out
echo "Importing Chase transfer out (\$1000 debit on Oct 28)..."
upload_ofx "$CHASE_ID" "test_data/chase_transfer_out.ofx" | jq '.'

# Import OnPoint transfer in
echo ""
echo "Importing OnPoint transfer in (\$1000 credit on Oct 28)..."
upload_ofx "$ONPOINT_ID" "test_data/onpoint_transfer_in.ofx" | jq '.'

# Check for suggestions
echo ""
echo "Checking for transfer suggestions..."
SUGGESTIONS=$(api_call GET /transfer-suggestions)
NUM_SUGGESTIONS=$(echo "$SUGGESTIONS" | jq '.suggestions | length')

echo "Found $NUM_SUGGESTIONS suggestion(s)"
echo ""

if [ "$NUM_SUGGESTIONS" -gt 0 ]; then
    echo "✓ PASS: Same-day transfer detected!"
    echo ""

    # Show the first suggestion details
    SUGGESTION=$(echo "$SUGGESTIONS" | jq '.suggestions[0]')
    SUGGESTION_ID=$(echo "$SUGGESTION" | jq -r '.id')
    CONFIDENCE=$(echo "$SUGGESTION" | jq -r '.confidence')
    SCORE=$(echo "$SUGGESTION" | jq -r '.score')
    IS_CC=$(echo "$SUGGESTION" | jq -r '.is_credit_payment')

    echo "Suggestion Details:"
    echo "  ID: $SUGGESTION_ID"
    echo "  Confidence: $CONFIDENCE"
    echo "  Score: $SCORE"
    echo "  Is Credit Payment: $IS_CC"

    if [ "$IS_CC" = "false" ]; then
        echo "✓ Correctly identified as regular transfer (not CC payment)"
    else
        echo "✗ FAIL: Incorrectly marked as credit card payment"
    fi

    # Test accepting the suggestion
    echo ""
    echo "Testing suggestion acceptance..."
    ACCEPT_RESULT=$(curl -s -X POST "$API_URL/transfer-suggestions/$SUGGESTION_ID/accept")

    # Check if response contains "success"
    if echo "$ACCEPT_RESULT" | grep -q '"success":true'; then
        echo "✓ PASS: Suggestion accepted successfully!"
        echo "$ACCEPT_RESULT" | jq '.'
    else
        echo "✗ FAIL: Failed to accept suggestion"
        echo "Response: $ACCEPT_RESULT"
    fi
else
    echo "✗ FAIL: No suggestions created for same-day transfer"
fi

echo ""
echo "=== Step 3: Testing Scenario 2 - Delayed Transfer (1 day apart) ==="
echo ""

# Import Chase delayed transfer out
echo "Importing Chase delayed transfer out (\$500 debit on Oct 26)..."
upload_ofx "$CHASE_ID" "test_data/chase_delayed_out.ofx" | jq '.'

# Import Savings delayed transfer in
echo ""
echo "Importing Savings delayed transfer in (\$500 credit on Oct 27)..."
upload_ofx "$SAVINGS_ID" "test_data/savings_delayed_in.ofx" | jq '.'

# Check for new suggestions
echo ""
echo "Checking for transfer suggestions..."
SUGGESTIONS2=$(api_call GET /transfer-suggestions?status=pending)
NUM_SUGGESTIONS2=$(echo "$SUGGESTIONS2" | jq '.suggestions | length')

echo "Found $NUM_SUGGESTIONS2 pending suggestion(s)"
echo ""

if [ "$NUM_SUGGESTIONS2" -gt 0 ]; then
    echo "✓ PASS: Delayed transfer detected!"

    SUGGESTION2=$(echo "$SUGGESTIONS2" | jq '.suggestions[0]')
    CONFIDENCE2=$(echo "$SUGGESTION2" | jq -r '.confidence')
    SCORE2=$(echo "$SUGGESTION2" | jq -r '.score')

    echo "  Confidence: $CONFIDENCE2"
    echo "  Score: $SCORE2"
else
    echo "✗ FAIL: No suggestions for delayed transfer"
fi

echo ""
echo "=== Step 4: Testing Scenario 3 - Credit Card Payment Detection ==="
echo ""

# Import Chase CC payment
echo "Importing Chase CC payment (\$750 debit on Oct 29)..."
upload_ofx "$CHASE_ID" "test_data/chase_cc_payment.ofx" | jq '.'

# Import Visa payment received
echo ""
echo "Importing Visa payment received (\$750 credit on Oct 29)..."
upload_ofx "$VISA_ID" "test_data/visa_cc_payment_received.ofx" | jq '.'

# Check for new suggestions
echo ""
echo "Checking for transfer suggestions..."
SUGGESTIONS3=$(api_call GET /transfer-suggestions?status=pending)
NUM_SUGGESTIONS3=$(echo "$SUGGESTIONS3" | jq '.suggestions | length')

echo "Found $NUM_SUGGESTIONS3 pending suggestion(s)"
echo ""

if [ "$NUM_SUGGESTIONS3" -gt 0 ]; then
    # Look for the credit card payment suggestion
    CC_SUGGESTION=$(echo "$SUGGESTIONS3" | jq '.suggestions[] | select(.is_credit_payment == true)' | head -1)

    if [ -n "$CC_SUGGESTION" ]; then
        echo "✓ PASS: Credit card payment detected!"

        IS_CC3=$(echo "$CC_SUGGESTION" | jq -r '.is_credit_payment')
        CONFIDENCE3=$(echo "$CC_SUGGESTION" | jq -r '.confidence')
        SCORE3=$(echo "$CC_SUGGESTION" | jq -r '.score')

        echo "  Is Credit Payment: $IS_CC3"
        echo "  Confidence: $CONFIDENCE3"
        echo "  Score: $SCORE3"

        if [ "$IS_CC3" = "true" ]; then
            echo "✓ PASS: Correctly marked as credit card payment!"
        fi
    else
        echo "✗ FAIL: Credit card payment not detected"
    fi
else
    echo "✗ FAIL: No suggestions for credit card payment"
fi

echo ""
echo "=== Step 5: Final Summary ==="
echo ""

# Get final counts
FINAL_SUGGESTIONS=$(api_call GET /transfer-suggestions)
TOTAL_COUNT=$(echo "$FINAL_SUGGESTIONS" | jq '.suggestions | length')
PENDING_COUNT=$(echo "$FINAL_SUGGESTIONS" | jq '[.suggestions[] | select(.status == "pending")] | length')
ACCEPTED_COUNT=$(echo "$FINAL_SUGGESTIONS" | jq '[.suggestions[] | select(.status == "accepted")] | length')

echo "Total suggestions created: $TOTAL_COUNT"
echo "Pending suggestions: $PENDING_COUNT"
echo "Accepted suggestions: $ACCEPTED_COUNT"

FINAL_TXNS=$(api_call GET /transactions)
TOTAL_TXNS=$(echo "$FINAL_TXNS" | jq '.transactions | length')
LINKED_TXNS=$(echo "$FINAL_TXNS" | jq '[.transactions[] | select(.type == "transfer")] | length')

echo "Total transactions: $TOTAL_TXNS"
echo "Linked transactions: $LINKED_TXNS"

echo ""
echo "=== Test Complete ==="
