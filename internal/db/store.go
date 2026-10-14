package db

import (
	"context"
	"errors"

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

// New establishes a connection pool using the driver defaults.
func New(ctx context.Context, url string) (*Store, error) {
	return NewWithPoolLimits(ctx, url, 0, 0)
}

// NewWithPoolLimits establishes a connection pool with optional explicit limits.
func NewWithPoolLimits(ctx context.Context, url string, maxConns, minIdleConns int32) (*Store, error) {
	if maxConns < 0 || minIdleConns < 0 || (maxConns > 0 && minIdleConns > maxConns) {
		return nil, errors.New("invalid database connection pool limits")
	}
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	if maxConns != 0 {
		cfg.MaxConns = maxConns
	}
	if minIdleConns != 0 {
		cfg.MinIdleConns = minIdleConns
	}
	if cfg.MinConns > cfg.MaxConns || cfg.MinIdleConns > cfg.MaxConns {
		return nil, errors.New("database minimum connection count exceeds maximum")
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Store{pool: pool}, nil
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
