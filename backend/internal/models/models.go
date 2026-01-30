package models

import (
	"time"
)

type User struct {
	ID        string    `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Password  string    `db:"password" json:"-"`
	FirstName string    `db:"first_name" json:"first_name"`
	LastName  string    `db:"last_name" json:"last_name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type Account struct {
	ID           string    `db:"id" json:"id"`
	UserID       string    `db:"user_id" json:"user_id"`
	Currency     string    `db:"currency" json:"currency"`
	BalanceCents int64     `db:"balance_cents" json:"balance_cents"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type Transaction struct {
	ID          string    `db:"id" json:"id"`
	Type        string    `db:"type" json:"type"`
	FromUserID  string    `db:"from_user_id" json:"from_user_id"`
	ToUserID    *string   `db:"to_user_id" json:"to_user_id,omitempty"`
	Currency    string    `db:"currency" json:"currency"`
	AmountCents int64     `db:"amount_cents" json:"amount_cents"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type LedgerEntry struct {
	ID            string    `db:"id" json:"id"`
	TransactionID string    `db:"transaction_id" json:"transaction_id"`
	AccountID     string    `db:"account_id" json:"account_id"`
	Currency      string    `db:"currency" json:"currency"`
	AmountCents   int64     `db:"amount_cents" json:"amount_cents"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}
