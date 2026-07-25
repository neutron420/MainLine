package database

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`); err != nil {
		return fmt.Errorf("creating migration tracking table: %w", err)
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("reading migrations directory: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, f := range files {
		var applied bool
		if err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename = $1)", f).Scan(&applied); err != nil {
			return fmt.Errorf("checking migration %s: %w", f, err)
		}
		if applied {
			continue
		}

		if err := runFile(ctx, pool, filepath.Join(migrationsDir, f)); err != nil {
			return fmt.Errorf("running migration %s: %w", f, err)
		}
		if _, err := pool.Exec(ctx, "INSERT INTO schema_migrations (filename) VALUES ($1)", f); err != nil {
			return fmt.Errorf("recording migration %s: %w", f, err)
		}
		fmt.Printf("migration applied: %s\n", f)
	}
	return nil
}

func runFile(ctx context.Context, pool *pgxpool.Pool, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	sql, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, string(sql))
	return err
}
