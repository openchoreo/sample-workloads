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

type URL struct {
	ID          int       `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	Title       string    `json:"title"`
	FaviconURL  string    `json:"favicon_url"`
	Username    string    `json:"username"`
	ClickCount  int64     `json:"click_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Store struct {
	db *sql.DB
}

func NewStore(dsn string) *Store {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		// sql.Open only validates DSN format; this is a programming error
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

const queryTimeout = 4 * time.Second

func (s *Store) InsertURL(ctx context.Context, shortCode, originalURL, username string) (*URL, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	ctx, span := tracer.Start(ctx, "store.InsertURL")
	defer span.End()
	span.SetAttributes(attribute.String("short_code", shortCode))

	u := &URL{}
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO urls (short_code, original_url, username)
		 VALUES ($1, $2, $3)
		 RETURNING id, short_code, original_url, title, favicon_url, username, click_count, created_at, updated_at`,
		shortCode, originalURL, username,
	).Scan(&u.ID, &u.ShortCode, &u.OriginalURL, &u.Title, &u.FaviconURL, &u.Username, &u.ClickCount, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert url: %w", err)
	}
	return u, nil
}

func (s *Store) GetURL(ctx context.Context, shortCode string) (*URL, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	ctx, span := tracer.Start(ctx, "store.GetURL")
	defer span.End()
	span.SetAttributes(attribute.String("short_code", shortCode))

	u := &URL{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, short_code, original_url, title, favicon_url, username, click_count, created_at, updated_at
		 FROM urls WHERE short_code = $1`,
		shortCode,
	).Scan(&u.ID, &u.ShortCode, &u.OriginalURL, &u.Title, &u.FaviconURL, &u.Username, &u.ClickCount, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Store) ListURLs(ctx context.Context, username string) ([]URL, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	ctx, span := tracer.Start(ctx, "store.ListURLs")
	defer span.End()
	span.SetAttributes(attribute.String("username", username))

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, short_code, original_url, title, favicon_url, username, click_count, created_at, updated_at
		 FROM urls WHERE username = $1 ORDER BY created_at DESC`,
		username,
	)
	if err != nil {
		return nil, fmt.Errorf("list urls: %w", err)
	}
	defer rows.Close()

	var urls []URL
	for rows.Next() {
		var u URL
		if err := rows.Scan(&u.ID, &u.ShortCode, &u.OriginalURL, &u.Title, &u.FaviconURL, &u.Username, &u.ClickCount, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan url: %w", err)
		}
		urls = append(urls, u)
	}
	return urls, rows.Err()
}

func (s *Store) DeleteURL(ctx context.Context, shortCode string) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	ctx, span := tracer.Start(ctx, "store.DeleteURL")
	defer span.End()
	span.SetAttributes(attribute.String("short_code", shortCode))

	result, err := s.db.ExecContext(ctx, `DELETE FROM urls WHERE short_code = $1`, shortCode)
	if err != nil {
		return fmt.Errorf("delete url: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) RecordClick(ctx context.Context, shortCode string) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	ctx, span := tracer.Start(ctx, "store.RecordClick")
	defer span.End()
	span.SetAttributes(attribute.String("short_code", shortCode))

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `INSERT INTO clicks (short_code) VALUES ($1)`, shortCode)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE urls SET click_count = click_count + 1, updated_at = NOW() WHERE short_code = $1`, shortCode)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) UpdateMetadata(ctx context.Context, shortCode, title, faviconURL string) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	ctx, span := tracer.Start(ctx, "store.UpdateMetadata")
	defer span.End()
	span.SetAttributes(attribute.String("short_code", shortCode))

	_, err := s.db.ExecContext(ctx,
		`UPDATE urls SET title = $1, favicon_url = $2, updated_at = NOW() WHERE short_code = $3`,
		title, faviconURL, shortCode,
	)
	return err
}

func (s *Store) ShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	ctx, span := tracer.Start(ctx, "store.ShortCodeExists")
	defer span.End()
	span.SetAttributes(attribute.String("short_code", shortCode))
	var exists bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM urls WHERE short_code = $1)`, shortCode).Scan(&exists)
	return exists, err
}
