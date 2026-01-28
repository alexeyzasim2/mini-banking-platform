package service

import (
	"context"
	"log/slog"
	"mini-banking-platform/internal/http/dto"
	"mini-banking-platform/internal/jwt"
	"mini-banking-platform/internal/repository"
	"os"
	"testing"
)

func TestRegistration_Atomic(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repos := repository.NewRepositories(db, logger)
	jwtService := jwt.NewService("test-secret-key-that-is-long-enough-for-jwt", 168)
	service := NewAuthService(repos.User, repos.Account, repos.Transaction, jwtService, logger)

	req := dto.RegisterRequest{
		Email:     "newuser@test.com",
		Password:  "password123",
		FirstName: "New",
		LastName:  "User",
	}

	resp, err := service.Register(context.Background(), req, 100000, 50000)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	if resp.User.Email != req.Email {
		t.Errorf("Expected email %s, got %s", req.Email, resp.User.Email)
	}

	user, err := repos.User.FindByEmail(context.Background(), req.Email)
	if err != nil {
		t.Fatalf("User not found after registration: %v", err)
	}

	accounts, err := repos.Account.FindByUserID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("Failed to get user accounts: %v", err)
	}

	if len(accounts) != 2 {
		t.Errorf("Expected 2 accounts, got %d", len(accounts))
	}

	currencies := make(map[string]bool)
	for _, acc := range accounts {
		currencies[acc.Currency] = true
	}

	if !currencies["USD"] || !currencies["EUR"] {
		t.Error("Expected both USD and EUR accounts")
	}

	for _, acc := range accounts {
		if acc.Currency == "USD" && acc.BalanceCents != 100000 {
			t.Errorf("Expected USD balance 100000, got %d", acc.BalanceCents)
		}
		if acc.Currency == "EUR" && acc.BalanceCents != 50000 {
			t.Errorf("Expected EUR balance 50000, got %d", acc.BalanceCents)
		}
	}

	for _, acc := range accounts {
		var ledgerSum int64
		err := db.Get(&ledgerSum, `
			SELECT COALESCE(SUM(amount_cents), 0) 
			FROM ledger_entries 
			WHERE account_id = $1`, acc.ID)
		if err != nil {
			t.Fatalf("Failed to get ledger sum: %v", err)
		}

		if ledgerSum != acc.BalanceCents {
			t.Errorf("Ledger sum mismatch for %s account: balance=%d, ledger=%d", 
				acc.Currency, acc.BalanceCents, ledgerSum)
		}
	}
}

func TestRegistration_DuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repos := repository.NewRepositories(db, logger)
	jwtService := jwt.NewService("test-secret-key-that-is-long-enough-for-jwt", 168)
	service := NewAuthService(repos.User, repos.Account, repos.Transaction, jwtService, logger)

	req := dto.RegisterRequest{
		Email:     "duplicate@test.com",
		Password:  "password123",
		FirstName: "First",
		LastName:  "User",
	}

	_, err := service.Register(context.Background(), req, 100000, 50000)
	if err != nil {
		t.Fatalf("First registration failed: %v", err)
	}

	_, err = service.Register(context.Background(), req, 100000, 50000)
	if err == nil {
		t.Error("Expected error for duplicate email registration")
	}
}

func TestLogin_Success(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repos := repository.NewRepositories(db, logger)
	jwtService := jwt.NewService("test-secret-key-that-is-long-enough-for-jwt", 168)
	service := NewAuthService(repos.User, repos.Account, repos.Transaction, jwtService, logger)

	regReq := dto.RegisterRequest{
		Email:     "login@test.com",
		Password:  "password123",
		FirstName: "Login",
		LastName:  "User",
	}
	_, err := service.Register(context.Background(), regReq, 100000, 50000)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	loginReq := dto.LoginRequest{
		Email:    "login@test.com",
		Password: "password123",
	}
	resp, err := service.Login(context.Background(), loginReq)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if resp.User.Email != loginReq.Email {
		t.Errorf("Expected email %s, got %s", loginReq.Email, resp.User.Email)
	}

	if resp.Token == "" {
		t.Error("Expected non-empty token")
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repos := repository.NewRepositories(db, logger)
	jwtService := jwt.NewService("test-secret-key-that-is-long-enough-for-jwt", 168)
	service := NewAuthService(repos.User, repos.Account, repos.Transaction, jwtService, logger)

	regReq := dto.RegisterRequest{
		Email:     "invalid@test.com",
		Password:  "password123",
		FirstName: "Invalid",
		LastName:  "User",
	}
	_, err := service.Register(context.Background(), regReq, 100000, 50000)
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	loginReq := dto.LoginRequest{
		Email:    "invalid@test.com",
		Password: "wrongpassword",
	}
	_, err = service.Login(context.Background(), loginReq)
	if err == nil {
		t.Error("Expected error for invalid password")
	}
}

func TestLogin_NonExistentUser(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	repos := repository.NewRepositories(db, logger)
	jwtService := jwt.NewService("test-secret-key-that-is-long-enough-for-jwt", 168)
	service := NewAuthService(repos.User, repos.Account, repos.Transaction, jwtService, logger)

	loginReq := dto.LoginRequest{
		Email:    "nonexistent@test.com",
		Password: "password123",
	}
	_, err := service.Login(context.Background(), loginReq)
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

