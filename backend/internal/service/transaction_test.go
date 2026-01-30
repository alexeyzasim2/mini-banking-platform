package service

import (
	"context"
	"fmt"
	"log/slog"
	"mini-banking-platform/internal/errorsx"
	"mini-banking-platform/internal/http/dto"
	"mini-banking-platform/internal/models"
	"mini-banking-platform/internal/repository"
	"os"
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

var testDB *sqlx.DB


func setupTestDB(t *testing.T) *sqlx.DB {
	if testDB != nil {
		return testDB
	}

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "Kd8mPq2Ln5Rv9Xt3Yw7Bz")
	dbName := getEnv("DB_NAME", "banking_platform_test")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("Failed to set goose dialect: %v", err)
	}

	if err := goose.Up(db.DB, "../../migrations"); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	testDB = db
	return db
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func cleanupTestData(t *testing.T, db *sqlx.DB) {
	tables := []string{"exchange_spreads", "ledger_entries", "transactions", "accounts", "users"}
	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Logf("Warning: failed to clean %s: %v", table, err)
		}
	}
}

func createTestUser(t *testing.T, db *sqlx.DB, email string) *models.User {
	user := &models.User{
		Email:     email,
		FirstName: "Test",
		LastName:  "User",
	}

	query := `INSERT INTO users (email, password, first_name, last_name) 
	          VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	err := db.QueryRow(query, email, "hashedpass", user.FirstName, user.LastName).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}

func createTestAccount(t *testing.T, db *sqlx.DB, userID, currency string, balanceCents int64) *models.Account {
	account := &models.Account{
		UserID:       userID,
		Currency:     currency,
		BalanceCents: balanceCents,
	}

	query := `INSERT INTO accounts (user_id, currency, balance_cents) 
	          VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	err := db.QueryRow(query, userID, currency, balanceCents).
		Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		t.Fatalf("Failed to create test account: %v", err)
	}

	
	if balanceCents != 0 {
		txQuery := `INSERT INTO transactions (type, from_user_id, currency, amount_cents, description) 
		            VALUES ($1, $2, $3, $4, $5) RETURNING id`
		var txID string
		err = db.QueryRow(txQuery, "initial_deposit", userID, currency, balanceCents, "Test initial deposit").Scan(&txID)
		if err != nil {
			t.Fatalf("Failed to create initial transaction: %v", err)
		}

		ledgerQuery := `INSERT INTO ledger_entries (transaction_id, account_id, currency, amount_cents) VALUES ($1, $2, $3, $4)`
		_, err = db.Exec(ledgerQuery, txID, account.ID, currency, balanceCents)
		if err != nil {
			t.Fatalf("Failed to create initial ledger entry: %v", err)
		}
	}

	return account
}

func createFXSystemAccounts(t *testing.T, db *sqlx.DB) {
	userQuery := `INSERT INTO users (id, email, password, first_name, last_name) VALUES ($1, $2, $3, $4, $5)`
	_, err := db.Exec(userQuery, models.FXSystemUserID, models.FXSystemUserEmail, "N/A", "FX", "System")
    if err != nil {
        t.Fatalf("Failed to create FX system user: %v", err)
    }

    accountQuery := `INSERT INTO accounts (user_id, currency, balance_cents, allow_negative)
                     VALUES ($1, $2, $3, $4)`
    _, err = db.Exec(accountQuery, models.FXSystemUserID, models.CurrencyUSD, 0, true)
    if err != nil {
        t.Fatalf("Failed to create FX USD account: %v", err)
    }
    _, err = db.Exec(accountQuery, models.FXSystemUserID, models.CurrencyEUR, 0, true)
    if err != nil {
        t.Fatalf("Failed to create FX EUR account: %v", err)
    }
}

func TestTransfer_Success(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repos := repository.NewRepositories(db, logger)
	service := NewTransactionService(repos.Account, repos.Transaction, repos.User, logger)

	userA := createTestUser(t, db, "usera@test.com")
	userB := createTestUser(t, db, "userb@test.com")
	createTestAccount(t, db, userA.ID, "USD", 10000)
	createTestAccount(t, db, userB.ID, "USD", 0)

	req := dto.TransferRequest{
		ToUserID:    userB.Email,
		Currency:    "USD",
		AmountCents: 5000,
	}

	tx, err := service.Transfer(context.Background(), userA.ID, req)
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}

	if tx.Type != "transfer" {
		t.Errorf("Expected type transfer, got %s", tx.Type)
	}
	if tx.AmountCents != 5000 {
		t.Errorf("Expected amount 5000, got %d", tx.AmountCents)
	}

	var balanceA, balanceB int64
	db.Get(&balanceA, "SELECT balance_cents FROM accounts WHERE user_id = $1 AND currency = 'USD'", userA.ID)
	db.Get(&balanceB, "SELECT balance_cents FROM accounts WHERE user_id = $1 AND currency = 'USD'", userB.ID)

	if balanceA != 5000 {
		t.Errorf("Expected userA balance 5000, got %d", balanceA)
	}
	if balanceB != 5000 {
		t.Errorf("Expected userB balance 5000, got %d", balanceB)
	}
}

