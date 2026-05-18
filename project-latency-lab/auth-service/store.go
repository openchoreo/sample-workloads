package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"
)

type Store struct {
	db *sql.DB
}

func NewStore(dsn string) *Store {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		slog.Error("invalid postgres DSN", "error", err)
		panic(err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		slog.Warn("postgres connection failed, will retry on first query", "error", err)
	}
	return &Store{db: db}
}

const queryTimeout = 35 * time.Second

// UpsertUser registers (or rotates the token of) a user.
func (s *Store) UpsertUser(ctx context.Context, username, token string) error {
	ctx, span := tracer.Start(ctx, "store.UpsertUser")
	defer span.End()
	span.SetAttributes(attribute.String("username", username))
	applyDelay(ctx, StageDB)
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO users (username, token) VALUES ($1, $2)
		 ON CONFLICT (username) DO UPDATE SET token = EXCLUDED.token, updated_at = NOW()`,
		username, token,
	)
	if err != nil {
		return fmt.Errorf("upsert user: %w", err)
	}
	return nil
}

func (s *Store) GetUsernameByToken(ctx context.Context, token string) (string, error) {
	ctx, span := tracer.Start(ctx, "store.GetUsernameByToken")
	defer span.End()
	applyDelay(ctx, StageDB)
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	var username string
	err := s.db.QueryRowContext(ctx,
		`SELECT username FROM users WHERE token = $1`, token,
	).Scan(&username)
	if err != nil {
		return "", err
	}
	span.SetAttributes(attribute.String("username", username))
	return username, nil
}
