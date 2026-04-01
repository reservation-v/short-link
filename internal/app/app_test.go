package app

import (
	"context"
	"strings"
	"testing"

	"github.com/reservation-v/short-link/internal/config"
)

func TestNewStorage_Memory(t *testing.T) {
	st, cleanup, err := newStorage(config.Config{Storage: config.StorageMemory})
	if err != nil {
		t.Fatalf("newStorage returned error: %v", err)
	}
	if st == nil {
		t.Fatal("newStorage returned nil storage")
	}

	cleanup()
}

func TestNewStorage_UnsupportedStorage(t *testing.T) {
	_, _, err := newStorage(config.Config{Storage: "redis"})
	if err == nil {
		t.Fatal("newStorage returned nil error, want unsupported storage error")
	}

	if !strings.Contains(err.Error(), "unsupported storage backend") {
		t.Fatalf("newStorage error = %q, want unsupported storage backend error", err)
	}
}

func TestRun_PostgresNotImplemented(t *testing.T) {
	err := Run(context.Background(), config.Config{Storage: config.StoragePostgres})
	if err == nil {
		t.Fatal("Run returned nil error, want postgres not implemented error")
	}

	if !strings.Contains(err.Error(), "not implemented") {
		t.Fatalf("Run error = %q, want postgres not implemented error", err)
	}
}
