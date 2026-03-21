package main

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
	"go.opentelemetry.io/otel/attribute"
)

type URLAnalytics struct {
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	Title       string    `json:"title"`
	FaviconURL  string    `json:"favicon_url"`
	Username    string    `json:"username"`
	ClickCount  int64     `json:"click_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type ClickRecord struct {
	ClickedAt time.Time `json:"clicked_at"`
}

type Store struct {
	db *sql.DB
}

func NewStore(dsn string) *Store {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		slog.Error("invalid postgres DSN", "error", err)
		panic(err)
	}
	db.SetMaxOpenConns(15)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		slog.Warn("postgres connection failed, will retry on first query", "error", err)
	}
	return &Store{db: db}
}

const queryTimeout = 4 * time.Second

func (s *Store) GetURLAnalytics(ctx context.Context, shortCode string) (*URLAnalytics, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	ctx, span := tracer.Start(ctx, "store.GetURLAnalytics")
	defer span.End()
	span.SetAttributes(attribute.String("short_code", shortCode))

	a := &URLAnalytics{}
	err := s.db.QueryRowContext(ctx,
		`SELECT short_code, original_url, title, favicon_url, username, click_count, created_at
		 FROM urls WHERE short_code = $1`,
		shortCode,
	).Scan(&a.ShortCode, &a.OriginalURL, &a.Title, &a.FaviconURL, &a.Username, &a.ClickCount, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Store) GetRecentClicks(ctx context.Context, shortCode string, limit int) ([]ClickRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	ctx, span := tracer.Start(ctx, "store.GetRecentClicks")
	defer span.End()
	span.SetAttributes(attribute.String("short_code", shortCode))

	rows, err := s.db.QueryContext(ctx,
		`SELECT clicked_at FROM clicks WHERE short_code = $1 ORDER BY clicked_at DESC LIMIT $2`,
		shortCode, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clicks []ClickRecord
	for rows.Next() {
		var c ClickRecord
		if err := rows.Scan(&c.ClickedAt); err != nil {
			return nil, err
		}
		clicks = append(clicks, c)
	}
	return clicks, rows.Err()
}

func (s *Store) GetUserAnalytics(ctx context.Context, username string) ([]URLAnalytics, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	ctx, span := tracer.Start(ctx, "store.GetUserAnalytics")
	defer span.End()
	span.SetAttributes(attribute.String("username", username))

	rows, err := s.db.QueryContext(ctx,
		`SELECT short_code, original_url, title, favicon_url, username, click_count, created_at
		 FROM urls WHERE username = $1 ORDER BY click_count DESC`,
		username,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analytics []URLAnalytics
	for rows.Next() {
		var a URLAnalytics
		if err := rows.Scan(&a.ShortCode, &a.OriginalURL, &a.Title, &a.FaviconURL, &a.Username, &a.ClickCount, &a.CreatedAt); err != nil {
			return nil, err
		}
		analytics = append(analytics, a)
	}
	return analytics, rows.Err()
}

func (s *Store) GetTopURLs(ctx context.Context, limit int) ([]URLAnalytics, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	ctx, span := tracer.Start(ctx, "store.GetTopURLs")
	defer span.End()

	rows, err := s.db.QueryContext(ctx,
		`SELECT short_code, original_url, title, favicon_url, username, click_count, created_at
		 FROM urls ORDER BY click_count DESC, created_at DESC LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []URLAnalytics
	for rows.Next() {
		var a URLAnalytics
		if err := rows.Scan(&a.ShortCode, &a.OriginalURL, &a.Title, &a.FaviconURL, &a.Username, &a.ClickCount, &a.CreatedAt); err != nil {
			return nil, err
		}
		urls = append(urls, a)
	}
	return urls, rows.Err()
}

func (s *Store) GetClickCount(ctx context.Context, shortCode string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	ctx, span := tracer.Start(ctx, "store.GetClickCount")
	defer span.End()

	var count int64
	err := s.db.QueryRowContext(ctx, `SELECT click_count FROM urls WHERE short_code = $1`, shortCode).Scan(&count)
	return count, err
}
