---
description: Scaffold a new feature following clean architecture patterns
argument-hint: <feature-name>
---

# Create New Feature

Create a new feature for the Budget application following clean architecture principles.

## Feature Name
{{arg}}

## Requirements

Please create a new feature with the following structure:

### 1. Domain Layer (if new entity needed)
Create entity in `internal/domain/`:
- Define the entity struct
- Add repository interface in `internal/domain/repository.go`
- Include proper validation methods on the entity

### 2. Application Layer
Create service in `internal/application/`:
- Implement business logic
- Use repository interfaces (not concrete implementations)
- Add proper error handling with context
- Follow existing service patterns

### 3. Infrastructure Layer

**Repository Implementation** (`internal/infrastructure/repository/`):
- Implement the repository interface
- Use parameterized SQL queries (prevent SQL injection)
- Handle errors appropriately
- Follow existing repository patterns

**HTTP Handler** (`internal/infrastructure/http/handlers/`):
- Parse request body/parameters
- Call service methods
- Return appropriate HTTP status codes:
  - 200: Success
  - 201: Created
  - 400: Bad Request (validation errors)
  - 404: Not Found
  - 500: Internal Server Error
- Use consistent JSON response format

### 4. Router Configuration
Update `internal/infrastructure/http/router.go`:
- Register all endpoints for the new feature
- Follow RESTful conventions
- Group related routes

### 5. Database Schema (if needed)
Update `internal/infrastructure/database/sqlite.go`:
- Add CREATE TABLE statement
- Include appropriate constraints
- Add indexes if needed
- Store amounts as INTEGER (cents)
- Use TEXT for UUIDs

### 6. Main.go (if needed)
Update `cmd/server/main.go`:
- Initialize new repository
- Initialize new service
- Initialize new handler
- Wire dependencies

## Clean Architecture Checklist

- [ ] Domain layer has no external dependencies
- [ ] Application layer only uses domain interfaces
- [ ] Infrastructure layer implements interfaces
- [ ] Dependencies point inward
- [ ] Business logic is in services, not handlers
- [ ] Database logic is in repositories, not services

## Code Quality

- [ ] All functions have clear, descriptive names
- [ ] Error messages include context
- [ ] Amounts are stored as integers (cents)
- [ ] SQL uses parameterized queries
- [ ] Proper HTTP status codes
- [ ] Consistent with existing code patterns

## Documentation

After creating the feature:
- Update Claude.md with entity and endpoint documentation
- Add comments to exported functions
- Document any complex business logic

## Next Steps

After you create the feature:
1. Show me the files created
2. Invoke the test-generator agent to create tests
3. Invoke the code-reviewer agent to review the implementation
