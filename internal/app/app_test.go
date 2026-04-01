package app

import (
	"context"
	"strings"
	"testing"

	"github.com/reservation-v/short-link/internal/config"
)

func TestNewStorage_Memory(t *testing.T) {
	st, cleanup, err := newStorage(context.Background(), config.Config{Storage: config.StorageMemory})
	if err != nil {
		t.Fatalf("newStorage returned error: %v", err)
	}
	if st == nil {
		t.Fatal("newStorage returned nil storage")
	}

	cleanup()
}

func TestNewStorage_UnsupportedStorage(t *testing.T) {
	_, _, err := newStorage(context.Background(), config.Config{Storage: "redis"})
	if err == nil {
		t.Fatal("newStorage returned nil error, want unsupported storage error")
	}

	if !strings.Contains(err.Error(), "unsupported storage backend") {
		t.Fatalf("newStorage error = %q, want unsupported storage backend error", err)
	}
}

func TestNewStorage_PostgresInvalidDSN(t *testing.T) {
	_, cleanup, err := newStorage(context.Background(), config.Config{
		Storage:     config.StoragePostgres,
		PostgresDSN: "postgres://%",
	})
	if err == nil {
		t.Fatal("newStorage returned nil error, want invalid dsn error")
	}

	if cleanup != nil {
		t.Fatal("newStorage returned cleanup on error")
	}

	if !strings.Contains(err.Error(), "create postgres pool") {
		t.Fatalf("newStorage error = %q, want create postgres pool error", err)
	}
}