func TestTransfer_InsufficientFunds(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repos := repository.NewRepositories(db, logger)
	service := NewTransactionService(repos.Account, repos.Transaction, repos.User, logger)

	userA := createTestUser(t, db, "usera@test.com")
	userB := createTestUser(t, db, "userb@test.com")
	createTestAccount(t, db, userA.ID, "USD", 5000)
	createTestAccount(t, db, userB.ID, "USD", 0)

	req := dto.TransferRequest{
		ToUserID:    userB.Email,
		Currency:    "USD",
		AmountCents: 10000,
	}

	_, err := service.Transfer(context.Background(), userA.ID, req)
	if err != errorsx.ErrInsufficientFunds {
		t.Errorf("Expected ErrInsufficientFunds, got %v", err)
	}

	var balanceA int64
	db.Get(&balanceA, "SELECT balance_cents FROM accounts WHERE user_id = $1 AND currency = 'USD'", userA.ID)
	if balanceA != 5000 {
		t.Errorf("Expected userA balance unchanged at 5000, got %d", balanceA)
	}
}

func TestTransfer_LedgerBalance(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repos := repository.NewRepositories(db, logger)
	service := NewTransactionService(repos.Account, repos.Transaction, repos.User, logger)

	userA := createTestUser(t, db, "usera@test.com")
	userB := createTestUser(t, db, "userb@test.com")
	accountA := createTestAccount(t, db, userA.ID, "USD", 10000)
	accountB := createTestAccount(t, db, userB.ID, "USD", 0)

	req := dto.TransferRequest{
		ToUserID:    userB.Email,
		Currency:    "USD",
		AmountCents: 5000,
	}

	tx, err := service.Transfer(context.Background(), userA.ID, req)
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}

	var sum int64
	err = db.Get(&sum, `
		SELECT COALESCE(SUM(amount_cents), 0) 
		FROM ledger_entries 
		WHERE transaction_id = $1`, tx.ID)
	if err != nil {
		t.Fatalf("Failed to sum ledger entries: %v", err)
	}

	if sum != 0 {
		t.Errorf("Expected ledger entries to sum to 0, got %d", sum)
	}

	var ledgerSumA, ledgerSumB int64
	db.Get(&ledgerSumA, "SELECT COALESCE(SUM(amount_cents), 0) FROM ledger_entries WHERE account_id = $1", accountA.ID)
	db.Get(&ledgerSumB, "SELECT COALESCE(SUM(amount_cents), 0) FROM ledger_entries WHERE account_id = $1", accountB.ID)

	var balanceA, balanceB int64
	db.Get(&balanceA, "SELECT balance_cents FROM accounts WHERE id = $1", accountA.ID)
	db.Get(&balanceB, "SELECT balance_cents FROM accounts WHERE id = $1", accountB.ID)

	if balanceA != ledgerSumA {
		t.Errorf("Account A balance mismatch: balance=%d, ledger=%d", balanceA, ledgerSumA)
	}
	if balanceB != ledgerSumB {
		t.Errorf("Account B balance mismatch: balance=%d, ledger=%d", balanceB, ledgerSumB)
	}
}

func TestConcurrentTransfers(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repos := repository.NewRepositories(db, logger)
	service := NewTransactionService(repos.Account, repos.Transaction, repos.User, logger)

	userA := createTestUser(t, db, "usera@test.com")
	userB := createTestUser(t, db, "userb@test.com")
	createTestAccount(t, db, userA.ID, "USD", 10000)
	createTestAccount(t, db, userB.ID, "USD", 0)

	var wg sync.WaitGroup
	numTransfers := 5
	amountPerTransfer := int64(2000)

	for i := 0; i < numTransfers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := dto.TransferRequest{
				ToUserID:    userB.Email,
				Currency:    "USD",
				AmountCents: amountPerTransfer,
			}
			service.Transfer(context.Background(), userA.ID, req)
		}()
	}

	wg.Wait()

	var balanceA, balanceB int64
	db.Get(&balanceA, "SELECT balance_cents FROM accounts WHERE user_id = $1 AND currency = 'USD'", userA.ID)
	db.Get(&balanceB, "SELECT balance_cents FROM accounts WHERE user_id = $1 AND currency = 'USD'", userB.ID)

	expectedBalanceA := int64(0)
	expectedBalanceB := int64(10000)

	if balanceA != expectedBalanceA {
		t.Errorf("Expected userA balance %d, got %d", expectedBalanceA, balanceA)
	}
	if balanceB != expectedBalanceB {
		t.Errorf("Expected userB balance %d, got %d", expectedBalanceB, balanceB)
	}

	if balanceA+balanceB != 10000 {
		t.Errorf("Total money not conserved: %d + %d = %d", balanceA, balanceB, balanceA+balanceB)
	}
}

