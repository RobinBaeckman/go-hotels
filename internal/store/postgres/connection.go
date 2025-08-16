package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgx-contrib/pgxotel"
)

func Connect(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// 🔹 Koppla in OpenTelemetry tracing för pgx v5
	// Name blir span-namespace för queries i denna pool.
	cfg.ConnConfig.Tracer = &pgxotel.QueryTracer{
		Name: "postgres", // valfritt: sätt t.ex. "go-hotels-db"
		// Instrument: pgxotel.Instrumentation… (om du vill sätta explicit)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create db pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}
	return pool, nil
}
