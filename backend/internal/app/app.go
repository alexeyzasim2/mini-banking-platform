package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"mini-banking-platform/internal/config"
	"mini-banking-platform/internal/http/dto"
	"mini-banking-platform/internal/http/handlers"
	"mini-banking-platform/internal/http/routes"
	"mini-banking-platform/internal/jwt"
	"mini-banking-platform/internal/repository"
	"mini-banking-platform/internal/service"
	"mini-banking-platform/pkg/logger"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

type App struct {
	cfg        *config.Config
	httpServer *http.Server
	db         *sqlx.DB
	logger     *slog.Logger
}

func NewApp() (*App, error) {
	log := logger.New()

	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	db, err := connectDatabase(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := runMigrations(db.DB, log); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	jwtService := jwt.NewService(cfg.JWTSecret, cfg.JWTExpiryHours)

	repos := repository.NewRepositories(db, log)
	authService := service.NewAuthService(repos.User, repos.Account, repos.Transaction, jwtService, log)
	accountService := service.NewAccountService(repos.Account, repos.Transaction, log)
	transactionService := service.NewTransactionService(repos.Account, repos.Transaction, repos.User, log)


	if err := seedUsers(context.Background(), authService, cfg, log); err != nil {
		log.Warn("failed to seed users (may already exist)", "error", err)
	}

	handler := handlers.NewHandler(authService, accountService, transactionService, cfg, jwtService, log)
	router := routes.NewRouter(handler, jwtService, log)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &App{
		cfg:        cfg,
		httpServer: httpServer,
		db:         db,
		logger:     log,
	}, nil
}

func (a *App) Run() error {
	a.logger.Info("server starting", "port", a.cfg.Port)
	err := a.httpServer.ListenAndServe()
	if err != nil && errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("shutting down server")
	if err := a.httpServer.Shutdown(ctx); err != nil {
		return err
	}
	return a.db.Close()
}

func (a *App) ShutdownTimeout() time.Duration {
	return 10 * time.Second
}

func connectDatabase(cfg *config.Config, log *slog.Logger) (*sqlx.DB, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Error("failed to connect to PostgreSQL", "error", err)
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Info("successful connection to PostgreSQL",
		"host", cfg.DBHost,
		"port", cfg.DBPort,
		"database", cfg.DBName)

	return db, nil
}

func runMigrations(db *sql.DB, log *slog.Logger) error {
	log.Info("running database migrations")

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Info("database migrations completed successfully")
	return nil
}

func seedUsers(ctx context.Context, authService *service.AuthService, cfg *config.Config, log *slog.Logger) error {
	testUsers := []struct {
		email     string
		password  string
		firstName string
		lastName  string
	}{
		{"alice@example.com", "password123", "Alice", "Smith"},
		{"bob@example.com", "password123", "Bob", "Johnson"},
		{"charlie@example.com", "password123", "Charlie", "Brown"},
	}

	for _, u := range testUsers {
		_, err := authService.Register(ctx, dto.RegisterRequest{
			Email:     u.email,
			Password:  u.password,
			FirstName: u.firstName,
			LastName:  u.lastName,
		}, cfg.InitialBalanceUSDCents, cfg.InitialBalanceEURCents)

		if err != nil {
			log.Debug("user seed skipped", "email", u.email, "reason", err.Error())
		} else {
			log.Info("test user seeded", "email", u.email)
		}
	}

	return nil
}

