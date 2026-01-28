# AI Usage Documentation

This document records the use of AI/LLM tools during the development of this project.

## Overview

AI tools were used for frontend components, unit tests, and API documentation. The backend business logic was implemented manually without AI assistance to ensure deep understanding of financial integrity requirements.

---

## AI-1: Initial React Project Structure

**Purpose:** Set up React 19 project structure with TypeScript and Tailwind CSS configuration.

**Tool & Model:** Claude Sonnet 4.5

**Prompt:**
```
I need to set up a React 19 project with TypeScript and Tailwind CSS. Requirements:
- Use Create React App as base
- Configure Tailwind CSS 3.x with PostCSS
- Set up TypeScript with strict mode
- Create basic folder structure: components/, pages/, hooks/, services/, types/
- Configure axios for API client
- Add React Router v6 for navigation

Provide the necessary configuration files (package.json, tailwind.config.js, postcss.config.js, tsconfig.json) and basic folder structure.
```

**How the response was used:**
Used configuration files as starting point. Modified package.json to include specific versions (React 19.2.3). Adapted Tailwind config to project needs. The folder structure was used as-is.

---

## AI-2: Authentication Context and Hook

**Purpose:** Implement React Context API for authentication state management.

**Tool & Model:** GPT-5

**Prompt:**
```
I need to implement JWT authentication in React using Context API. Requirements:

1. AuthContext that manages:
   - User state (User | null)
   - Loading state during initial token validation
   - Login function (email, password) -> calls POST /auth/login
   - Register function (email, password, first_name, last_name) -> calls POST /auth/register
   - Logout function (clears token from localStorage)

2. On mount, check localStorage for token:
   - If exists, validate by calling GET /auth/me
   - If invalid, clear token and set user to null

3. AuthProvider component that wraps app
4. useAuth custom hook for consuming context

Use TypeScript with proper types. API calls should use axios instance from services/api.ts.
Don't include the API service implementation, just the auth context/hook.
```

**How the response was used:**
Used the structure as-is with minor modifications to error handling. Added try-catch blocks for better error management. The localStorage pattern and initial token validation logic were adopted directly.

---

## AI-3: Protected Route Component

**Purpose:** Create route wrapper for authentication-required pages.

**Tool & Model:** Claude Sonnet 4.5

**Prompt:**
```
Using React Router v6, create a ProtectedRoute component that:
1. Uses useAuth hook to check if user is authenticated
2. Shows loading state while checking authentication
3. Redirects to /login if user is not authenticated
4. Renders children (wrapped page) if authenticated
5. Uses Navigate component for redirects

Component should use TypeScript and accept children as prop.
```

**How the response was used:**
Integrated directly into App.tsx. The loading state pattern was particularly useful. No significant modifications needed.

---

## AI-4: Dashboard Page Layout

**Purpose:** Create main dashboard page with balance display and transaction forms.

**Tool & Model:** GPT-5

**Prompt:**
```
Create a React Dashboard page component with the following layout:

1. Top section: Display user's account balances
   - Two cards side by side (USD and EUR accounts)
   - Show currency symbol, account type, and balance
   - Use Tailwind CSS for styling with cards and grid layout

2. Middle section: Two forms in grid (Transfer and Exchange)
   - Each form in its own card
   - Forms should be the same height (use flexbox)
   - Buttons aligned to bottom of each form

3. Bottom section: Recent transactions (last 5)
   - Display in table or list format
   - Show transaction type, amount, description, date

Use TypeScript, fetch data with axios on component mount.
API endpoints:
- GET /accounts -> returns { accounts: Account[] }
- GET /transactions?limit=5 -> returns { transactions: Transaction[] }

Don't implement the forms yet, just placeholders.
Use proper loading states and error handling.
```

**How the response was used:**
Used the layout structure and grid configuration. Modified to include actual form components (TransferForm, ExchangeForm) instead of placeholders. The loading state pattern was adopted. Changed data fetching to use Promise.all for concurrent requests.

---

## AI-5: Transfer Form Component

**Purpose:** Implement form for transferring money between users.

**Tool & Model:** Claude Sonnet 4.5

**Prompt:**
```
Create a TransferForm React component with TypeScript for money transfers:

Form fields:
1. Recipient email (text input, type="email", required)
2. Currency selector (dropdown: USD or EUR)
3. Amount (number input, step="0.01", min="0.01", required)
4. Submit button

Requirements:
- Use controlled inputs with useState
- Show loading state during submission (disable button)
- Display error messages if API call fails
- Call onSuccess callback after successful transfer
- API endpoint: POST /transactions/transfer
- Request body: { to_user_id: string, currency: string, amount: number }

Use Tailwind CSS for styling. Make form responsive.
Add proper validation (required fields, minimum amount).
```

