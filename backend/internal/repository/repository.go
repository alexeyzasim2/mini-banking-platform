package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"mini-banking-platform/internal/errorsx"
	"mini-banking-platform/internal/models"

	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	User        *UserRepository
	Account     *AccountRepository
	Transaction *TransactionRepository
}

func NewRepositories(db *sqlx.DB, logger *slog.Logger) *Repositories {
	return &Repositories{
		User:        NewUserRepository(db, logger),
		Account:     NewAccountRepository(db, logger),
		Transaction: NewTransactionRepository(db, logger),
	}
}

type UserRepository struct {
	db     *sqlx.DB
	logger *slog.Logger
}

func NewUserRepository(db *sqlx.DB, logger *slog.Logger) *UserRepository {
	return &UserRepository{db: db, logger: logger}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User, hashedPassword string) error {
	query := `
		INSERT INTO users (email, password, first_name, last_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query, user.Email, hashedPassword, user.FirstName, user.LastName).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		r.logger.Error("repository: failed to create user", "error", err, "email", user.Email)
		return fmt.Errorf("repository: error creating user: %w", err)
	}

	r.logger.Info("repository: user created", "userID", user.ID, "email", user.Email)
	return nil
}

func (r *UserRepository) CreateInTx(ctx context.Context, tx *sqlx.Tx, user *models.User, hashedPassword string) error {
	query := `
		INSERT INTO users (email, password, first_name, last_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err := tx.QueryRowContext(ctx, query, user.Email, hashedPassword, user.FirstName, user.LastName).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		r.logger.Error("repository: failed to create user in tx", "error", err, "email", user.Email)
		return fmt.Errorf("repository: error creating user: %w", err)
	}

	r.logger.Info("repository: user created in tx", "userID", user.ID, "email", user.Email)
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, email, password, first_name, last_name, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errorsx.ErrUserNotFound
		}
		r.logger.Error("repository: failed to find user by email", "error", err, "email", email)
		return nil, fmt.Errorf("repository: error finding user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, email, password, first_name, last_name, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errorsx.ErrUserNotFound
		}
		r.logger.Error("repository: failed to find user by ID", "error", err, "userID", id)
		return nil, fmt.Errorf("repository: error finding user: %w", err)
	}

	return &user, nil
}

type AccountRepository struct {
	db     *sqlx.DB
	logger *slog.Logger
}

func NewAccountRepository(db *sqlx.DB, logger *slog.Logger) *AccountRepository {
	return &AccountRepository{db: db, logger: logger}
}

func (r *AccountRepository) Create(ctx context.Context, account *models.Account) error {
	query := `
		INSERT INTO accounts (user_id, currency, balance_cents)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query, account.UserID, account.Currency, account.BalanceCents).
		Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt)
	
	if err != nil {
		r.logger.Error("repository: failed to create account", "error", err, "userID", account.UserID)
		return fmt.Errorf("repository: error creating account: %w", err)
	}

	r.logger.Info("repository: account created", "accountID", account.ID, "currency", account.Currency)
	return nil
}

func (r *AccountRepository) CreateInTx(ctx context.Context, tx *sqlx.Tx, account *models.Account) error {
	query := `
		INSERT INTO accounts (user_id, currency, balance_cents)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	err := tx.QueryRowContext(ctx, query, account.UserID, account.Currency, account.BalanceCents).
		Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt)
	
	if err != nil {
		r.logger.Error("repository: failed to create account in tx", "error", err, "userID", account.UserID)
		return fmt.Errorf("repository: error creating account: %w", err)
	}

	r.logger.Info("repository: account created in tx", "accountID", account.ID, "currency", account.Currency)
	return nil
}

func (r *AccountRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		r.logger.Error("repository: failed to begin transaction", "error", err)
		return nil, fmt.Errorf("repository: error beginning transaction: %w", err)
	}
	return tx, nil
}

func (r *AccountRepository) FindByUserID(ctx context.Context, userID string) ([]models.Account, error) {
	var accounts []models.Account
	query := `
		SELECT id, user_id, currency, balance_cents, created_at, updated_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY currency
	`
	err := r.db.SelectContext(ctx, &accounts, query, userID)
	if err != nil {
		r.logger.Error("repository: failed to find accounts", "error", err, "userID", userID)
		return nil, fmt.Errorf("repository: error finding accounts: %w", err)
	}

	return accounts, nil
}

func (r *AccountRepository) FindByID(ctx context.Context, id string) (*models.Account, error) {
	var account models.Account
	query := `
		SELECT id, user_id, currency, balance_cents, created_at, updated_at
		FROM accounts
		WHERE id = $1
	`
	err := r.db.GetContext(ctx, &account, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errorsx.ErrAccountNotFound
		}
		r.logger.Error("repository: failed to find account", "error", err, "accountID", id)
		return nil, fmt.Errorf("repository: error finding account: %w", err)
	}

	return &account, nil
}

func (r *AccountRepository) FindByUserAndCurrency(ctx context.Context, userID, currency string) (*models.Account, error) {
	var account models.Account
	query := `
		SELECT id, user_id, currency, balance_cents, created_at, updated_at
		FROM accounts
		WHERE user_id = $1 AND currency = $2
	`
	err := r.db.GetContext(ctx, &account, query, userID, currency)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errorsx.ErrAccountNotFound
		}
		r.logger.Error("repository: failed to find account", "error", err, "userID", userID, "currency", currency)
		return nil, fmt.Errorf("repository: error finding account: %w", err)
	}

	return &account, nil
}

