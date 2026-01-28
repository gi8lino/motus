package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthChecker is the minimal interface needed for health checks.
type HealthChecker interface {
	Ping(ctx context.Context) error
}

// Store wraps all database access.
type Store struct {
	pool *pgxpool.Pool
}

// New establishes a connection pool.
func New(ctx context.Context, url string) (*Store, error) {
	// Initialize the pgx connection pool with the provided URL.
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	return &Store{pool: pool}, nil
}

// Close releases the underlying connection pool.
func (s *Store) Close() {
	// Guard against nil pool during shutdown.
	if s.pool != nil {
		s.pool.Close()
	}
}

// Ping validates the connection.
func (s *Store) Ping(ctx context.Context) error {
	// Delegate to the underlying pool health check.
	return s.pool.Ping(ctx)
}
