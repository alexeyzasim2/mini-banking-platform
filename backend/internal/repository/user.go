package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"mini-banking-platform/internal/errorsx"
	"mini-banking-platform/internal/models"

	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db     *sqlx.DB
	logger *slog.Logger
}

func NewUserRepository(db *sqlx.DB, logger *slog.Logger) *UserRepository {
	return &UserRepository{db: db, logger: logger}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User, hashedPassword string) error {
	query := `
		INSERT INTO users (email, password, first_name, last_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query, user.Email, hashedPassword, user.FirstName, user.LastName).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		r.logger.Error("repository: failed to create user", "error", err, "email", user.Email)
		return fmt.Errorf("repository: error creating user: %w", err)
	}

	r.logger.Info("repository: user created", "userID", user.ID, "email", user.Email)
	return nil
}

func (r *UserRepository) CreateInTx(ctx context.Context, tx *sqlx.Tx, user *models.User, hashedPassword string) error {
	query := `
		INSERT INTO users (email, password, first_name, last_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err := tx.QueryRowContext(ctx, query, user.Email, hashedPassword, user.FirstName, user.LastName).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		r.logger.Error("repository: failed to create user in tx", "error", err, "email", user.Email)
		return fmt.Errorf("repository: error creating user: %w", err)
	}

	r.logger.Info("repository: user created in tx", "userID", user.ID, "email", user.Email)
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, email, password, first_name, last_name, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errorsx.ErrUserNotFound
		}
		r.logger.Error("repository: failed to find user by email", "error", err, "email", email)
		return nil, fmt.Errorf("repository: error finding user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, email, password, first_name, last_name, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errorsx.ErrUserNotFound
		}
		r.logger.Error("repository: failed to find user by ID", "error", err, "userID", id)
		return nil, fmt.Errorf("repository: error finding user: %w", err)
	}

	return &user, nil
}