func TestExchange_Success(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repos := repository.NewRepositories(db, logger)
	service := NewTransactionService(repos.Account, repos.Transaction, repos.User, logger)

	createFXSystemAccounts(t, db)

	user := createTestUser(t, db, "user@test.com")
	createTestAccount(t, db, user.ID, "USD", 10000)
	createTestAccount(t, db, user.ID, "EUR", 0)

	req := dto.ExchangeRequest{
		FromCurrency: "USD",
		AmountCents:  10000,
	}

	tx, err := service.Exchange(context.Background(), user.ID, req)
	if err != nil {
		t.Fatalf("Exchange failed: %v", err)
	}

	if tx.Type != "exchange" {
		t.Errorf("Expected type exchange, got %s", tx.Type)
	}

	var balanceUSD, balanceEUR int64
	db.Get(&balanceUSD, "SELECT balance_cents FROM accounts WHERE user_id = $1 AND currency = 'USD'", user.ID)
	db.Get(&balanceEUR, "SELECT balance_cents FROM accounts WHERE user_id = $1 AND currency = 'EUR'", user.ID)

	if balanceUSD != 0 {
		t.Errorf("Expected USD balance 0, got %d", balanceUSD)
	}

	expectedEUR := int64(9200)
	if balanceEUR != expectedEUR {
		t.Errorf("Expected EUR balance %d, got %d", expectedEUR, balanceEUR)
	}

	// Verify double-entry bookkeeping: SUM(amount_cents) should equal 0 for the transaction
	// All amounts should be in the same currency (source currency) for proper balancing
	var totalSum int64
	db.Get(&totalSum, "SELECT COALESCE(SUM(amount_cents), 0) FROM ledger_entries WHERE transaction_id = $1", tx.ID)
	
	if totalSum != 0 {
		t.Errorf("Expected total sum of ledger entries to be 0 (double-entry), got %d", totalSum)
	}
}

func TestExchange_NoMoneyMinting(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repos := repository.NewRepositories(db, logger)
	service := NewTransactionService(repos.Account, repos.Transaction, repos.User, logger)

	createFXSystemAccounts(t, db)

	user := createTestUser(t, db, "user@test.com")
	createTestAccount(t, db, user.ID, "USD", 10000)
	createTestAccount(t, db, user.ID, "EUR", 0)

	req1 := dto.ExchangeRequest{
		FromCurrency: "USD",
		AmountCents:  10000,
	}
	_, err := service.Exchange(context.Background(), user.ID, req1)
	if err != nil {
		t.Fatalf("First exchange failed: %v", err)
	}

	var balanceEUR int64
	db.Get(&balanceEUR, "SELECT balance_cents FROM accounts WHERE user_id = $1 AND currency = 'EUR'", user.ID)

	req2 := dto.ExchangeRequest{
		FromCurrency: "EUR",
		AmountCents:  balanceEUR,
	}
	_, err = service.Exchange(context.Background(), user.ID, req2)
	if err != nil {
		t.Fatalf("Second exchange failed: %v", err)
	}

	var finalBalanceUSD, finalBalanceEUR int64
	db.Get(&finalBalanceUSD, "SELECT balance_cents FROM accounts WHERE user_id = $1 AND currency = 'USD'", user.ID)
	db.Get(&finalBalanceEUR, "SELECT balance_cents FROM accounts WHERE user_id = $1 AND currency = 'EUR'", user.ID)

	if finalBalanceUSD > 10000 {
		t.Errorf("Money minting detected! Started with 10000 USD, ended with %d USD", finalBalanceUSD)
	}

	if finalBalanceEUR != 0 {
		t.Errorf("Expected EUR balance 0 after round trip, got %d", finalBalanceEUR)
	}
}

func TestExchange_MinimumAmount(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)



	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repos := repository.NewRepositories(db, logger)
	service := NewTransactionService(repos.Account, repos.Transaction, repos.User, logger)

	createFXSystemAccounts(t, db)

	user := createTestUser(t, db, "user@test.com")
	createTestAccount(t, db, user.ID, "USD", 1000)
	createTestAccount(t, db, user.ID, "EUR", 0)

	req := dto.ExchangeRequest{
		FromCurrency: "USD",
		AmountCents:  5,
	}
	_, err := service.Exchange(context.Background(), user.ID, req)
	if err == nil {
		t.Error("Expected error for amount below minimum")
	}

	req.AmountCents = 10
	_, err = service.Exchange(context.Background(), user.ID, req)
	if err != nil {
		t.Errorf("Exchange at minimum amount should succeed: %v", err)
	}
}

func TestExchange_IntegerOverflowProtection(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)

	

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repos := repository.NewRepositories(db, logger)
	service := NewTransactionService(repos.Account, repos.Transaction, repos.User, logger)

	createFXSystemAccounts(t, db)

	user := createTestUser(t, db, "user@test.com")
	largeAmount := int64(9223372036854775807)
	createTestAccount(t, db, user.ID, "USD", largeAmount)
	createTestAccount(t, db, user.ID, "EUR", 0)

	req := dto.ExchangeRequest{
		FromCurrency: "USD",
		AmountCents:  largeAmount,
	}

	_, err := service.Exchange(context.Background(), user.ID, req)
	if err == nil {
		t.Error("Expected error for amount that would cause overflow")
	}
}

