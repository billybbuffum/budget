#!/bin/bash
API_URL="http://localhost:8080/api"

# Get current suggestion ID
SUGGESTION_ID=$(curl -s "$API_URL/transfer-suggestions" | jq -r '.suggestions[0].id')

echo "Suggestion ID: $SUGGESTION_ID"

if [ "$SUGGESTION_ID" = "null" ] || [ -z "$SUGGESTION_ID" ]; then
    echo "No suggestions found. Creating new ones..."

    # Create accounts
    CHASE=$(curl -s -X POST "$API_URL/accounts" -H "Content-Type: application/json" -d '{"name": "Test A", "type": "checking", "balance": 1000000}')
    CHASE_ID=$(echo "$CHASE" | jq -r '.id')

    ONPOINT=$(curl -s -X POST "$API_URL/accounts" -H "Content-Type: application/json" -d '{"name": "Test B", "type": "checking", "balance": 700000}')
    ONPOINT_ID=$(echo "$ONPOINT" | jq -r '.id')

    # Import
    curl -s -X POST "$API_URL/transactions/import" -F "account_id=$CHASE_ID" -F "file=@test_data/chase_transfer_out.ofx" > /dev/null
    curl -s -X POST "$API_URL/transactions/import" -F "account_id=$ONPOINT_ID" -F "file=@test_data/onpoint_transfer_in.ofx" > /dev/null

    # Get new suggestion
    SUGGESTION_ID=$(curl -s "$API_URL/transfer-suggestions" | jq -r '.suggestions[0].id')
    echo "New suggestion ID: $SUGGESTION_ID"
fi

echo ""
echo "Accepting suggestion..."
RESULT=$(curl -s -X POST "$API_URL/transfer-suggestions/$SUGGESTION_ID/accept")
echo "$RESULT" | jq '.'
