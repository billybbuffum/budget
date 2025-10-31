---
description: Generate comprehensive API documentation
allowed-tools: [Read, Write, Glob, Task]
---

# Generate API Documentation

Create comprehensive API documentation for the Budget application.

## Documentation Generation Steps

### 1. Analyze Codebase

Scan for:
- All HTTP handlers in `internal/infrastructure/http/handlers/`
- All routes in `internal/infrastructure/http/router.go`
- Entity structures in `internal/domain/`

### 2. Invoke API Documenter Agent

Use the api-documenter sub agent to generate documentation:
```
Invoke api-documenter agent to create comprehensive API documentation
```

### 3. Documentation Sections to Generate

**API Overview:**
- Base URL
- Authentication (currently none)
- Content types
- Error format
- Status codes

**Entity Documentation:**
For each entity (Account, Category, Transaction, Allocation):
- Field descriptions
- Data types
- Validation rules
- Example JSON

**Endpoint Documentation:**
For each endpoint:
- HTTP method and path
- Description
- Request parameters/body
- Response format
- Status codes
- curl examples
- Notes and gotchas

### 4. Documentation Outputs

Create/Update these files:
- `API.md` - Complete API reference
- Update `Claude.md` - Add/update endpoint documentation
- `EXAMPLES.md` - Usage examples and workflows

## Documentation Format

### Endpoint Template
```markdown
## [METHOD] /api/path

**Description:** What this endpoint does

**Request:**
- Method: GET/POST/PUT/DELETE
- Path: /api/path
- Headers:
  - Content-Type: application/json
- Query Parameters:
  - param1 (type, required/optional): Description
- Body: (if POST/PUT)

**Response:**
- Success (200/201):
- Error (400/404/500):

**Example:**
```bash
curl -X METHOD http://localhost:8080/api/path
```

**Notes:**
- Important considerations
- Side effects
- Business rules
```

## Key Documentation Areas

### Accounts API
- Account types (checking, savings, credit_card)
- Balance in cents
- Credit card special behavior
- Account summary endpoint

### Categories API
- Category fields
- Color format
- Category groups
- System-managed payment categories

### Transactions API
- Positive vs negative amounts
- Date format (RFC3339)
- Balance update side effects
- Query filters (account, category, date range)

### Allocations API
- Zero-based budgeting logic
- Period format (YYYY-MM)
- Rollover behavior
- Ready to Assign calculation
- Allocation summary

## Special Documentation Topics

### Money Handling
- All amounts in cents (integer)
- No floating point arithmetic
- Conversion: dollars * 100 = cents

### Zero-Based Budgeting
- Ready to Assign formula
- Available calculation
- Rollover mechanics
- One allocation per category per period

### Credit Cards
- Negative balances
- Payment category auto-creation
- Payment category behavior

### Date Handling
- RFC3339 format
- UTC timezone
- Period format for allocations

## Example Workflows to Document

### Getting Started
```markdown
1. Create an account
2. Create categories
3. Add income transaction
4. Allocate money to categories
5. Record expenses
6. Check budget status
```

### Monthly Budgeting
```markdown
1. Check Ready to Assign
2. Create allocations for the month
3. Allocate until Ready to Assign = $0
4. Track spending
5. Adjust allocations as needed
```

### Credit Card Usage
```markdown
1. Create credit card account
2. Payment category auto-created
3. Spend on credit card
4. Check payment category
5. Pay credit card bill
```

## Documentation Quality Checklist

- [ ] All endpoints documented
- [ ] Request/response examples provided
- [ ] curl commands are copy-pasteable
- [ ] Field descriptions are clear
- [ ] Validation rules documented
- [ ] Error cases explained
- [ ] Business rules highlighted
- [ ] Complete workflows shown

## Invoke Agent

After reviewing the code structure:

```
Invoke the api-documenter agent to generate complete API documentation
```

The agent will:
1. Read handler files
2. Read router configuration
3. Read entity definitions
4. Generate comprehensive documentation
5. Include examples and workflows
6. Document error cases
7. Explain business rules

## Output Location

Documentation will be created in:
- `docs/API.md` - Main API reference
- `docs/EXAMPLES.md` - Usage examples
- Updated `Claude.md` - Technical reference
