---
description: Add a new API endpoint to an existing feature
argument-hint: <endpoint-description>
---

# Create New API Endpoint

Add a new API endpoint to the Budget application.

## Endpoint Description
{{arg}}

## Requirements

Please create the endpoint with the following steps:

### 1. Identify the Component
Determine which handler this endpoint belongs to:
- AccountHandler: `/api/accounts/*`
- CategoryHandler: `/api/categories/*`
- TransactionHandler: `/api/transactions/*`
- AllocationHandler: `/api/allocations/*`

### 2. Add Service Method (if needed)
If new business logic is required:
- Add method to appropriate service in `internal/application/`
- Implement business logic
- Use repository interfaces
- Return appropriate errors

### 3. Add Handler Method
In `internal/infrastructure/http/handlers/`:
- Parse request (body, query params, path params)
- Validate input
- Call service method
- Return JSON response with appropriate status code

**Status Codes:**
- `200 OK`: Successful GET/PUT
- `201 Created`: Successful POST
- `204 No Content`: Successful DELETE
- `400 Bad Request`: Invalid input
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

### 4. Register Route
Update `internal/infrastructure/http/router.go`:
- Add route with appropriate HTTP method
- Follow RESTful conventions
- Use path parameters where appropriate (e.g., `{id}`)

### 5. Error Handling
- Validate all inputs
- Return descriptive error messages
- Wrap errors with context
- Handle edge cases

## API Design Guidelines

**RESTful Conventions:**
- `GET` for retrieval
- `POST` for creation
- `PUT` for full updates
- `PATCH` for partial updates (if needed)
- `DELETE` for deletion

**Path Structure:**
```
GET    /api/resources         - List all
POST   /api/resources         - Create new
GET    /api/resources/{id}    - Get by ID
PUT    /api/resources/{id}    - Update by ID
DELETE /api/resources/{id}    - Delete by ID
```

**Query Parameters:**
Use for filtering, pagination, sorting:
```
GET /api/transactions?account_id=xxx&start_date=2024-01-01
```

**Request/Response Format:**
- Always JSON
- Use snake_case for JSON fields (Go struct tags)
- Return complete entity on create/update
- Return array for list endpoints

## Example Implementation

```go
// Handler method
func (h *Handler) GetResourceByID(w http.ResponseWriter, r *http.Request) {
    // 1. Parse path parameter
    id := r.PathValue("id")

    // 2. Call service
    resource, err := h.service.GetByID(id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            http.Error(w, "Resource not found", http.StatusNotFound)
            return
        }
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // 3. Return JSON response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resource)
}
```

## Testing

After creating the endpoint:
- Test with curl or Postman
- Verify status codes
- Test error cases
- Check response format

## Documentation

Update documentation:
- Add endpoint to Claude.md
- Document request/response format
- Provide curl example
- Document error responses

## Next Steps

1. Show me the implementation
2. Test the endpoint
3. Invoke api-documenter agent to document it
