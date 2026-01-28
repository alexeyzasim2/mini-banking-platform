package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"mini-banking-platform/internal/errorsx"
	"mini-banking-platform/internal/http/dto"
	"mini-banking-platform/internal/models"
	"mini-banking-platform/internal/repository"
	"strings"
)

type TransactionService struct {
	accountRepo     *repository.AccountRepository
	transactionRepo *repository.TransactionRepository
	userRepo        *repository.UserRepository
	logger          *slog.Logger
}

func NewTransactionService(
	accountRepo *repository.AccountRepository,
	transactionRepo *repository.TransactionRepository,
	userRepo *repository.UserRepository,
	logger *slog.Logger,
) *TransactionService {
	return &TransactionService{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
		logger:          logger,
	}
}

func (s *TransactionService) Transfer(ctx context.Context, fromUserID string, req dto.TransferRequest) (*models.Transaction, error) {
	amountCents := req.AmountCents
	if amountCents <= 0 {
		return nil, errorsx.ErrInvalidAmount
	}
	if req.Currency != "USD" && req.Currency != "EUR" {
		return nil, errorsx.ErrInvalidCurrency
	}

	var toUser *models.User
	var err error
	if strings.Contains(req.ToUserID, "@") {
		toUser, err = s.userRepo.FindByEmail(ctx, req.ToUserID)
	} else {
		toUser, err = s.userRepo.FindByID(ctx, req.ToUserID)
	}
	if err != nil {
		return nil, errorsx.ErrUserNotFound
	}

	if fromUserID == toUser.ID {
		return nil, errorsx.ErrCannotTransferToSelf
	}

	tx, err := s.transactionRepo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	fromAccount, err := s.accountRepo.FindByUserAndCurrency(ctx, fromUserID, req.Currency)
	if err != nil {
		return nil, err
	}

	balanceCents, err := s.accountRepo.GetBalanceCentsForUpdate(ctx, tx, fromAccount.ID)
	if err != nil {
		return nil, err
	}
	if balanceCents < amountCents {
		s.logger.Warn("insufficient funds", "userID", fromUserID, "available", balanceCents, "required", amountCents)
		return nil, errorsx.ErrInsufficientFunds
	}

	toAccount, err := s.accountRepo.FindByUserAndCurrency(ctx, toUser.ID, req.Currency)
	if err != nil {
		return nil, err
	}

	if _, err := s.accountRepo.GetBalanceCentsForUpdate(ctx, tx, toAccount.ID); err != nil {
		return nil, err
	}

	transaction := &models.Transaction{
		Type:        "transfer",
		FromUserID:  fromUserID,
		ToUserID:    &toUser.ID,
		Currency:    req.Currency,
		AmountCents: amountCents,
		Description: fmt.Sprintf("Transfer to %s %s", toUser.FirstName, toUser.LastName),
	}

	if err := s.transactionRepo.Create(ctx, tx, transaction); err != nil {
		return nil, err
	}

	debitEntry := &models.LedgerEntry{
		TransactionID: transaction.ID,
		AccountID:     fromAccount.ID,
		AmountCents:   -amountCents,
	}
	if err := s.transactionRepo.CreateLedgerEntry(ctx, tx, debitEntry); err != nil {
		return nil, err
	}

	creditEntry := &models.LedgerEntry{
		TransactionID: transaction.ID,
		AccountID:     toAccount.ID,
		AmountCents:   amountCents,
	}
	if err := s.transactionRepo.CreateLedgerEntry(ctx, tx, creditEntry); err != nil {
		return nil, err
	}

	if err := s.accountRepo.UpdateBalanceCents(ctx, tx, fromAccount.ID, -amountCents); err != nil {
		return nil, err
	}
	if err := s.accountRepo.UpdateBalanceCents(ctx, tx, toAccount.ID, amountCents); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("failed to commit transfer", "error", err)
		return nil, fmt.Errorf("error committing transfer: %w", err)
	}

	s.logger.Info("transfer completed", "transactionID", transaction.ID, "from", fromUserID, "to", req.ToUserID, "amountCents", amountCents)
	return transaction, nil
}

