# Budget App

A budgeting application built with Go using clean architecture principles. Track your income, expenses, and budgets with ease.

## Architecture

This project follows clean architecture with clear separation of concerns:

```
budget/
├── cmd/server/           # Application entry point
├── config/              # Configuration management
├── internal/
│   ├── domain/          # Domain layer (entities, interfaces)
│   ├── application/     # Application layer (business logic, use cases)
│   └── infrastructure/  # Infrastructure layer (database, HTTP, repositories)
```

### Layers

- **Domain Layer**: Core business entities (User, Category, Transaction, Budget) and repository interfaces
- **Application Layer**: Business logic and use cases (services)
- **Infrastructure Layer**: Implementation details (database, HTTP handlers, repository implementations)

## Features

- User management
- Income and expense categories
- Transaction tracking
- Budget planning and monitoring
- Budget vs actual spending summaries

## Prerequisites

- Go 1.22 or higher
- SQLite3

## Getting Started

### Installation

1. Clone the repository:
```bash
git clone https://github.com/billybbuffum/budget.git
cd budget
```

2. Install dependencies:
```bash
go mod download
```

### Running the Application

Start the server:
```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080` by default.

### Configuration

Configure the application using environment variables:

- `PORT`: Server port (default: 8080)
- `DB_PATH`: SQLite database file path (default: budget.db)

Example:
```bash
PORT=3000 DB_PATH=/path/to/database.db go run cmd/server/main.go
```

## API Documentation

### Health Check

```
GET /health
```

Returns: `200 OK`

### Users

#### Create User
```
POST /api/users
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com"
}
```

#### Get User
```
GET /api/users/{id}
```

#### List Users
```
GET /api/users
```

#### Update User
```
PUT /api/users/{id}
Content-Type: application/json

{
  "name": "Jane Doe",
  "email": "jane@example.com"
}
```

#### Delete User
```
DELETE /api/users/{id}
```

### Categories

#### Create Category
```
POST /api/categories
Content-Type: application/json

{
  "name": "Groceries",
  "type": "expense",
  "description": "Food and household items",
  "color": "#FF5733"
}
```

Types: `income` or `expense`

#### Get Category
```
GET /api/categories/{id}
```

#### List Categories
```
GET /api/categories
GET /api/categories?type=expense
```

#### Update Category
```
PUT /api/categories/{id}
Content-Type: application/json

{
  "name": "Food & Groceries",
  "color": "#FF6644"
}
```

#### Delete Category
```
DELETE /api/categories/{id}
```

### Transactions

#### Create Transaction
```
POST /api/transactions
Content-Type: application/json

{
  "user_id": "user-uuid",
  "category_id": "category-uuid",
  "amount": 52.99,
  "description": "Weekly groceries",
  "date": "2024-10-29T10:00:00Z"
}
```

#### Get Transaction
```
GET /api/transactions/{id}
```

#### List Transactions
```
GET /api/transactions
GET /api/transactions?user_id=user-uuid
GET /api/transactions?category_id=category-uuid
GET /api/transactions?start_date=2024-10-01T00:00:00Z&end_date=2024-10-31T23:59:59Z
```

#### Update Transaction
```
PUT /api/transactions/{id}
Content-Type: application/json

{
  "amount": 55.00,
  "description": "Weekly groceries (updated)"
}
```

#### Delete Transaction
```
DELETE /api/transactions/{id}
```

### Budgets

#### Create Budget
```
POST /api/budgets
Content-Type: application/json

{
  "category_id": "category-uuid",
  "amount": 500.00,
  "period": "2024-10",
  "notes": "Monthly grocery budget"
}
```

Period format: `YYYY-MM`

#### Get Budget
```
GET /api/budgets/{id}
```

#### List Budgets
```
GET /api/budgets
GET /api/budgets?period=2024-10
```

#### Get Budget Summary
```
GET /api/budgets/summary?period=2024-10
```

Returns budget vs actual spending with percentages.

#### Update Budget
```
PUT /api/budgets/{id}
Content-Type: application/json

{
  "amount": 600.00,
  "notes": "Increased budget for holidays"
}
```

#### Delete Budget
```
DELETE /api/budgets/{id}
```

## Example Workflow

1. **Create users** for you and your wife
2. **Create categories** for your expenses (groceries, rent, utilities) and income (salary)
3. **Set budgets** for each expense category for the current month
4. **Track transactions** as they occur
5. **Check budget summary** to see how you're doing against your budget

## Building for Production

```bash
go build -o budget-server cmd/server/main.go
./budget-server
```

## Development

Run tests:
```bash
go test ./...
```

## License

MIT
