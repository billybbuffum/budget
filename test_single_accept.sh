#!/bin/bash

API_URL="http://localhost:8080/api"

echo "Creating fresh accounts and testing accept..."
echo ""

# Create accounts
CHASE=$(curl -s -X POST "$API_URL/accounts" -H "Content-Type: application/json" -d '{"name": "Test Chase", "type": "checking", "balance": 1000000}')
CHASE_ID=$(echo "$CHASE" | jq -r '.id')

ONPOINT=$(curl -s -X POST "$API_URL/accounts" -H "Content-Type: application/json" -d '{"name": "Test OnPoint", "type": "checking", "balance": 700000}')
ONPOINT_ID=$(echo "$ONPOINT" | jq -r '.id')

echo "Chase: $CHASE_ID"
echo "OnPoint: $ONPOINT_ID"
echo ""

# Import transactions
echo "Importing transactions..."
curl -s -X POST "$API_URL/transactions/import" -F "account_id=$CHASE_ID" -F "file=@test_data/chase_transfer_out.ofx" > /dev/null
curl -s -X POST "$API_URL/transactions/import" -F "account_id=$ONPOINT_ID" -F "file=@test_data/onpoint_transfer_in.ofx" > /dev/null

echo "Done"
echo ""

# Get suggestion
echo "Fetching suggestion..."
SUGGESTION=$(curl -s "$API_URL/transfer-suggestions" | jq '.suggestions[0]')
SUGGESTION_ID=$(echo "$SUGGESTION" | jq -r '.id')
STATUS=$(echo "$SUGGESTION" | jq -r '.status')

echo "Suggestion ID: $SUGGESTION_ID"
echo "Status: $STATUS"
echo ""

# Try to accept
echo "Accepting suggestion $SUGGESTION_ID..."
RESPONSE=$(curl -s -X POST "$API_URL/transfer-suggestions/$SUGGESTION_ID/accept")
echo "Response: $RESPONSE"
