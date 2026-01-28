package service

import (
	"context"
	"fmt"
	"log/slog"
	"mini-banking-platform/internal/errorsx"
	"mini-banking-platform/internal/models"
	"mini-banking-platform/internal/repository"
)

type AccountService struct {
	accountRepo     *repository.AccountRepository
	transactionRepo *repository.TransactionRepository
	logger          *slog.Logger
}

func NewAccountService(accountRepo *repository.AccountRepository, transactionRepo *repository.TransactionRepository, logger *slog.Logger) *AccountService {
	return &AccountService{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		logger:          logger,
	}
}

func (s *AccountService) GetUserAccounts(ctx context.Context, userID string) ([]models.Account, error) {
	accounts, err := s.accountRepo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get user accounts", "error", err, "userID", userID)
		return nil, fmt.Errorf("error getting accounts: %w", err)
	}
	return accounts, nil
}

func (s *AccountService) GetAccountBalance(ctx context.Context, userID, accountID string) (*models.Account, error) {
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		s.logger.Error("failed to get account balance", "error", err, "accountID", accountID)
		return nil, err
	}

	if account.UserID != userID {
		s.logger.Warn("unauthorized account access", "userID", userID, "accountID", accountID)
		return nil, errorsx.ErrAccountNotFound
	}

	return account, nil
}

type ReconciliationResult struct {
	AccountID      string `json:"account_id"`
	Currency       string `json:"currency"`
	BalanceCents   int64  `json:"balance_cents"`
	LedgerSumCents int64  `json:"ledger_sum_cents"`
	DifferenceCents int64  `json:"difference_cents"`
	IsBalanced     bool   `json:"is_balanced"`
}

func (s *AccountService) ReconcileBalances(ctx context.Context, userID string) ([]ReconciliationResult, error) {
	accounts, err := s.accountRepo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get user accounts for reconciliation", "error", err, "userID", userID)
		return nil, fmt.Errorf("error getting accounts: %w", err)
	}

	results := make([]ReconciliationResult, 0, len(accounts))

	for _, account := range accounts {
		ledgerSumCents, err := s.transactionRepo.GetLedgerSumCents(ctx, account.ID)
		if err != nil {
			s.logger.Error("failed to get ledger sum", "error", err, "accountID", account.ID)
			return nil, fmt.Errorf("error getting ledger sum: %w", err)
		}

		differenceCents := account.BalanceCents - ledgerSumCents
		isBalanced := differenceCents == 0

		results = append(results, ReconciliationResult{
			AccountID:       account.ID,
			Currency:        account.Currency,
			BalanceCents:    account.BalanceCents,
			LedgerSumCents:  ledgerSumCents,
			DifferenceCents: differenceCents,
			IsBalanced:      isBalanced,
		})

		if !isBalanced {
			s.logger.Warn("balance mismatch detected",
				"accountID", account.ID,
				"currency", account.Currency,
				"balanceCents", account.BalanceCents,
				"ledgerSumCents", ledgerSumCents,
				"differenceCents", differenceCents,
			)
		}
	}

	s.logger.Info("balance reconciliation completed", "userID", userID, "accountsChecked", len(results))
	return results, nil
}

