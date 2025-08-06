package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create db pool: %w", err)
	}
	// Optional: immediate ping to verify connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}
	return pool, nil
}
