# Mini Banking Platform

A full-stack banking application with double-entry ledger system for financial integrity.

## User Management Approach

**Chosen: Both Option A and Option B**

Each user receives on creation:
- 1 USD account with initial balance `100000` cents ($1000.00).
- 1 EUR account with initial balance `50000` cents (€500.00).

## Tech Stack

**Backend:** Go 1.24, PostgreSQL, JWT, Goose migrations, Docker

**Frontend:** React 19, TypeScript, Tailwind CSS, React Router

## Quick Start

### Run with Docker

```bash
docker-compose up --build
```

- Frontend: http://localhost:3000
- Backend: http://localhost:8080

### Test Users

```
alice@example.com / password123
bob@example.com / password123
charlie@example.com / password123
```

## Architecture

### Double-Entry Ledger Design

**Database Schema:**

```sql
users (
  id          uuid primary key,
  email       text unique not null,
  password    text not null,
  first_name  text not null,
  last_name   text not null,
  created_at  timestamp default now(),
  updated_at  timestamp default now()
);

accounts (
  id            uuid primary key,
  user_id       uuid not null references users(id) on delete cascade,
  currency      varchar(3) not null check (currency in ('USD','EUR')),
  balance_cents bigint not null default 0,
  created_at    timestamp default now(),
  updated_at    timestamp default now(),
  unique (user_id, currency)
);

transactions (
  id           uuid primary key,
  type         text not null check (type in ('initial_deposit','transfer','exchange')),
  from_user_id uuid not null,
  to_user_id   uuid,
  currency     varchar(3) not null,
  amount_cents bigint not null,
  description  text,
  created_at   timestamp default now()
);

ledger_entries (
  id             uuid primary key,
  transaction_id uuid not null references transactions(id) on delete cascade,
  account_id     uuid not null references accounts(id) on delete cascade,
  amount_cents   bigint not null,
  created_at     timestamp default now()
);
```

**Transaction Flow (in cents):**

Initial Deposit (new user):
- Transaction: `initial_deposit`, `amount_cents = 100000` (USD) and `50000` (EUR).
- Ledger Entries: `+100000` to USD account, `+50000` to EUR account.

Transfer (User A → User B, $100 USD):
- Transaction: `transfer`, `amount_cents = 10000`.
- Ledger Entry 1: debit A `-10000`.
- Ledger Entry 2: credit B `+10000`.
- Sum over all USD accounts for this transaction: `0` (double-entry).

Exchange (User A, $100 USD → EUR):
- Fixed rate `USD→EUR = 23/25` (0.92) modeled as integers.
- Transaction: `exchange`, `amount_cents = 10000` (USD cents).
- Ledger Entry 1: debit USD `-10000`.
- Ledger Entry 2: credit EUR `+9200` (computed as `(10000*23)/25`).

The ledger (`ledger_entries`) is the authoritative audit trail. `accounts.balance_cents` are maintained for performance and can be reconciled back to the ledger at any time.

### Balance Consistency Approach

**Problem:** Keep accounts.balance synchronized with ledger during concurrent operations.

**Solution:**

1. Atomic Transactions:
```go
tx := db.BeginTx()
defer tx.Rollback()

tx.Commit()
```

2. Row-Level Locking:
```go
SELECT balance FROM accounts WHERE id = $1 FOR UPDATE
```
Prevents race conditions. Other transactions wait.

3. Reconciliation Endpoint:
```text
GET /api/v1/accounts/reconcile
```
For each of the authenticated user's accounts, the system returns:
- `balance_cents` from `accounts`.
- `ledger_sum_cents` as `SUM(ledger_entries.amount_cents)` for that account.
- `difference_cents = balance_cents - ledger_sum_cents`.
- `is_balanced` (true if `difference_cents == 0`).

## Design Decisions and Trade-offs

### 1. Atomic Transactions

**Decision:** All financial operations wrapped in database transactions.

**Why:** Ensures data consistency. Either all changes succeed or all are rolled back.

**Trade-off:** Slightly more complex error handling. Cannot partially complete operations.

### 2. Row-Level Locking

**Decision:** Use SELECT FOR UPDATE for account balance queries.

**Why:** Prevents race conditions in concurrent operations. Multiple users can transfer simultaneously without corrupting balances.

**Trade-off:** Reduced concurrency. Transactions wait for locks to be released.

### 3. Monorepo Structure

**Decision:** Single repository for frontend and backend.

**Why:** Easier development, shared configuration, faster iteration.

**Trade-off:** Larger repository, difficult to scale separate teams.

### 4. Fixed Exchange Rate

**Decision:** Use hardcoded rational rates in integer math: `USD→EUR = 23/25`, `EUR→USD = 25/23`.

**Why:** Avoids floating-point rounding issues, keeps all amounts in integer cents, and makes residuals (spread) explicit and auditable.

**Trade-off:** Rate changes require a code/config change and redeploy. No dynamic FX feeds.

### 5. Pre-seeded Users and Registration

**Decision:** Support both pre-seeded demo users and full registration flow.

**Why:** Pre-seeded users make evaluation and demos instant; registration is required for a realistic banking flow.

**Trade-off:** There is extra code to maintain for seeding plus regular auth logic.

## Known Limitations

1. **Fixed Exchange Rate** - Hardcoded, requires code changes to update
2. **No Email Verification** - Registration accepts any email
3. **Simple Authorization** - No roles or permissions system
4. **JWT Not Revocable** - Tokens valid until expiry
5. **No Transaction Reversal** - Completed transactions cannot be reversed
6. **Limited Currencies** - Only USD and EUR supported
7. **No Rate Limiting** - No protection against API abuse
8. **Basic Error Messages** - Generic errors for some cases


## Testing

### Unit Tests

```bash
cd backend
go test -v ./internal/service/
```

Service-level tests cover, among other things:
- **Transfer operations**: success, insufficient funds, ledger balance verification.
- **Exchange operations**: success, no money minting on round-trip, minimum amount enforcement, integer overflow protection.
- **Concurrent operations**: race-condition prevention with row-level locking and correct total balances.
- **Authentication**: atomic registration with initial deposits, duplicate email handling, login validation.
- **Account operations**: listing user accounts, per-account balances, and reconciliation of `balance_cents` vs ledger entries.

Test database required: `banking_platform_test`

To create test database:
```bash
psql -U postgres -c "CREATE DATABASE banking_platform_test;"
```

## Development

### Backend
```bash
cd backend
make install
make run
```

### Frontend
```bash
cd frontend
npm install
npm start
```

### Database Migrations

In Docker, database migrations are executed automatically on backend startup via `goose` in the application bootstrap.
For manual control during local development:

```bash
cd backend
make migrate-up   # apply all pending migrations using current DB_* env vars
make migrate-down # roll back the last migration
```

## API Documentation

OpenAPI specification: `docs/openapi.yaml`

View interactive docs:
```bash
docker run -p 8081:8080 -e SWAGGER_JSON=/api/openapi.yaml \
  -v $(pwd)/docs/openapi.yaml:/api/openapi.yaml \
  swaggerapi/swagger-ui
```

Open http://localhost:8081


