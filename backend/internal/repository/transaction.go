package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"mini-banking-platform/internal/models"

	"github.com/jmoiron/sqlx"
)

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
		INSERT INTO ledger_entries (transaction_id, account_id, currency, amount_cents)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	err := tx.QueryRowContext(ctx, query, entry.TransactionID, entry.AccountID, entry.Currency, entry.AmountCents).
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

func (r *TransactionRepository) CreateExchangeSpread(ctx context.Context, tx *sqlx.Tx, spread *models.ExchangeSpread) error {
	query := `
		INSERT INTO exchange_spreads (transaction_id, residual_numerator, residual_denominator, target_currency)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	
	r.logger.Info("repository: attempting to create exchange spread",
		"transactionID", spread.TransactionID,
		"residualNumerator", spread.ResidualNumerator,
		"residualDenominator", spread.ResidualDenominator,
		"targetCurrency", spread.TargetCurrency)
	
	err := tx.QueryRowContext(ctx, query, 
		spread.TransactionID, 
		spread.ResidualNumerator, 
		spread.ResidualDenominator, 
		spread.TargetCurrency,
	).Scan(&spread.ID, &spread.CreatedAt)

	if err != nil {
		r.logger.Error("repository: failed to create exchange spread", 
			"error", err,
			"transactionID", spread.TransactionID,
			"residualNumerator", spread.ResidualNumerator,
			"residualDenominator", spread.ResidualDenominator,
			"targetCurrency", spread.TargetCurrency)
		return fmt.Errorf("repository: error creating exchange spread: %w", err)
	}

	r.logger.Info("repository: exchange spread recorded successfully", 
		"spreadID", spread.ID, 
		"transactionID", spread.TransactionID,
		"residual", fmt.Sprintf("%d/%d %s", spread.ResidualNumerator, spread.ResidualDenominator, spread.TargetCurrency))
	return nil
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

func (r *TransactionRepository) GetExchangeSpreadsByTransaction(ctx context.Context, transactionID string) (*models.ExchangeSpread, error) {
	var spread models.ExchangeSpread
	query := `
		SELECT id, transaction_id, residual_numerator, residual_denominator, target_currency, created_at
		FROM exchange_spreads
		WHERE transaction_id = $1
	`
	
	err := r.db.GetContext(ctx, &spread, query, transactionID)
	if err == sql.ErrNoRows {
		return nil, nil 
	}
	if err != nil {
		r.logger.Error("repository: failed to get exchange spread", "error", err, "transactionID", transactionID)
		return nil, fmt.Errorf("repository: error getting exchange spread: %w", err)
	}

	return &spread, nil
}

