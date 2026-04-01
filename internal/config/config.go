package config

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"net"
	"strings"
)

const (
	StorageMemory   = "memory"
	StoragePostgres = "postgres"
	DefaultHTTPAddr = ":8080"
)

// Config contains runtime options for the application bootstrap
type Config struct {
	Storage     string
	PostgresDSN string
	HTTPAddr    string
	BaseURL     string
}

func Parse(args []string) (Config, error) {
	cfg := Config{}

	fs := flag.NewFlagSet("short-link", flag.ContinueOnError)
	var parseOutput bytes.Buffer
	// parseOutput captures flag parsing errors and usage text
	fs.SetOutput(&parseOutput)

	fs.StringVar(&cfg.Storage, "storage", StorageMemory, "storage backend: memory|postgres")
	fs.StringVar(&cfg.PostgresDSN, "postgres-dsn", "", "postgres connection string")
	fs.StringVar(&cfg.HTTPAddr, "http-addr", DefaultHTTPAddr, "http listen address")

	if err := fs.Parse(args); err != nil {
		return Config{}, formatParseError(err, parseOutput.String())
	}

	if fs.NArg() != 0 {
		return Config{}, fmt.Errorf("unexpected positional arguments: %s", strings.Join(fs.Args(), " "))
	}

	cfg.Storage = strings.TrimSpace(cfg.Storage)
	cfg.PostgresDSN = strings.TrimSpace(cfg.PostgresDSN)
	cfg.HTTPAddr = strings.TrimSpace(cfg.HTTPAddr)

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	baseURL, err := baseURLFromHTTPAddr(cfg.HTTPAddr)
	if err != nil {
		return Config{}, err
	}
	cfg.BaseURL = baseURL

	return cfg, nil
}

func (c Config) Validate() error {
	switch c.Storage {
	case StorageMemory, StoragePostgres:
	default:
		return fmt.Errorf("invalid storage %q: must be %q or %q", c.Storage, StorageMemory, StoragePostgres)
	}

	if c.HTTPAddr == "" {
		return errors.New("http-addr is required")
	}

	if _, _, err := net.SplitHostPort(c.HTTPAddr); err != nil {
		return fmt.Errorf("invalid http-addr %q: %w", c.HTTPAddr, err)
	}

	if c.Storage == StoragePostgres && c.PostgresDSN == "" {
		return errors.New("postgres-dsn is required when storage=postgres")
	}

	return nil
}

func baseURLFromHTTPAddr(httpAddr string) (string, error) {
	host, port, err := net.SplitHostPort(httpAddr)
	if err != nil {
		return "", fmt.Errorf("invalid http-addr %q: %w", httpAddr, err)
	}

	if port == "" {
		return "", fmt.Errorf("invalid http-addr %q: missing port", httpAddr)
	}

	return "http://" + net.JoinHostPort(normalizeBaseURLHost(host), port), nil
}

func normalizeBaseURLHost(host string) string {
	switch host {
	case "", "0.0.0.0", "::":
		return "localhost"
	default:
		return host
	}
}

func formatParseError(err error, usage string) error {
	usage = strings.TrimSpace(usage)
	if usage == "" {
		return err
	}

	return fmt.Errorf("%w\n%s", err, usage)
}
