# OFX/QFX Transaction Import Feature

## Context

This is a personal budget application built with Go + SQLite backend and vanilla JavaScript frontend, following **Clean Architecture principles** (Domain → Application → Infrastructure layers).

**Current State**: Users manually create transactions.

**Goal**: Enable users to import transactions from bank OFX/QFX files exported from:
- OnPoint Community Credit Union (checking, savings, premium savings, credit card)
- JP Morgan Chase (checking, savings, 2 credit cards)
- Wells Fargo (checking)

All these institutions support OFX/QFX export, which is a standardized XML-based financial data format.

## Problem to Solve

1. **Current model requires category_id**: The Transaction model currently requires every transaction to have a category (`category_id TEXT NOT NULL`). Imported transactions won't have categories initially.

2. **Need duplicate detection**: Re-importing the same file shouldn't create duplicate transactions.

3. **Need user workflow**: Import → Review uncategorized → Assign categories → Include in budget.

## Requirements

### Core Functionality

**Import OFX/QFX Files**
- User selects an account and uploads an OFX/QFX file
- Parse the file and extract transactions (date, amount, description, merchant)
- Convert amounts from dollars to cents (stored as int64)
- Check for duplicates (same account + date + amount + description)
- Import new transactions as uncategorized (category_id = null)
- Update account balance to reflect imported transactions
- Return import summary (total, imported, skipped, errors)

**Manage Uncategorized Transactions**
- List all transactions where category_id is null
- Allow user to assign categories to individual transactions
- Support bulk categorization (assign same category to multiple transactions)
- Once categorized, transactions participate normally in budget allocations

**Data Integrity**
- Schema migration to make category_id nullable
- Maintain foreign key constraints (when category_id is set)
- Preserve all existing data during migration
- Handle null values throughout the codebase

### Technical Constraints

**Library**: Use `github.com/aclindsa/ofxgo` for parsing OFX/QFX files

**Clean Architecture**: Maintain clear separation of concerns
- **Domain layer**: Core business logic, no external dependencies
- **Application layer**: Use cases and service orchestration
- **Infrastructure layer**: Database, HTTP, file parsing, external libraries

**Database**: SQLite with proper indexing and foreign keys

**API Design**: RESTful endpoints following existing patterns

**File Upload**: Limit to 10MB, validate file extensions (.ofx, .qfx)

## Key Design Decisions

**Nullable Category ID**
- Change `category_id TEXT NOT NULL` → `category_id TEXT` (nullable)
- Update domain model: `CategoryID *string` (pointer for nullable)
- This is cleaner than creating a fake "Uncategorized" category

**Duplicate Detection Strategy**
- Match on: account_id + date + amount + description
- Skip duplicates silently (count in results)
- No need for FitID tracking at this stage

**Transaction Description**
- Combine OFX Name and Memo fields intelligently
- Format: "Name - Memo" (if both exist and differ)
- Or just: "Name" (if Memo is empty or same as Name)

## Acceptance Criteria

- [ ] Existing transactions continue to work after migration
- [ ] Can import OFX/QFX files from all three financial institutions
- [ ] Duplicate transactions are detected and skipped
- [ ] Account balances update correctly after import
- [ ] Can list all uncategorized transactions
- [ ] Can assign category to individual transaction
- [ ] Can bulk assign category to multiple transactions
- [ ] UI shows import results (imported count, skipped count, any errors)
- [ ] Uncategorized transactions display in UI with ability to categorize
- [ ] Clear error messages for invalid files or failed imports

## Implementation Guidelines

**Follow Clean Architecture**:
1. Start with domain changes (make Transaction.CategoryID nullable)
2. Update repository interface and implementation
3. Create application service for import logic (ImportService)
4. Add infrastructure for OFX parsing (separate from business logic)
5. Build HTTP handlers and routes
6. Update frontend to support new features

**Database Migration**:
- Create a migration system (not just initSchema)
- SQLite doesn't support ALTER COLUMN, use table recreation pattern
- Track applied migrations in schema_migrations table
- Run migrations automatically on startup

**Error Handling**:
- Validate file format before parsing
- Handle missing/corrupted OFX data gracefully
- Return helpful error messages to user
- Don't fail entire import if one transaction has issues

**Testing**:
- Test with real OFX files from each bank
- Verify duplicate detection (import same file twice)
- Test with and without categories
- Verify account balance calculations
- Test edge cases (empty files, invalid formats, missing fields)

## API Endpoints to Add

```
POST   /api/transactions/import           # Upload OFX file, import transactions
GET    /api/transactions?uncategorized=true  # List uncategorized transactions
PUT    /api/transactions/bulk-categorize  # Assign category to multiple transactions
```

## Resources

- OFX Specification: https://www.ofx.net/
- ofxgo Library: https://github.com/aclindsa/ofxgo
- Existing codebase structure: Follow patterns in `internal/`

## Remember

- **Clean Architecture**: Domain logic should be pure Go, no external dependencies
- **Nullable Handling**: Use `*string` for CategoryID, check for nil throughout
- **User Experience**: Clear feedback on import success/failure
- **Data Safety**: Never lose existing transactions, preserve balances
- **Simplicity**: Start with core functionality, iterate based on real usage

Good luck! Reach out if you need clarification on existing architecture patterns or business logic.