func (r *AccountRepository) UpdateBalanceCents(ctx context.Context, tx *sqlx.Tx, accountID string, amountCents int64) error {
	query := `
		UPDATE accounts
		SET balance_cents = balance_cents + $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	result, err := tx.ExecContext(ctx, query, amountCents, accountID)
	if err != nil {
		r.logger.Error("repository: failed to update balance", "error", err, "accountID", accountID)
		return fmt.Errorf("repository: error updating balance: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errorsx.ErrAccountNotFound
	}

	return nil
}

func (r *AccountRepository) GetBalanceCentsForUpdate(ctx context.Context, tx *sqlx.Tx, accountID string) (int64, error) {
	var balanceCents int64
	query := `SELECT balance_cents FROM accounts WHERE id = $1 FOR UPDATE`
	err := tx.GetContext(ctx, &balanceCents, query, accountID)
	if err != nil {
		r.logger.Error("repository: failed to get balance for update", "error", err, "accountID", accountID)
		return 0, fmt.Errorf("repository: error getting balance: %w", err)
	}
	return balanceCents, nil
}

type TransactionRepository struct {
	db     *sqlx.DB
	logger *slog.Logger
}

func NewTransactionRepository(db *sqlx.DB, logger *slog.Logger) *TransactionRepository {
	return &TransactionRepository{db: db, logger: logger}
}

func (r *TransactionRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		r.logger.Error("repository: failed to begin transaction", "error", err)
		return nil, fmt.Errorf("repository: error beginning transaction: %w", err)
	}
	return tx, nil
}

func (r *TransactionRepository) Create(ctx context.Context, tx *sqlx.Tx, transaction *models.Transaction) error {
	query := `
		INSERT INTO transactions (type, from_user_id, to_user_id, currency, amount_cents, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	err := tx.QueryRowContext(ctx, query,
		transaction.Type,
		transaction.FromUserID,
		transaction.ToUserID,
		transaction.Currency,
		transaction.AmountCents,
		transaction.Description,
	).Scan(&transaction.ID, &transaction.CreatedAt)

	if err != nil {
		r.logger.Error("repository: failed to create transaction", "error", err)
		return fmt.Errorf("repository: error creating transaction: %w", err)
	}

	r.logger.Info("repository: transaction created", "transactionID", transaction.ID, "type", transaction.Type)
	return nil
}

func (r *TransactionRepository) CreateInTx(ctx context.Context, tx *sqlx.Tx, transaction *models.Transaction) error {
	return r.Create(ctx, tx, transaction)
}

func (r *TransactionRepository) CreateLedgerEntry(ctx context.Context, tx *sqlx.Tx, entry *models.LedgerEntry) error {
	query := `
		INSERT INTO ledger_entries (transaction_id, account_id, amount_cents)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := tx.QueryRowContext(ctx, query, entry.TransactionID, entry.AccountID, entry.AmountCents).
		Scan(&entry.ID, &entry.CreatedAt)

	if err != nil {
		r.logger.Error("repository: failed to create ledger entry", "error", err)
		return fmt.Errorf("repository: error creating ledger entry: %w", err)
	}

	return nil
}

func (r *TransactionRepository) CreateLedgerEntryInTx(ctx context.Context, tx *sqlx.Tx, entry *models.LedgerEntry) error {
	return r.CreateLedgerEntry(ctx, tx, entry)
}

func (r *TransactionRepository) FindByUserID(ctx context.Context, userID string, transactionType string, page, limit int) ([]models.Transaction, int, error) {
	offset := (page - 1) * limit

	baseQuery := `
		SELECT id, type, from_user_id, to_user_id, currency, amount_cents, description, created_at
		FROM transactions
		WHERE (from_user_id = $1 OR to_user_id = $1)
	`
	countQuery := `
		SELECT COUNT(*)
		FROM transactions
		WHERE (from_user_id = $1 OR to_user_id = $1)
	`

	args := []interface{}{userID}
	argCount := 1

	if transactionType != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND type = $%d", argCount)
		countQuery += fmt.Sprintf(" AND type = $%d", argCount)
		args = append(args, transactionType)
	}

	baseQuery += " ORDER BY created_at DESC"
	
	argCount++
	baseQuery += fmt.Sprintf(" LIMIT $%d", argCount)
	args = append(args, limit)
	
	argCount++
	baseQuery += fmt.Sprintf(" OFFSET $%d", argCount)
	args = append(args, offset)

	var total int
	err := r.db.GetContext(ctx, &total, countQuery, args[:len(args)-2]...)
	if err != nil {
		r.logger.Error("repository: failed to count transactions", "error", err)
		return nil, 0, fmt.Errorf("repository: error counting transactions: %w", err)
	}

	var transactions []models.Transaction
	err = r.db.SelectContext(ctx, &transactions, baseQuery, args...)
	if err != nil {
		r.logger.Error("repository: failed to find transactions", "error", err, "userID", userID)
		return nil, 0, fmt.Errorf("repository: error finding transactions: %w", err)
	}

	return transactions, total, nil
}

func (r *TransactionRepository) GetLedgerSumCents(ctx context.Context, accountID string) (int64, error) {
	var sumCents sql.NullInt64
	query := `SELECT COALESCE(SUM(amount_cents), 0) FROM ledger_entries WHERE account_id = $1`
	
	err := r.db.GetContext(ctx, &sumCents, query, accountID)
	if err != nil {
		r.logger.Error("repository: failed to get ledger sum", "error", err, "accountID", accountID)
		return 0, fmt.Errorf("repository: error getting ledger sum: %w", err)
	}

	return sumCents.Int64, nil
}
