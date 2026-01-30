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

func (r *AccountRepository) FindFXAccountByCurrency(ctx context.Context, currency string) (*models.Account, error) {
    return r.FindByUserAndCurrency(ctx, models.FXSystemUserID, currency)
}
