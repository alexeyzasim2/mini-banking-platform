package handlers

import (
	"log/slog"

	"mini-banking-platform/internal/config"
	"mini-banking-platform/internal/jwt"
	"mini-banking-platform/internal/service"
)

type Handler struct {
	authService        *service.AuthService
	accountService     *service.AccountService
	transactionService *service.TransactionService
	config             *config.Config
	jwtService         *jwt.Service
	logger             *slog.Logger
}

func NewHandler(
	authService *service.AuthService,
	accountService *service.AccountService,
	transactionService *service.TransactionService,
	config *config.Config,
	jwtService *jwt.Service,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		authService:        authService,
		accountService:     accountService,
		transactionService: transactionService,
		config:             config,
		jwtService:         jwtService,
		logger:             logger,
	}
}

