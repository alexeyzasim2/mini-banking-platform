package service

import (
	"context"
	"log/slog"
	"mini-banking-platform/internal/repository"
	"os"
	"testing"
)

func TestReconciliation_MatchesLedger(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repos := repository.NewRepositories(db, logger)
	accountService := NewAccountService(repos.Account, repos.Transaction, logger)

	user := createTestUser(t, db, "reconcile@test.com")
	createTestAccount(t, db, user.ID, "USD", 100000)
	createTestAccount(t, db, user.ID, "EUR", 50000)

	results, err := accountService.ReconcileBalances(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("Reconciliation failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 reconciliation results, got %d", len(results))
	}

	for _, result := range results {
		if !result.IsBalanced {
			t.Errorf("Account %s not balanced: balance=%d, ledger=%d, diff=%d",
				result.Currency, result.BalanceCents, result.LedgerSumCents, result.DifferenceCents)
		}

		if result.BalanceCents != result.LedgerSumCents {
			t.Errorf("Balance mismatch for %s: balance=%d, ledger=%d",
				result.Currency, result.BalanceCents, result.LedgerSumCents)
		}
	}
}

func TestGetUserAccounts(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repos := repository.NewRepositories(db, logger)
	service := NewAccountService(repos.Account, repos.Transaction, logger)

	user := createTestUser(t, db, "accounts@test.com")
	createTestAccount(t, db, user.ID, "USD", 100000)
	createTestAccount(t, db, user.ID, "EUR", 50000)

	accounts, err := service.GetUserAccounts(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("Failed to get user accounts: %v", err)
	}

	if len(accounts) != 2 {
		t.Errorf("Expected 2 accounts, got %d", len(accounts))
	}

	currencies := make(map[string]int64)
	for _, acc := range accounts {
		currencies[acc.Currency] = acc.BalanceCents
	}

	if currencies["USD"] != 100000 {
		t.Errorf("Expected USD balance 100000, got %d", currencies["USD"])
	}
	if currencies["EUR"] != 50000 {
		t.Errorf("Expected EUR balance 50000, got %d", currencies["EUR"])
	}
}

func TestGetAccountBalance_Success(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repos := repository.NewRepositories(db, logger)
	service := NewAccountService(repos.Account, repos.Transaction, logger)

	user := createTestUser(t, db, "balance@test.com")
	account := createTestAccount(t, db, user.ID, "USD", 100000)

	result, err := service.GetAccountBalance(context.Background(), user.ID, account.ID)
	if err != nil {
		t.Fatalf("Failed to get account balance: %v", err)
	}

	if result.BalanceCents != 100000 {
		t.Errorf("Expected balance 100000, got %d", result.BalanceCents)
	}
}

func TestGetAccountBalance_UnauthorizedAccess(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repos := repository.NewRepositories(db, logger)
	service := NewAccountService(repos.Account, repos.Transaction, logger)

	userA := createTestUser(t, db, "usera@test.com")
	userB := createTestUser(t, db, "userb@test.com")
	accountA := createTestAccount(t, db, userA.ID, "USD", 100000)

	_, err := service.GetAccountBalance(context.Background(), userB.ID, accountA.ID)
	if err == nil {
		t.Error("Expected error when accessing another user's account")
	}
}

