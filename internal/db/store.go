package db

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store wraps all database access.
type Store struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// New establishes a connection pool.
func New(ctx context.Context, url string, logger *slog.Logger) (*Store, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Store{pool: pool, logger: logger}, nil
}

// Close releases the underlying connection pool.
func (s *Store) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// Ping validates the connection.
func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}
