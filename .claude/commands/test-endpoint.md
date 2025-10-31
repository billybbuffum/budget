---
description: Test an API endpoint with curl and verify the response
argument-hint: <endpoint-path>
allowed-tools: [Bash]
---

# Test API Endpoint

Test an API endpoint of the Budget application.

## Endpoint to Test
{{arg}}

## Testing Steps

### 1. Check if Server is Running
```bash
curl -s http://localhost:8080/health || echo "Server not running. Start with: docker-compose up -d"
```

### 2. Identify Endpoint Type

Based on the endpoint, determine the HTTP method and required data:

**Account Endpoints:**
- `POST /api/accounts` - Create account
- `GET /api/accounts` - List accounts
- `GET /api/accounts/{id}` - Get account
- `PUT /api/accounts/{id}` - Update account
- `DELETE /api/accounts/{id}` - Delete account

**Category Endpoints:**
- `POST /api/categories` - Create category
- `GET /api/categories` - List categories
- `GET /api/categories/{id}` - Get category
- `PUT /api/categories/{id}` - Update category

**Transaction Endpoints:**
- `POST /api/transactions` - Create transaction
- `GET /api/transactions` - List transactions
- `GET /api/transactions/{id}` - Get transaction

**Allocation Endpoints:**
- `POST /api/allocations` - Create/update allocation
- `GET /api/allocations/summary?period=YYYY-MM` - Get summary
- `GET /api/allocations/ready-to-assign` - Get available to assign

### 3. Run Test Command

**For GET requests:**
```bash
curl -v http://localhost:8080{{arg}}
```

**For POST requests (examples):**

Create Account:
```bash
curl -v -X POST http://localhost:8080/api/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Checking",
    "type": "checking",
    "balance": 100000
  }'
```

Create Category:
```bash
curl -v -X POST http://localhost:8080/api/categories \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Category",
    "color": "#22c55e",
    "description": "Test description"
  }'
```

Create Transaction:
```bash
curl -v -X POST http://localhost:8080/api/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "account_id": "ACCOUNT_UUID",
    "category_id": "CATEGORY_UUID",
    "amount": -5000,
    "description": "Test transaction",
    "date": "2024-01-15T10:30:00Z"
  }'
```

Create Allocation:
```bash
curl -v -X POST http://localhost:8080/api/allocations \
  -H "Content-Type: application/json" \
  -d '{
    "category_id": "CATEGORY_UUID",
    "amount": 50000,
    "period": "2024-01"
  }'
```

### 4. Verify Response

Check:
- [ ] **Status Code**: Is it correct?
  - 200 OK for successful GET/PUT
  - 201 Created for successful POST
  - 400 Bad Request for invalid input
  - 404 Not Found for missing resource
  - 500 Internal Server Error

- [ ] **Response Body**: Is it valid JSON?
- [ ] **Response Content**: Does it match expectations?
- [ ] **Error Handling**: Try invalid input

### 5. Test Error Cases

**Invalid Input:**
```bash
# Missing required field
curl -v -X POST http://localhost:8080/api/accounts \
  -H "Content-Type: application/json" \
  -d '{"name": "Test"}'
```

**Non-existent Resource:**
```bash
# Invalid ID
curl -v http://localhost:8080/api/accounts/invalid-uuid-12345
```

**Invalid Data Type:**
```bash
# String instead of integer for balance
curl -v -X POST http://localhost:8080/api/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test",
    "type": "checking",
    "balance": "not a number"
  }'
```

### 6. Pretty Print JSON Response

For easier reading:
```bash
curl -s http://localhost:8080{{arg}} | python3 -m json.tool
```

Or use jq if available:
```bash
curl -s http://localhost:8080{{arg}} | jq .
```

## Test Checklist

- [ ] Endpoint responds
- [ ] Status code is correct
- [ ] Response format is valid JSON
- [ ] Response contains expected fields
- [ ] Error responses are handled correctly
- [ ] Invalid input returns 400
- [ ] Non-existent resource returns 404

## Integration Tests

After testing one endpoint, test the workflow:

**Example: Account → Transaction → Balance Update**
```bash
# 1. Create account
ACCOUNT_ID=$(curl -s -X POST http://localhost:8080/api/accounts \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","type":"checking","balance":100000}' \
  | jq -r '.id')

# 2. Get account and verify balance
curl -s http://localhost:8080/api/accounts/$ACCOUNT_ID | jq '.balance'

# 3. Create category
CATEGORY_ID=$(curl -s -X POST http://localhost:8080/api/categories \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Category","color":"#22c55e"}' \
  | jq -r '.id')

# 4. Create transaction (spend $50)
curl -s -X POST http://localhost:8080/api/transactions \
  -H "Content-Type: application/json" \
  -d "{
    \"account_id\":\"$ACCOUNT_ID\",
    \"category_id\":\"$CATEGORY_ID\",
    \"amount\":-5000,
    \"description\":\"Test\",
    \"date\":\"2024-01-15T10:30:00Z\"
  }"

# 5. Verify balance updated
curl -s http://localhost:8080/api/accounts/$ACCOUNT_ID | jq '.balance'
# Should be 95000 (100000 - 5000)
```

## Report Results

After testing, report:
- Endpoint tested
- Status code received
- Response sample
- Any errors or issues
- Whether it behaves as expected
