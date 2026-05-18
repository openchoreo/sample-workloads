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

type Note struct {
	ID        string    `json:"id"`
	Author    string    `json:"author"`
	Body      string    `json:"body"`
	ViewCount int64     `json:"view_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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

func (s *Store) InsertNote(ctx context.Context, id, author, body string) (*Note, error) {
	ctx, span := tracer.Start(ctx, "store.InsertNote")
	defer span.End()
	span.SetAttributes(attribute.String("note.id", id))
	applyDelay(ctx, StageDB)
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	n := &Note{}
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO notes (id, author, body)
		 VALUES ($1, $2, $3)
		 RETURNING id, author, body, view_count, created_at, updated_at`,
		id, author, body,
	).Scan(&n.ID, &n.Author, &n.Body, &n.ViewCount, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert note: %w", err)
	}
	return n, nil
}

func (s *Store) GetNote(ctx context.Context, id string) (*Note, error) {
	ctx, span := tracer.Start(ctx, "store.GetNote")
	defer span.End()
	span.SetAttributes(attribute.String("note.id", id))
	applyDelay(ctx, StageDB)
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	n := &Note{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, author, body, view_count, created_at, updated_at
		 FROM notes WHERE id = $1`,
		id,
	).Scan(&n.ID, &n.Author, &n.Body, &n.ViewCount, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (s *Store) ListNotes(ctx context.Context, author string) ([]Note, error) {
	ctx, span := tracer.Start(ctx, "store.ListNotes")
	defer span.End()
	span.SetAttributes(attribute.String("author", author))
	applyDelay(ctx, StageDB)
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	var (
		rows *sql.Rows
		err  error
	)
	if author == "" {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, author, body, view_count, created_at, updated_at
			 FROM notes ORDER BY created_at DESC LIMIT 100`)
	} else {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, author, body, view_count, created_at, updated_at
			 FROM notes WHERE author = $1 ORDER BY created_at DESC LIMIT 100`,
			author)
	}
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Author, &n.Body, &n.ViewCount, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (s *Store) DeleteNote(ctx context.Context, id string) error {
	ctx, span := tracer.Start(ctx, "store.DeleteNote")
	defer span.End()
	span.SetAttributes(attribute.String("note.id", id))
	applyDelay(ctx, StageDB)
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	res, err := s.db.ExecContext(ctx, `DELETE FROM notes WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) RecordView(ctx context.Context, id string) error {
	ctx, span := tracer.Start(ctx, "store.RecordView")
	defer span.End()
	span.SetAttributes(attribute.String("note.id", id))
	applyDelay(ctx, StageDB)
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `INSERT INTO views (note_id) VALUES ($1)`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE notes SET view_count = view_count + 1, updated_at = NOW() WHERE id = $1`, id); err != nil {
		return err
	}
	return tx.Commit()
}
