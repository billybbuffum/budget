#!/bin/bash

API_URL="http://localhost:8080/api"

echo "=== Testing Suggestion Accept Flow ==="
echo ""

# Create two accounts
echo "Creating accounts..."
CHASE=$(curl -s -X POST "$API_URL/accounts" -H "Content-Type: application/json" -d '{"name": "Chase", "type": "checking", "balance": 1000000}')
CHASE_ID=$(echo "$CHASE" | jq -r '.id')
echo "Chase ID: $CHASE_ID"

ONPOINT=$(curl -s -X POST "$API_URL/accounts" -H "Content-Type: application/json" -d '{"name": "OnPoint", "type": "checking", "balance": 700000}')
ONPOINT_ID=$(echo "$ONPOINT" | jq -r '.id')
echo "OnPoint ID: $ONPOINT_ID"

echo ""
echo "Importing transactions..."

# Import first transaction
curl -s -X POST "$API_URL/transactions/import" \
    -F "account_id=$CHASE_ID" \
    -F "file=@test_data/chase_transfer_out.ofx" | jq '.'

echo ""

# Import second transaction
curl -s -X POST "$API_URL/transactions/import" \
    -F "account_id=$ONPOINT_ID" \
    -F "file=@test_data/onpoint_transfer_in.ofx" | jq '.'

echo ""
echo "Fetching suggestions..."
SUGGESTIONS=$(curl -s "$API_URL/transfer-suggestions")
echo "$SUGGESTIONS" | jq '.'

echo ""
NUM=$(echo "$SUGGESTIONS" | jq '.suggestions | length')
echo "Found $NUM suggestion(s)"

if [ "$NUM" -gt 0 ]; then
    SUGGESTION_ID=$(echo "$SUGGESTIONS" | jq -r '.suggestions[0].id')
    STATUS=$(echo "$SUGGESTIONS" | jq -r '.suggestions[0].status')

    echo ""
    echo "First suggestion:"
    echo "  ID: $SUGGESTION_ID"
    echo "  Status: $STATUS"

    echo ""
    echo "Attempting to accept suggestion..."
    ACCEPT_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$API_URL/transfer-suggestions/$SUGGESTION_ID/accept")

    # Extract status code
    HTTP_STATUS=$(echo "$ACCEPT_RESPONSE" | grep "HTTP_STATUS" | cut -d: -f2)
    RESPONSE_BODY=$(echo "$ACCEPT_RESPONSE" | sed '/HTTP_STATUS/d')

    echo "HTTP Status: $HTTP_STATUS"
    echo "Response:"
    echo "$RESPONSE_BODY" | jq '.' 2>/dev/null || echo "$RESPONSE_BODY"

    if [ "$HTTP_STATUS" = "200" ]; then
        echo ""
        echo "✓ SUCCESS: Suggestion accepted!"
    else
        echo ""
        echo "✗ FAIL: Failed to accept suggestion"
    fi
else
    echo "✗ FAIL: No suggestions found"
fi

echo ""
echo "=== Test Complete ==="
