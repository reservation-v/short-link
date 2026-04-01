package config

import (
	"strings"
	"testing"
)

func TestParse_Defaults(t *testing.T) {
	cfg, err := Parse(nil)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cfg.Storage != StorageMemory {
		t.Fatalf("Storage = %q, want %q", cfg.Storage, StorageMemory)
	}
	if cfg.HTTPAddr != DefaultHTTPAddr {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, DefaultHTTPAddr)
	}
	if cfg.BaseURL != "http://localhost:8080" {
		t.Fatalf("BaseURL = %q, want %q", cfg.BaseURL, "http://localhost:8080")
	}
}

func TestParse_CustomHTTPAddr(t *testing.T) {
	cfg, err := Parse([]string{"--http-addr=127.0.0.1:9090"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cfg.BaseURL != "http://127.0.0.1:9090" {
		t.Fatalf("BaseURL = %q, want %q", cfg.BaseURL, "http://127.0.0.1:9090")
	}
}

func TestParse_WildcardAddrUsesLocalhostInBaseURL(t *testing.T) {
	cfg, err := Parse([]string{"--http-addr=0.0.0.0:8081"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cfg.BaseURL != "http://localhost:8081" {
		t.Fatalf("BaseURL = %q, want %q", cfg.BaseURL, "http://localhost:8081")
	}
}

func TestParse_PostgresRequiresDSN(t *testing.T) {
	_, err := Parse([]string{"--storage=postgres"})
	if err == nil {
		t.Fatal("Parse returned nil error, want postgres-dsn validation error")
	}

	if !strings.Contains(err.Error(), "postgres-dsn is required") {
		t.Fatalf("Parse error = %q, want postgres-dsn validation error", err)
	}
}

func TestParse_InvalidStorage(t *testing.T) {
	_, err := Parse([]string{"--storage=redis"})
	if err == nil {
		t.Fatal("Parse returned nil error, want invalid storage error")
	}

	if !strings.Contains(err.Error(), "invalid storage") {
		t.Fatalf("Parse error = %q, want invalid storage error", err)
	}
}

func TestParse_InvalidHTTPAddr(t *testing.T) {
	_, err := Parse([]string{"--http-addr=localhost"})
	if err == nil {
		t.Fatal("Parse returned nil error, want invalid http-addr error")
	}

	if !strings.Contains(err.Error(), "invalid http-addr") {
		t.Fatalf("Parse error = %q, want invalid http-addr error", err)
	}
}

func TestParse_UnexpectedPositionalArgs(t *testing.T) {
	_, err := Parse([]string{"extra"})
	if err == nil {
		t.Fatal("Parse returned nil error, want unexpected positional arguments error")
	}

	if !strings.Contains(err.Error(), "unexpected positional arguments") {
		t.Fatalf("Parse error = %q, want unexpected positional arguments error", err)
	}
}
