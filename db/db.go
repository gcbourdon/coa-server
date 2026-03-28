package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool is the shared database connection pool.
var Pool *pgxpool.Pool

// Connect initialises the PostgreSQL connection pool.
// dsn example: "postgres://user:password@localhost:5432/coa"
func Connect(ctx context.Context, dsn string) error {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connecting to postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("pinging postgres: %w", err)
	}
	Pool = pool
	return nil
}

// Close shuts down the connection pool.
func Close() {
	if Pool != nil {
		Pool.Close()
	}
}
