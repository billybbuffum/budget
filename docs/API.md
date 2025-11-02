# Budget API Documentation

## Allocation Endpoints

### POST /api/allocations/cover-underfunded

Manually allocate funds from Ready to Assign to cover an underfunded payment category.

**Purpose:** Creates an allocation to cover a payment category that doesn't have enough money to cover its associated credit card balance.

**Request:**
```http
POST /api/allocations/cover-underfunded
Content-Type: application/json

{
  "payment_category_id": "550e8400-e29b-41d4-a716-446655440000",
  "period": "2025-10"
}
```

**Request Parameters:**
- `payment_category_id` (string, required): UUID of the payment category to cover
- `period` (string, required): Budget period in YYYY-MM format

**Validation:**
- `payment_category_id` must be a valid UUID
- Category must exist and be a payment category
- `period` must be in YYYY-MM format (e.g., "2025-10")
- Period must be within 2 years past to 5 years future
- Payment category must be underfunded (available < credit card balance)
- Ready to Assign must have sufficient funds to cover the underfunded amount

**Success Response (201 Created):**
```json
{
  "allocation": {
    "id": "660e8400-e29b-41d4-a716-446655440000",
    "category_id": "550e8400-e29b-41d4-a716-446655440000",
    "amount": 20000,
    "period": "2025-10",
    "notes": "Cover underfunded credit card spending",
    "created_at": "2025-10-31T10:30:00Z",
    "updated_at": "2025-10-31T10:30:00Z"
  },
  "underfunded_amount": 20000,
  "ready_to_assign_after": 330000
}
```

**Response Fields:**
- `allocation`: The created or updated allocation record
- `underfunded_amount`: The amount that was underfunded (in cents)
- `ready_to_assign_after`: Ready to Assign amount after this allocation (in cents)

**Error Responses:**

**400 Bad Request - Invalid UUID:**
```json
{
  "error": "invalid UUID format"
}
```

**400 Bad Request - Invalid Period Format:**
```json
{
  "error": "invalid period format, expected YYYY-MM"
}
```

**400 Bad Request - Period Out of Range:**
```json
{
  "error": "period is too far in the past (more than 2 years)"
}
```

**404 Not Found - Category Not Found:**
```json
{
  "error": "payment category not found"
}
```

**400 Bad Request - Not a Payment Category:**
```json
{
  "error": "category is not a payment category"
}
```

**400 Bad Request - Not Underfunded:**
```json
{
  "error": "payment category is not underfunded"
}
```

**400 Bad Request - Insufficient Funds:**
```json
{
  "error": "insufficient funds: Ready to Assign: $33.00, Underfunded: $200.00"
}
```

**500 Internal Server Error:**
```json
{
  "error": "Failed to process allocation request"
}
```

**Usage Example (curl):**
```bash
curl -X POST http://localhost:8080/api/allocations/cover-underfunded \
  -H "Content-Type: application/json" \
  -d '{
    "payment_category_id": "550e8400-e29b-41d4-a716-446655440000",
    "period": "2025-10"
  }'
```

**Usage Example (JavaScript):**
```javascript
async function coverUnderfundedPaymentCategory(categoryId, period) {
  const response = await fetch('/api/allocations/cover-underfunded', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      payment_category_id: categoryId,
      period: period
    })
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error);
  }

  return await response.json();
}

// Usage
try {
  const result = await coverUnderfundedPaymentCategory(
    '550e8400-e29b-41d4-a716-446655440000',
    '2025-10'
  );
  console.log('Allocation created:', result.allocation);
  console.log('Covered amount:', result.underfunded_amount / 100, 'dollars');
  console.log('Ready to Assign after:', result.ready_to_assign_after / 100, 'dollars');
} catch (error) {
  console.error('Failed to cover underfunded:', error.message);
}
```

**Notes:**
- All amounts are in cents (integer values)
- Uses upsert behavior: if allocation exists for (category_id, period), it updates; otherwise creates new
- Operation is atomic to prevent race conditions
- Real-time allocation logic (during transaction creation) remains unchanged
- This endpoint provides manual control for users to cover underfunded payment categories