**How the response was used:**
Used as base implementation. Added confirmation modal before submission (not in original prompt). Modified recipient field to accept both email and user ID. Added helper text showing available test users.

---

## AI-6: Exchange Form Component

**Purpose:** Implement currency exchange form with live rate display.

**Tool & Model:** GPT-5

**Prompt:**
```
Create ExchangeForm React component for currency exchange (USD <-> EUR):

Form fields:
1. From currency selector (USD or EUR)
2. Amount input (number, step="0.01", min="0.01", required)
3. Display section showing:
   - Exchange rate (1 USD = 0.92 EUR)
   - Calculated conversion amount (live update as user types)

Requirements:
- When USD selected, show EUR conversion and vice versa
- Add swap button to quickly reverse direction
- Calculate conversion on every amount change
- Show loading state during submission
- Display error messages
- Call onSuccess callback after successful exchange
- API endpoint: POST /transactions/exchange
- Request body: { from_currency: string, amount: number }

Use TypeScript and Tailwind CSS. The "to_currency" is determined automatically (USD->EUR or EUR->USD).
```

**How the response was used:**
Adopted the structure and live calculation logic. Added confirmation modal before submission. Modified to show exchange rate in two places (form and preview). The swap button functionality was used as-is.

---

## AI-7: Transaction History Page with Filters

**Purpose:** Create page showing all transactions with type filter and pagination.

**Tool & Model:** Claude Sonnet 4.5

**Prompt:**
```
Create a Transactions page component in React with TypeScript:

Features:
1. Filter dropdown:
   - Options: All, Transfer, Exchange
   - When changed, reload transactions with ?type=<value>

2. Transaction table/list:
   - Columns: Type (badge), Amount (with currency), Description, Date
   - Use color coding: blue for transfers, green for exchanges

3. Pagination controls:
   - Previous and Next buttons
   - Show current page number
   - Disable Previous on page 1
   - Disable Next when results < limit

API endpoint: GET /transactions?type=<type>&page=<page>&limit=10
Response: { transactions: Transaction[], page: number, limit: number, total: number }

Use useEffect to reload when filter or page changes.
Use useState for filter, page, transactions, and loading state.
Apply filters automatically without separate "Apply" button.
```

**How the response was used:**
Used as foundation. Fixed the useEffect dependencies issue by wrapping loadTransactions in useCallback. The automatic filter application pattern was adopted directly. Modified table styling for better responsiveness.

---

## AI-8: Confirmation Modal Component

**Purpose:** Create reusable modal for transaction confirmation.

**Tool & Model:** GPT-5

**Prompt:**
```
Create a reusable ConfirmationModal React component with TypeScript:

Props:
- isOpen: boolean
- title: string
- message: string | ReactNode (to support formatted content)
- onConfirm: () => void
- onCancel: () => void
- confirmText?: string (default "Confirm")
- cancelText?: string (default "Cancel")
- confirmColor?: string (Tailwind classes for button color)

Features:
- Fixed position overlay (backdrop with opacity)
- Centered modal card with shadow
- Title at top
- Message in middle (can be multi-line or formatted)
- Two buttons at bottom: Cancel (gray) and Confirm (customizable color)
- Click outside modal to close (calls onCancel)
- Accessibility: focus trap, ESC key to close

Use Tailwind CSS for styling. Modal should be responsive.
```

**How the response was used:**
Used structure directly. Simplified accessibility features (removed ESC key handler and focus trap for time constraints). The backdrop click-to-close was implemented as suggested.

---

## AI-9: Transaction Detail Modal / Receipt

**Purpose:** Create modal to display detailed transaction information as receipt.

**Tool & Model:** Claude Sonnet 4.5

**Prompt:**
```
Create a TransactionDetailModal component to show transaction receipt:

Props:
- isOpen: boolean
- transaction: Transaction | null
- onClose: () => void

Display sections:
1. Header: "Transaction Receipt" + short ID (first 8 chars)
2. Details (with visual separators):
   - Type: badge (blue for transfer, green for exchange)
   - Amount: large bold text with currency
   - Date & Time: formatted (use Intl.DateTimeFormat or toLocaleString)
   - Exchange Rate: only show for exchange transactions (1 USD = 0.92 EUR)
   - Description: transaction description text
   - Status: "Completed" badge (always completed)
   - Full Transaction ID: monospace font

3. Footer:
   - Close button (primary color)
   - Small text: "This receipt is for your records"

Use fixed overlay like confirmation modal. Style with Tailwind CSS.
Make it look like a real receipt with proper spacing and hierarchy.
```

**How the response was used:**
Used as-is with minor styling adjustments. Added logic to determine exchange rate direction based on transaction description. Integrated into TransactionList component with click handlers.

---

## AI-10: Registration Page

**Purpose:** Create user registration page with form.

**Tool & Model:** GPT-5

