//go:build integration

package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	containerImage          = "postgres:18-alpine"
	containerDB             = "shortener"
	containerUser           = "postgres"
	containerPassword       = "postgres"
	containerStartupTimeout = 90 * time.Second
	queryTimeout            = 5 * time.Second
)

var (
	pool      *pgxpool.Pool
	storage   *Storage
	container testcontainers.Container
)

func resetLinksTable(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	if _, err := pool.Exec(ctx, "TRUNCATE TABLE links RESTART IDENTITY"); err != nil {
		t.Fatalf("truncate links table: %v", err)
	}
}

func startPostgresContainer(ctx context.Context) (testcontainers.Container, string, error) {
	startedContainer, err := tcpostgres.Run(ctx,
		containerImage,
		tcpostgres.WithDatabase(containerDB),
		tcpostgres.WithUsername(containerUser),
		tcpostgres.WithPassword(containerPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(containerStartupTimeout),
		),
	)
	if err != nil {
		return nil, "", err
	}

	dsn, err := startedContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = startedContainer.Terminate(ctx)
		return nil, "", fmt.Errorf("build connection string: %w", err)
	}

	return startedContainer, dsn, nil
}

func cleanupContainer(ctx context.Context) {
	if container == nil {
		return
	}
	_ = container.Terminate(ctx)
}

func applyMigrations(dsn string) error {
	migrationsDir, err := findMigrationsDir()
	if err != nil {
		return err
	}

	migrationSourceURL := "file://" + filepath.ToSlash(migrationsDir)
	migrationDatabaseURL, err := toPGXURL(dsn)
	if err != nil {
		return err
	}

	m, err := migrate.New(migrationSourceURL, migrationDatabaseURL)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}
	defer closeMigrator(m)

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run up migrations: %w", err)
	}

	return nil
}

func toPGXURL(dsn string) (string, error) {
	switch {
	case strings.HasPrefix(dsn, "postgres://"):
		return "pgx5://" + strings.TrimPrefix(dsn, "postgres://"), nil
	case strings.HasPrefix(dsn, "postgresql://"):
		return "pgx5://" + strings.TrimPrefix(dsn, "postgresql://"), nil
	case strings.HasPrefix(dsn, "pgx5://"):
		return dsn, nil
	default:
		return "", fmt.Errorf("unsupported dsn scheme for migrate: %q", dsn)
	}
}

func closeMigrator(m *migrate.Migrate) {
	srcErr, dbErr := m.Close()
	if srcErr != nil {
		fmt.Fprintf(os.Stderr, "close migrate source: %v\n", srcErr)
	}
	if dbErr != nil {
		fmt.Fprintf(os.Stderr, "close migrate database: %v\n", dbErr)
	}
}

func findMigrationsDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	dir := wd
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return filepath.Join(dir, "migrations"), nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find project root from %s", wd)
		}
		dir = parent
	}
}
