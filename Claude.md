# Claude.md - Project Technical Documentation

This document provides detailed technical information about the Budget App for Claude and other developers working on the project.

## Project Overview

**Type**: Zero-Based Budgeting Application
**Language**: Go 1.23
**Database**: SQLite3
**Architecture**: Clean Architecture with clear layer separation
**API Style**: RESTful HTTP API

## Architecture

This project follows **Clean Architecture** principles with three distinct layers:

### 1. Domain Layer (`internal/domain/`)
- Contains core business entities and repository interfaces
- No external dependencies
- Defines the contract that other layers must follow

**Entities:**
- `Account`: Financial accounts (checking, savings, cash)
- `Category`: Income and expense categories
- `Transaction`: Money movements between accounts and categories
- `Allocation`: Zero-based budget allocations per category per period

**Interfaces:**
- `AccountRepository`, `CategoryRepository`, `TransactionRepository`, `AllocationRepository`

### 2. Application Layer (`internal/application/`)
- Business logic and use cases (services)
- Orchestrates domain entities
- Depends only on domain layer interfaces

**Services:**
- `AccountService`: Account management and balance calculations
- `CategoryService`: Category CRUD operations
- `TransactionService`: Transaction management with account balance updates
- `AllocationService`: Zero-based budgeting logic with rollover support

### 3. Infrastructure Layer (`internal/infrastructure/`)
- Implementation details for data persistence and HTTP
- Implements repository interfaces from domain layer
- HTTP handlers, routing, and server setup

**Components:**
- `database/`: SQLite setup and schema
- `http/`: Server, router, and handlers
- `repository/`: Concrete implementations of repository interfaces

## Core Entities Explained

### Account
Financial accounts that hold money.

**Fields:**
- `ID`: UUID
- `Name`: Account name (e.g., "Chase Checking")
- `Type`: checking, savings, or cash
- `Balance`: Current balance in cents
- `CreatedAt`, `UpdatedAt`: Timestamps

**Key Logic:**
- Balance stored in cents to avoid floating-point precision issues
- Summary endpoint returns total balance across all accounts

### Category
Budget categories for organizing transactions.

**Fields:**
- `ID`: UUID
- `Name`: Category name (e.g., "Groceries")
- `Type`: income or expense
- `Description`: Optional description
- `Color`: Hex color for UI (e.g., "#FF5733")
- `CreatedAt`, `UpdatedAt`: Timestamps

**Key Logic:**
- Can only allocate money to expense categories
- Categories can have multiple transactions
- Filterable by type (income/expense)

### Transaction
Individual money movements.

**Fields:**
- `ID`: UUID
- `AccountID`: Which account the transaction belongs to
- `CategoryID`: Which category the transaction belongs to
- `Amount`: Amount in cents (positive = inflow, negative = outflow)
- `Description`: Transaction description
- `Date`: Transaction date
- `CreatedAt`, `UpdatedAt`: Timestamps

**Key Logic:**
- Creating/updating/deleting a transaction automatically updates the account balance
- Transactions can be filtered by account, category, and date range
- Used to calculate actual spending vs allocated budget

### Allocation
Zero-based budget allocations.

**Fields:**
- `ID`: UUID
- `CategoryID`: Which expense category this allocation is for
- `Amount`: Allocated amount in cents
- `Period`: Budget period in YYYY-MM format
- `CreatedAt`, `UpdatedAt`: Timestamps

**Key Logic:**
- One allocation per category per period (unique constraint)
- POST/PUT uses upsert logic to update existing allocations
- Rollover support: Unspent money from previous periods carries forward
- "Available" calculation accounts for all historical allocations and spending

## Zero-Based Budgeting Implementation

The core concept: Every dollar should be allocated to a category, leaving zero unassigned.

### Key Calculations

1. **Ready to Assign**:
   ```
   Total Account Balance - Total Allocated Amount (across all time)
   ```
   Shows how much money is available to allocate to categories.

