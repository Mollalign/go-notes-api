package database

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"go-notes-api/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(
	ctx context.Context,
	cfg config.DatabaseConfig,
) (*sql.DB, error) {
	// Build the connection URL safely, including the password
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host: net.JoinHostPort(
			cfg.Host,
			strconv.Itoa(cfg.Port),
		),
		Path: cfg.Name,
	}

	query := u.Query()
	query.Set("sslmode", cfg.SSLMode)
	u.RawQuery = query.Encode()

	// Create the database handle
	db, err := sql.Open("pgx", u.String())
	if err != nil {
		return nil, fmt.Errorf(
			"create database handle: %w",
			err,
		)
	}

	// Set basics connection-pool limits.
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxIdleTime(30 * time.Minute)

	// Verify that Postgres is reachable
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()

		return nil, fmt.Errorf(
			"verify PostgreSQL connection: %w",
			err,
		)
	}

	return db, nil
}
