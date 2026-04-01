package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	stdhttp "net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reservation-v/short-link/internal/config"
	apphttp "github.com/reservation-v/short-link/internal/http"
	"github.com/reservation-v/short-link/internal/service"
	"github.com/reservation-v/short-link/internal/storage"
	"github.com/reservation-v/short-link/internal/storage/memory"
	"github.com/reservation-v/short-link/internal/storage/postgres"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 10 * time.Second
	idleTimeout       = time.Minute
	shutdownTimeout   = 5 * time.Second
)

// Run wires the application dependencies and serves HTTP until shutdown
func Run(ctx context.Context, cfg config.Config) error {
	st, cleanup, err := newStorage(ctx, cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	svc := service.New(st, cfg.BaseURL)
	handler := apphttp.NewHandler(svc)
	router := apphttp.NewRouter(handler)

	server := &stdhttp.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	serverErrCh := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, stdhttp.ErrServerClosed) {
			serverErrCh <- err
			return
		}
		close(serverErrCh)
	}()

	log.Printf("short-link listening on %s using %s storage", cfg.HTTPAddr, cfg.Storage)

	select {
	case err := <-serverErrCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown http server: %w", err)
		}

		return nil
	}
}

func newStorage(ctx context.Context, cfg config.Config) (storage.Storage, func(), error) {
	switch cfg.Storage {
	case config.StorageMemory:
		return memory.New(), func() {}, nil
	case config.StoragePostgres:
		pool, err := pgxpool.New(ctx, cfg.PostgresDSN)
		if err != nil {
			return nil, nil, fmt.Errorf("create postgres pool: %w", err)
		}

		if err := pool.Ping(ctx); err != nil {
			pool.Close()
			return nil, nil, fmt.Errorf("ping postgres: %w", err)
		}

		return postgres.New(pool), pool.Close, nil
	default:
		return nil, nil, fmt.Errorf("unsupported storage backend %q", cfg.Storage)
	}
}