func (s *TransactionService) Exchange(ctx context.Context, userID string, req dto.ExchangeRequest) (*models.Transaction, error) {
	fromAmountCents := req.AmountCents
	if fromAmountCents <= 0 {
		return nil, errorsx.ErrInvalidAmount
	}

	if fromAmountCents < 10 {
		return nil, errorsx.BadRequest("minimum exchange amount is 10 cents ($0.10 or €0.10)")
	}

	fromCurrency := req.FromCurrency
	var toCurrency string
	var rateNum, rateDenom int64

	if fromCurrency == "USD" {
		toCurrency = "EUR"
		rateNum = 23
		rateDenom = 25
	} else if fromCurrency == "EUR" {
		toCurrency = "USD"
		rateNum = 25
		rateDenom = 23
	} else {
		return nil, errorsx.ErrInvalidCurrency
	}

	if fromCurrency == toCurrency {
		return nil, errorsx.ErrCurrenciesMustDiffer
	}

	maxSafeAmount := int64(math.MaxInt64 / rateNum)
	if fromAmountCents > maxSafeAmount {
		s.logger.Error("exchange amount too large, would cause overflow",
			"fromAmountCents", fromAmountCents, "maxSafe", maxSafeAmount)
		return nil, errorsx.BadRequest("amount too large")
	}

	toAmountCents := (fromAmountCents * rateNum) / rateDenom
	
	residualInSourceCurrency := (fromAmountCents * rateNum) % rateDenom
	if residualInSourceCurrency > 0 {
		s.logger.Debug("exchange residual captured as spread", 
			"fromAmount", fromAmountCents, 
			"fromCurrency", fromCurrency,
			"toAmount", toAmountCents,
			"toCurrency", toCurrency,
			"residualFraction", fmt.Sprintf("%d/%d %s", residualInSourceCurrency, rateDenom, toCurrency))
	}

	tx, err := s.transactionRepo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	fromAccount, err := s.accountRepo.FindByUserAndCurrency(ctx, userID, fromCurrency)
	if err != nil {
		return nil, err
	}

	balanceCents, err := s.accountRepo.GetBalanceCentsForUpdate(ctx, tx, fromAccount.ID)
	if err != nil {
		return nil, err
	}
	if balanceCents < fromAmountCents {
		s.logger.Warn("insufficient funds for exchange", "userID", userID, "available", balanceCents, "required", fromAmountCents)
		return nil, errorsx.ErrInsufficientFunds
	}

	toAccount, err := s.accountRepo.FindByUserAndCurrency(ctx, userID, toCurrency)
	if err != nil {
		return nil, err
	}

	if _, err := s.accountRepo.GetBalanceCentsForUpdate(ctx, tx, toAccount.ID); err != nil {
		return nil, err
	}

	transaction := &models.Transaction{
		Type:        "exchange",
		FromUserID:  userID,
		Currency:    fromCurrency,
		AmountCents: fromAmountCents,
		Description: fmt.Sprintf("Exchange %d cents %s to %d cents %s (rate: %d/%d)", fromAmountCents, fromCurrency, toAmountCents, toCurrency, rateNum, rateDenom),
	}

	if err := s.transactionRepo.Create(ctx, tx, transaction); err != nil {
		return nil, err
	}

	debitEntry := &models.LedgerEntry{
		TransactionID: transaction.ID,
		AccountID:     fromAccount.ID,
		AmountCents:   -fromAmountCents,
	}
	if err := s.transactionRepo.CreateLedgerEntry(ctx, tx, debitEntry); err != nil {
		return nil, err
	}

	creditEntry := &models.LedgerEntry{
		TransactionID: transaction.ID,
		AccountID:     toAccount.ID,
		AmountCents:   toAmountCents,
	}
	if err := s.transactionRepo.CreateLedgerEntry(ctx, tx, creditEntry); err != nil {
		return nil, err
	}

	if err := s.accountRepo.UpdateBalanceCents(ctx, tx, fromAccount.ID, -fromAmountCents); err != nil {
		return nil, err
	}
	if err := s.accountRepo.UpdateBalanceCents(ctx, tx, toAccount.ID, toAmountCents); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("failed to commit exchange", "error", err)
		return nil, fmt.Errorf("error committing exchange: %w", err)
	}

	s.logger.Info("exchange completed", "transactionID", transaction.ID, "from", fromCurrency, "to", toCurrency, "amountCents", fromAmountCents)
	return transaction, nil
}

func (s *TransactionService) GetTransactions(ctx context.Context, userID, transactionType string, page, limit int) ([]models.Transaction, int, error) {
	transactions, total, err := s.transactionRepo.FindByUserID(ctx, userID, transactionType, page, limit)
	if err != nil {
		s.logger.Error("failed to get transactions", "error", err, "userID", userID)
		return nil, 0, fmt.Errorf("error getting transactions: %w", err)
	}
	return transactions, total, nil
}