**Prompt:**
```
Create a Register page component for user registration:

Form fields (all required):
1. First Name (text input)
2. Last Name (text input)
3. Email (email input)
4. Password (password input, minLength 6)

Requirements:
- Similar layout to Login page (centered card)
- Show loading state during registration
- Display error messages from API
- Call register function from useAuth hook
- After successful registration, redirect to dashboard (/)
- Add link to Login page: "Already have an account? Sign in"

API call handled by useAuth context, just call register({ email, password, first_name, last_name })
Use TypeScript and Tailwind CSS matching Login page style.
```

**How the response was used:**
Used structure directly. The navigation flow and form validation were adopted as-is. Modified styling slightly to match exact Login page appearance.

---

## AI-11: Unit Tests for Exchange Logic

**Purpose:** Write comprehensive unit tests for currency exchange operations to prevent financial bugs.

**Tool & Model:** Claude Sonnet 4.5

**Prompt:**
```
Write Go unit tests for currency exchange service to verify financial integrity:

Critical test cases:
1. TestExchange_NoMoneyMinting: Verify round-trip exchange (USD->EUR->USD) doesn't create money
   - Start with $100 USD
   - Exchange to EUR (should get €92)
   - Exchange back to USD
   - Verify balance is exactly $100 (no profit/loss)
   - Test with multiple amounts: small ($0.10), medium ($100), large ($10000)

2. TestExchange_LedgerBalance: Verify ledger entries are balanced
   - Perform exchange
   - Sum all ledger entries for that transaction
   - Verify sum equals zero (double-entry principle)

3. TestExchange_IntegerOverflowProtection: Test overflow handling
   - Attempt exchange with amount near math.MaxInt64
   - Should return error before overflow
   - Verify no database changes occurred

4. TestExchange_MinimumAmount: Verify minimum exchange enforcement
   - Attempt exchange with $0.05 (below minimum)
   - Should return ErrInvalidAmount
   - Attempt with $0.10 (at minimum) should succeed

5. TestExchange_PreventsPrecisionLoss: Test repeated small exchanges
   - Perform 100 small exchanges (e.g., $1.00 each)
   - Verify total ledger entries match total balance changes
   - Check no cumulative rounding errors

Use testify/assert for assertions. Set up test database with proper cleanup.
Mock dependencies where appropriate. Use table-driven tests for multiple scenarios.
```

**How the response was used:**
Used the test structure and several key scenarios as foundation. Implemented tests for successful exchange, no money minting on round-trip exchanges, minimum amount enforcement, and integer overflow protection. Some of the more advanced scenarios (like long repeated small exchanges) were kept as future ideas and were not fully implemented.

---

## AI-12: Unit Tests for Concurrent Operations

**Purpose:** Test row-level locking prevents race conditions.

**Tool & Model:** GPT-5

**Prompt:**
```
Write Go unit test for concurrent transfers to verify database locking:

TestConcurrentTransfers:
1. Create 2 users with known balances (User1: $100, User2: $0)
2. Launch 5 goroutines simultaneously
3. Each goroutine transfers $20 from User1 to User2
4. Use WaitGroup to synchronize
5. After all complete, verify:
   - User1 balance: $0 (100 - 5*20)
   - User2 balance: $100 (0 + 5*20)
   - No race conditions (one transfer should fail due to insufficient funds)

The test should pass if row-level locking (FOR UPDATE) works correctly.
If locking fails, balance will be incorrect due to race conditions.

Use Go's sync.WaitGroup and goroutines. Test should complete within 10 seconds.
```

**How the response was used:**
Adopted the test structure and goroutine pattern. Modified to expect exactly 5 successful transfers (since balance is sufficient). Added better error handling to track which transfers succeeded/failed. The WaitGroup usage was used as-is.

---

## AI-13: Unit Tests for Atomic Registration

**Purpose:** Verify user and account creation happens atomically.

**Tool & Model:** Claude Sonnet 4.5

**Prompt:**
```
Write Go unit test to verify atomic user registration:

TestRegistration_Atomic:
1. Mock scenario where account creation fails mid-transaction
2. Attempt to register user with email "test@example.com"
3. Verify:
   - User is NOT created in database (transaction rolled back)
   - No orphaned accounts exist
   - Can retry registration with same email (no "already exists" error)

Alternative approach (if mocking is complex):
1. Register user normally
2. Verify user exists
3. Verify exactly 2 accounts created (USD and EUR)
4. Verify accounts have correct initial balances
5. Query database: user and accounts created in same transaction (check timestamps)

Use transaction testing patterns. Verify rollback behavior.
```

**How the response was used:**
Implemented the alternative approach (simpler and more reliable). Added verification of user existence, account count, currencies, and balances. Did not implement mock failure scenario due to complexity. The basic atomic verification was sufficient.

---

## AI-14: Additional Financial Tests

