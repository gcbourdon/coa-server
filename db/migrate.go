package db

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

// Migrate runs all pending *.up.sql files found in migrationsFS in lexicographic order.
// It creates a schema_migrations tracking table on first run.
// Pass the embed.FS (or any fs.FS) from the caller so this package avoids the
// //go:embed "../.." restriction.
func Migrate(ctx context.Context, migrationsFS fs.FS) error {
	if Pool == nil {
		return fmt.Errorf("db.Migrate: pool is nil — call Connect first")
	}

	if _, err := Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT        PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`); err != nil {
		return fmt.Errorf("creating schema_migrations: %w", err)
	}

	entries, err := fs.Glob(migrationsFS, "*.up.sql")
	if err != nil {
		return fmt.Errorf("listing migration files: %w", err)
	}
	sort.Strings(entries)

	for _, path := range entries {
		version := migrationVersion(path)

		var exists bool
		if err := Pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`,
			version,
		).Scan(&exists); err != nil {
			return fmt.Errorf("checking migration %s: %w", version, err)
		}
		if exists {
			continue
		}

		sql, err := fs.ReadFile(migrationsFS, path)
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", path, err)
		}

		if _, err := Pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("applying migration %s: %w", version, err)
		}

		if _, err := Pool.Exec(ctx,
			`INSERT INTO schema_migrations (version) VALUES ($1)`,
			version,
		); err != nil {
			return fmt.Errorf("recording migration %s: %w", version, err)
		}

		fmt.Printf("applied migration: %s\n", version)
	}

	return nil
}

// migrationVersion extracts the base filename (without extension) as the version key.
// e.g. "001_create_users.up.sql" → "001_create_users"
func migrationVersion(path string) string {
	base := path[strings.LastIndex(path, "/")+1:]
	return strings.TrimSuffix(base, ".up.sql")
}
