package service

import (
	"context"
	"fmt"
	"log/slog"
	"mini-banking-platform/internal/errorsx"
	"mini-banking-platform/internal/http/dto"
	"mini-banking-platform/internal/jwt"
	"mini-banking-platform/internal/models"
	"mini-banking-platform/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo        *repository.UserRepository
	accountRepo     *repository.AccountRepository
	transactionRepo *repository.TransactionRepository
	jwtService      *jwt.Service
	logger          *slog.Logger
}

func NewAuthService(userRepo *repository.UserRepository, accountRepo *repository.AccountRepository, transactionRepo *repository.TransactionRepository, jwtService *jwt.Service, logger *slog.Logger) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		jwtService:      jwtService,
		logger:          logger,
	}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest, initialBalanceUSDCents, initialBalanceEURCents int64) (*dto.AuthResponse, error) {
	if req.Email == "" {
		return nil, errorsx.BadRequest("email is required")
	}
	if len(req.Password) < 6 {
		return nil, errorsx.BadRequest("password must be at least 6 characters")
	}
	if req.FirstName == "" {
		return nil, errorsx.BadRequest("first name is required")
	}
	if req.LastName == "" {
		return nil, errorsx.BadRequest("last name is required")
	}

	existingUser, _ := s.userRepo.FindByEmail(ctx, req.Email)
	if existingUser != nil {
		s.logger.Warn("registration failed - user exists", "email", req.Email)
		return nil, errorsx.ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", "error", err)
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	user := &models.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	tx, err := s.accountRepo.BeginTx(ctx)
	if err != nil {
		s.logger.Error("failed to begin transaction", "error", err)
		return nil, fmt.Errorf("error beginning transaction: %w", err)
	}
	defer tx.Rollback()

	if err := s.userRepo.CreateInTx(ctx, tx, user, string(hashedPassword)); err != nil {
		s.logger.Error("failed to create user", "error", err)
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	usdAccount := &models.Account{
		UserID:       user.ID,
		Currency:     "USD",
		BalanceCents: initialBalanceUSDCents,
	}
	if err := s.accountRepo.CreateInTx(ctx, tx, usdAccount); err != nil {
		s.logger.Error("failed to create USD account", "error", err)
		return nil, fmt.Errorf("error creating USD account: %w", err)
	}

	eurAccount := &models.Account{
		UserID:       user.ID,
		Currency:     "EUR",
		BalanceCents: initialBalanceEURCents,
	}
	if err := s.accountRepo.CreateInTx(ctx, tx, eurAccount); err != nil {
		s.logger.Error("failed to create EUR account", "error", err)
		return nil, fmt.Errorf("error creating EUR account: %w", err)
	}

	if initialBalanceUSDCents > 0 {
		usdTransaction := &models.Transaction{
			Type:        "initial_deposit",
			FromUserID:  user.ID,
			Currency:    "USD",
			AmountCents: initialBalanceUSDCents,
			Description: "Initial deposit",
		}
		if err := s.transactionRepo.CreateInTx(ctx, tx, usdTransaction); err != nil {
			s.logger.Error("failed to create USD initial transaction", "error", err)
			return nil, fmt.Errorf("error creating USD initial transaction: %w", err)
		}

		usdLedgerEntry := &models.LedgerEntry{
			TransactionID: usdTransaction.ID,
			AccountID:     usdAccount.ID,
			AmountCents:   initialBalanceUSDCents,
		}
		if err := s.transactionRepo.CreateLedgerEntryInTx(ctx, tx, usdLedgerEntry); err != nil {
			s.logger.Error("failed to create USD ledger entry", "error", err)
			return nil, fmt.Errorf("error creating USD ledger entry: %w", err)
		}
	}

	if initialBalanceEURCents > 0 {
		eurTransaction := &models.Transaction{
			Type:        "initial_deposit",
			FromUserID:  user.ID,
			Currency:    "EUR",
			AmountCents: initialBalanceEURCents,
			Description: "Initial deposit",
		}
		if err := s.transactionRepo.CreateInTx(ctx, tx, eurTransaction); err != nil {
			s.logger.Error("failed to create EUR initial transaction", "error", err)
			return nil, fmt.Errorf("error creating EUR initial transaction: %w", err)
		}

		eurLedgerEntry := &models.LedgerEntry{
			TransactionID: eurTransaction.ID,
			AccountID:     eurAccount.ID,
			AmountCents:   initialBalanceEURCents,
		}
		if err := s.transactionRepo.CreateLedgerEntryInTx(ctx, tx, eurLedgerEntry); err != nil {
			s.logger.Error("failed to create EUR ledger entry", "error", err)
			return nil, fmt.Errorf("error creating EUR ledger entry: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		s.logger.Error("failed to commit registration", "error", err)
		return nil, fmt.Errorf("error committing registration: %w", err)
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		s.logger.Error("failed to generate token", "error", err)
		return nil, fmt.Errorf("error generating token: %w", err)
	}

	s.logger.Info("user registered successfully", "userID", user.ID, "email", user.Email)

	return &dto.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Warn("login failed", "email", req.Email)
		return nil, errorsx.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.logger.Warn("login failed - invalid password", "email", req.Email)
		return nil, errorsx.ErrInvalidCredentials
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		s.logger.Error("failed to generate token", "error", err)
		return nil, fmt.Errorf("error generating token: %w", err)
	}

	s.logger.Info("user logged in successfully", "userID", user.ID, "email", user.Email)

	return &dto.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *AuthService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

func (s *AuthService) generateToken(userID string) (string, error) {
	return s.jwtService.GenerateToken(userID)
}