**Purpose:** Generate additional test cases for edge cases and validation.

**Tool & Model:** GPT-5

**Prompt:**
```
Generate additional Go unit tests for banking operations:

1. TestTransfer_InsufficientFunds:
   - User has $50, attempts to transfer $100
   - Should return ErrInsufficientFunds
   - Verify no database changes

2. TestTransfer_OwnershipValidation:
   - User A attempts to transfer from User B's account
   - Should return ownership error
   - Verify account ownership checks work

3. TestReconciliation_MatchesLedger:
   - Create transactions
   - Call ReconcileBalances endpoint
   - Verify account.balance == SUM(ledger_entries.amount) for each account
   - Verify IsBalanced flag is true

4. TestTransfer_LedgerBalance:
   - Similar to exchange test
   - Verify transfer creates balanced ledger entries
   - Sum of entries should equal zero

Write these tests following Go testing conventions and using the same patterns as previous tests.
```

**How the response was used:**
Implemented most of the suggested tests with modifications. `TestReconciliation_MatchesLedger` required adding initial ledger entries to match seeded balances, and ledger balance tests needed SQL fixes for joining tables. Some scenarios (like an explicit ownership validation test) are covered indirectly by service logic and were left as potential future work rather than separate dedicated tests.

---

## AI-15: OpenAPI Documentation

**Purpose:** Generate OpenAPI 3.0 specification for REST API.

**Tool & Model:** Claude Sonnet 4.5

**Prompt:**
```
Generate OpenAPI 3.0 specification (YAML) for banking REST API:

Endpoints to document:

Authentication:
- POST /api/v1/auth/login: {email, password} -> {token, user}
- POST /api/v1/auth/register: {email, password, first_name, last_name} -> {token, user}
- GET /api/v1/auth/me: -> {user} (requires Bearer token)

Accounts:
- GET /api/v1/accounts: -> {accounts: [...]} (requires auth)
- GET /api/v1/accounts/{id}/balance: -> {account_id, currency, balance} (requires auth)
- GET /api/v1/accounts/reconcile: -> {results: [...]} (requires auth)

Transactions:
- POST /api/v1/transactions/transfer: {to_user_id, currency, amount} -> {message, transaction} (requires auth)
- POST /api/v1/transactions/exchange: {from_currency, amount} -> {message, transaction} (requires auth)
- GET /api/v1/transactions: ?type&page&limit -> {transactions, page, limit, total} (requires auth)

Health:
- GET /health: -> {status: "ok"}

Include:
- Complete schemas for User, Account, Transaction, LedgerEntry
- Security scheme: Bearer token (JWT)
- Request/response examples
- Error responses (400, 401, 404, 500)
- Description for each endpoint
- Tags for grouping (Authentication, Accounts, Transactions)

Format: OpenAPI 3.0.0 YAML
```

**How the response was used:**
Used as complete documentation with minor additions. Added more detailed descriptions for complex operations (exchange calculation, ledger system). Included example values in schemas. The security scheme and error responses were used exactly as generated.

---

## AI-16: README Documentation Sections

**Purpose:** Generate comprehensive README sections for architecture and design decisions.

**Tool & Model:** GPT-5

**Prompt:**
```
Write README sections explaining the double-entry ledger system and architecture:

Section 1: Double-Entry Ledger Design
- Explain what double-entry bookkeeping is
- Show database schema (users, accounts, transactions, ledger_entries)
- Provide transaction flow examples (Transfer and Exchange)
- Explain why this approach ensures financial integrity
- Include SQL snippets or pseudo-code

Section 2: Balance Consistency Approach
- Explain the problem: keeping accounts.balance in sync with ledger
- Describe the solution:
  1. Atomic transactions (all-or-nothing)
  2. Row-level locking (FOR UPDATE)
  3. Reconciliation endpoint
- Show code examples
- Mention testing approach

Section 3: Design Decisions & Trade-offs
Cover these decisions:
- Integer arithmetic for exchange (why? trade-off?)
- Minimum exchange amount (why? trade-off?)
- Fixed exchange rate (why? trade-off?)
- Monorepo structure (why? trade-off?)
- Pre-seeded users + registration (why? trade-off?)

For each decision:
- What was decided
- Why this approach
- What was traded off
- Any limitations

Use technical but accessible language. Include code examples where helpful.
Format in Markdown for README.
```

**How the response was used:**
Used structure and explanations as base. Expanded with specific implementation details from the codebase. Added more technical depth to integer arithmetic explanation. The trade-offs section was particularly useful and used with minor modifications. Added SQL schema examples manually.

---

## Summary

Total AI interactions: 16

Breakdown by category:
- Frontend components: 10 interactions (AI-1 through AI-10)
- Unit tests: 5 interactions (AI-11 through AI-15)
- Documentation: 2 interactions (AI-15, AI-16)

