# Mini Banking Platform

Simplified banking platform with a double-entry ledger. Focus is on data integrity,
transaction atomicity, and clean audit trails rather than UI polish.

## Scope and Requirements

- Two currencies: USD and EUR
- Double-entry ledger for every transaction
- Account balances stored for performance and reconciled to ledger
- JWT authentication
- REST API with transfer, exchange, and transaction history

## Tech Stack

- Backend: Go 1.24, PostgreSQL, JWT, Goose, Docker
- Frontend: React 19, TypeScript, Tailwind CSS, React Router

## Quick Start (Docker)

```bash
docker-compose up --build
```

- Frontend: http://localhost:3000
- Backend: http://localhost:8080

Test users:
```
alice@example.com / password123
bob@example.com / password123
charlie@example.com / password123
```

## System Design

### Data Model (high level)

- `users`: user identities
- `accounts`: per-user currency wallets (USD, EUR)
- `transactions`: user-facing history
- `ledger_entries`: authoritative double-entry audit trail

### Ledger and Balances

All amounts are stored in integer cents. Each transaction produces balanced
ledger entries. `accounts.balance_cents` is derived and maintained inside the
same DB transaction as ledger writes.

### Exchange and FX System Accounts

Exchange uses a fixed rate modeled as integer ratios:
- USD -> EUR: 23/25
- EUR -> USD: 25/23

To keep double-entry strict per currency, exchange uses FX system accounts
in both currencies. Example for $100 USD -> EUR:

- User USD: `-10000`, FX USD: `+10000`
- FX EUR: `-9200`, User EUR: `+9200`

Each currency sums to `0` for the transaction.

### Consistency Guarantees

- All financial operations are wrapped in a DB transaction.
- Row-level locks (`SELECT ... FOR UPDATE`) prevent race conditions.
- Reconciliation endpoint verifies `accounts.balance_cents` vs ledger sum.

## API Summary

Authentication:
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /api/v1/auth/me`

Accounts:
- `GET /api/v1/accounts`
- `GET /api/v1/accounts/:id/balance`
- `GET /api/v1/accounts/reconcile`

Transactions:
- `POST /api/v1/transactions/transfer`
- `POST /api/v1/transactions/exchange`
- `GET /api/v1/transactions?type=transfer|exchange|initial_deposit`

`initial_deposit` entries are written on user creation and are included in
`/api/v1/transactions` by default (filterable via `type=initial_deposit`).

Full spec: `docs/openapi.yaml`

## Configuration

Required:
- `DB_PASSWORD`
- `JWT_SECRET` (min 32 chars)

Optional:
- `DB_HOST` (default `localhost`)
- `DB_PORT` (default `5432`)
- `DB_USER` (default `postgres`)
- `DB_NAME` (default `banking_platform`)
- `SERVER_PORT` (default `8080`)
- `JWT_EXPIRY_HOURS` (default `168`)
- `INITIAL_BALANCE_USD_CENTS` (default `100000`)
- `INITIAL_BALANCE_EUR_CENTS` (default `50000`)
- `CORS_ALLOW_ORIGIN` (comma-separated, default `*`)

Example:
```bash
export CORS_ALLOW_ORIGIN=http://localhost:3000,https://your-domain.com
```

## Database Migrations

Migrations are applied at backend startup via Goose.
Manual control:

```bash
cd backend
make migrate-up
make migrate-down
```

FX system accounts are seeded by migration `00006_create_fx.sql`.

## Testing

```bash
cd backend
go test -v ./internal/service/
```

Tests cover:
- transfers and insufficient funds
- exchange invariants
- ledger vs balance reconciliation
- concurrency behavior

## Known Limitations

- Fixed FX rate (no dynamic feeds)
- No email verification
- No roles/permissions
- JWT not revocable
- No rate limiting

## API Docs (Swagger UI)

```bash
docker run -p 8081:8080 -e SWAGGER_JSON=/api/openapi.yaml \
  -v $(pwd)/docs/openapi.yaml:/api/openapi.yaml \
  swaggerapi/swagger-ui
```

Open http://localhost:8081
