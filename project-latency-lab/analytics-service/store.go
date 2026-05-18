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

type TopNote struct {
	ID        string `json:"id"`
	Author    string `json:"author"`
	ViewCount int64  `json:"view_count"`
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

func (s *Store) GetViewCount(ctx context.Context, id string) (int64, error) {
	ctx, span := tracer.Start(ctx, "store.GetViewCount")
	defer span.End()
	span.SetAttributes(attribute.String("note.id", id))
	applyDelay(ctx, StageDB)
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	var count int64
	err := s.db.QueryRowContext(ctx,
		`SELECT view_count FROM notes WHERE id = $1`, id,
	).Scan(&count)
	return count, err
}

func (s *Store) RecentViews(ctx context.Context, id string, limit int) ([]string, error) {
	ctx, span := tracer.Start(ctx, "store.RecentViews")
	defer span.End()
	span.SetAttributes(attribute.String("note.id", id))
	applyDelay(ctx, StageDB)
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	rows, err := s.db.QueryContext(ctx,
		`SELECT viewed_at FROM views WHERE note_id = $1 ORDER BY viewed_at DESC LIMIT $2`,
		id, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("recent views: %w", err)
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var t time.Time
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		out = append(out, t.UTC().Format(time.RFC3339))
	}
	return out, rows.Err()
}

func (s *Store) TopNotes(ctx context.Context, limit int) ([]TopNote, error) {
	ctx, span := tracer.Start(ctx, "store.TopNotes")
	defer span.End()
	applyDelay(ctx, StageDB)
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, author, view_count FROM notes ORDER BY view_count DESC LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("top notes: %w", err)
	}
	defer rows.Close()

	var out []TopNote
	for rows.Next() {
		var n TopNote
		if err := rows.Scan(&n.ID, &n.Author, &n.ViewCount); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}