2. **Available for Category** (with rollover):
   ```
   Sum of all allocations for category - Sum of all spending for category
   ```
   This accounts for all history, so unspent money automatically rolls over.

3. **Allocation Summary** (per period):
   - **Allocated**: Amount budgeted for the category in the period
   - **Spent**: Actual spending (sum of negative transactions) in the period
   - **Available**: Available amount (see #2 above)

### Allocation Workflow

1. User adds money to an account (transaction with income category)
2. Account balance increases, "Ready to Assign" increases
3. User creates allocations for expense categories for a specific period
4. "Ready to Assign" decreases by allocated amounts
5. User records expenses (transactions with expense categories)
6. Account balance decreases, category spending increases
7. "Available" for each category reflects allocated minus spent
8. Unspent money automatically carries to future periods

## API Endpoints

### Health Check
- `GET /health` - Server health check

### Accounts
- `POST /api/accounts` - Create account
- `GET /api/accounts` - List all accounts
- `GET /api/accounts/summary` - Get total balance across all accounts
- `GET /api/accounts/{id}` - Get account by ID
- `PUT /api/accounts/{id}` - Update account
- `DELETE /api/accounts/{id}` - Delete account

### Categories
- `POST /api/categories` - Create category
- `GET /api/categories` - List all categories (filterable by type)
- `GET /api/categories/{id}` - Get category by ID
- `PUT /api/categories/{id}` - Update category
- `DELETE /api/categories/{id}` - Delete category

### Transactions
- `POST /api/transactions` - Create transaction
- `GET /api/transactions` - List transactions (filterable by account, category, date range)
- `GET /api/transactions/{id}` - Get transaction by ID
- `PUT /api/transactions/{id}` - Update transaction
- `DELETE /api/transactions/{id}` - Delete transaction

**Query Parameters:**
- `account_id`: Filter by account
- `category_id`: Filter by category
- `start_date`: Filter by start date (RFC3339 format)
- `end_date`: Filter by end date (RFC3339 format)

### Allocations
- `POST /api/allocations` - Create/update allocation (upsert by category+period)
- `GET /api/allocations` - List all allocations
- `GET /api/allocations/summary?period=YYYY-MM` - Get allocation summary for period
- `GET /api/allocations/ready-to-assign` - Get amount available to allocate
- `POST /api/allocations/cover-underfunded` - Manually allocate to cover underfunded credit card
- `GET /api/allocations/{id}` - Get allocation by ID
- `DELETE /api/allocations/{id}` - Delete allocation

## Database Schema

SQLite database with 4 tables:

```sql
CREATE TABLE accounts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('checking', 'savings', 'cash')),
    balance INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE categories (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('income', 'expense')),
    description TEXT,
    color TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE TABLE transactions (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    category_id TEXT NOT NULL,
    amount INTEGER NOT NULL,
    description TEXT NOT NULL,
    date DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);

CREATE TABLE allocations (
    id TEXT PRIMARY KEY,
    category_id TEXT NOT NULL,
    amount INTEGER NOT NULL,
    period TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE,
    UNIQUE(category_id, period)
);
```

**Key Design Decisions:**
- All amounts stored as INTEGER (cents) for precision
- Foreign key constraints with CASCADE delete
- Unique constraint on (category_id, period) for allocations
- Timestamps in UTC

## Configuration

**Environment Variables:**
- `PORT` (default: 8080) - Server port
- `DB_PATH` (default: budget.db) - SQLite database file path

**Docker Configuration:**
- Database path in container: `/app/data/budget.db`
- Persisted via Docker volume: `budget-data`
- Port mapping: 8080:8080

## Running the Application

### Docker Compose (Recommended)
```bash
docker-compose up -d
```

### Local Go Development
```bash
# Install npm dependencies and build CSS
npm install
npm run build:css

# Run the Go application
go mod download
go run cmd/server/main.go
```

**Note:** During development, you can run Tailwind in watch mode to automatically rebuild CSS on changes:
```bash
npm run watch:css
```

### Docker Direct
```bash
docker build -t budget-app .
docker run -d -p 8080:8080 -v budget-data:/app/data budget-app
```

## Code Organization

```
/home/user/budget/
├── cmd/server/main.go              # Entry point, dependency injection
├── config/config.go                # Configuration from environment
├── internal/
│   ├── domain/                     # Core entities and interfaces
│   │   ├── account.go
│   │   ├── category.go
│   │   ├── transaction.go
│   │   ├── allocation.go
│   │   └── repository.go           # Repository interface definitions
│   ├── application/                # Business logic
│   │   ├── account_service.go
│   │   ├── category_service.go
│   │   ├── transaction_service.go
│   │   └── allocation_service.go
│   └── infrastructure/             # Implementation details
│       ├── database/sqlite.go      # Database setup and schema
│       ├── http/
│       │   ├── server.go           # HTTP server wrapper
│       │   ├── router.go           # Route definitions
│       │   └── handlers/           # HTTP request handlers
│       └── repository/             # Repository implementations
├── Dockerfile                      # Multi-stage build
├── docker-compose.yml              # Docker Compose config
├── go.mod                          # Go dependencies
└── README.md                       # High-level documentation
```

## Development Guidelines

### Adding a New Feature

1. **Define entity in domain layer** (if needed)
   - Add struct in `internal/domain/`
   - Add repository interface in `internal/domain/repository.go`

2. **Create service in application layer**
   - Add service in `internal/application/`
   - Implement business logic using repository interfaces

3. **Implement repository in infrastructure layer**
   - Add repository implementation in `internal/infrastructure/repository/`
   - Implement SQL queries

4. **Create HTTP handler**
   - Add handler in `internal/infrastructure/http/handlers/`
   - Parse request, call service, return response

5. **Register routes**
   - Update `internal/infrastructure/http/router.go`

6. **Update database schema**
   - Modify `internal/infrastructure/database/sqlite.go`

### Error Handling

- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- Return appropriate HTTP status codes:
  - 200: Success
  - 201: Created
  - 204: No Content (delete)
  - 400: Bad Request (validation errors)
  - 404: Not Found
  - 500: Internal Server Error

### Testing Approach

- Unit test services with mock repositories
- Integration test repositories with test database
- End-to-end test HTTP handlers

## Common Tasks

### Adding a New Endpoint

1. Add handler method in appropriate handler file
2. Register route in `router.go`
3. Add service method if needed
4. Add repository method if needed

### Modifying Database Schema

1. Update schema in `sqlite.go`
2. Update domain entity struct
3. Update repository SQL queries
4. Consider migration strategy for existing data

### Adding Query Filters

1. Add query parameters to handler
2. Pass filters to service method
3. Update repository to support new filters
4. Modify SQL WHERE clause

## Key Dependencies

- `github.com/google/uuid` - UUID generation
- `github.com/mattn/go-sqlite3` - SQLite driver
- Standard library only for HTTP server

## Future Enhancements

Potential features to consider:
- User authentication and multi-user support
- Recurring transactions/allocations
- Budget templates
- Reports and analytics
- Export to CSV/Excel
- Web frontend
- Mobile apps

## Notes for Claude

### Git Configuration
**IMPORTANT:** All commits in this repository must be attributed to the repository owner, not Claude.

Before making any commits, verify and set the git configuration:

```bash
# Check current configuration
git config user.name
git config user.email

# If not set correctly, configure it:
git config user.name "billybbuffum"
git config user.email "billybbuffum@users.noreply.github.com"
```

**Required git author for all commits:**
- Name: `billybbuffum`
- Email: `billybbuffum@users.noreply.github.com`

This configuration should be automatically read from `.git/config`, but if it's not working in your session, manually set it using the commands above before making any commits.

### Budget Application Guidelines

- Always use cents for money amounts (INTEGER in database)
- Allocations only work with expense categories
- One allocation per category per period (upsert behavior)
- Transaction operations must update account balances atomically
- Ready to Assign = Total Balance - Total Allocated
- Available per category includes all history (rollover support)
