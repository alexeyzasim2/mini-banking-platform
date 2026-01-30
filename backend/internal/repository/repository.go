package repository

import (
	"log/slog"

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
