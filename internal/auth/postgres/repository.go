package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"go-notes-api/internal/auth"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(
	ctx context.Context,
	email string,
	passwordHash string,
) (auth.User, error) {
	const query = `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, password_hash, created_at
	`

	var user auth.User

	err := r.db.QueryRowContext(
		ctx,
		query,
		email,
		passwordHash,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return auth.User{}, auth.ErrEmailTaken
		}

		return auth.User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (r *Repository) GetByEmail(
	ctx context.Context,
	email string,
) (auth.User, error) {
	const query = `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE LOWER(email) = $1
	`

	var user auth.User

	err := r.db.QueryRowContext(
		ctx,
		query,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return auth.User{}, auth.ErrUserNotFound
	}

	if err != nil {
		return auth.User{}, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}
