//go:build integration

package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reservation-v/short-link/internal/domain"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	startedContainer, dsn, err := startPostgresContainer(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "start postgres test container: %v\n", err)
		os.Exit(1)
	}
	container = startedContainer

	startedPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		cleanupContainer(ctx)
		fmt.Fprintf(os.Stderr, "create postgres pool: %v\n", err)
		os.Exit(1)
	}
	pool = startedPool

	pingCtx, cancelPing := context.WithTimeout(ctx, queryTimeout)
	err = pool.Ping(pingCtx)
	cancelPing()
	if err != nil {
		pool.Close()
		cleanupContainer(ctx)
		fmt.Fprintf(os.Stderr, "ping postgres: %v\n", err)
		os.Exit(1)
	}

	if err := applyMigrations(dsn); err != nil {
		pool.Close()
		cleanupContainer(ctx)
		fmt.Fprintf(os.Stderr, "apply migrations: %v\n", err)
		os.Exit(1)
	}

	storage = New(pool)

	exitCode := m.Run()

	pool.Close()
	cleanupContainer(ctx)

	os.Exit(exitCode)
}

func TestStorageIntegration_CreateOrGet_NewAndExisting(t *testing.T) {
	resetLinksTable(t)

	ctx := context.Background()
	url := "https://example.com/some/path"

	id1, created1, err := storage.CreateOrGet(ctx, url)
	if err != nil {
		t.Fatalf("first CreateOrGet returned error: %v", err)
	}
	if !created1 {
		t.Fatalf("first CreateOrGet created = false, want true")
	}
	if id1 <= 0 {
		t.Fatalf("first CreateOrGet id = %d, want > 0", id1)
	}

	id2, created2, err := storage.CreateOrGet(ctx, url)
	if err != nil {
		t.Fatalf("second CreateOrGet returned error: %v", err)
	}
	if created2 {
		t.Fatalf("second CreateOrGet created = true, want false")
	}
	if id2 != id1 {
		t.Fatalf("second CreateOrGet id = %d, want %d", id2, id1)
	}
}

func TestStorageIntegration_GetByID_Success(t *testing.T) {
	resetLinksTable(t)

	ctx := context.Background()
	wantURL := "https://example.com/path"

	id, created, err := storage.CreateOrGet(ctx, wantURL)
	if err != nil {
		t.Fatalf("CreateOrGet returned error: %v", err)
	}
	if !created {
		t.Fatalf("CreateOrGet created = false, want true")
	}

	gotURL, err := storage.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if gotURL != wantURL {
		t.Fatalf("GetByID url = %q, want %q", gotURL, wantURL)
	}
}

func TestStorageIntegration_GetByID_NotFound(t *testing.T) {
	resetLinksTable(t)

	_, err := storage.GetByID(context.Background(), 9999)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("GetByID error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestStorageIntegration_CreateOrGet_ConcurrentSameURL(t *testing.T) {
	resetLinksTable(t)

	const workers = 50
	url := "https://example.com/concurrent"

	type result struct {
		id      int64
		created bool
		err     error
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := make(chan struct{})
	results := make(chan result, workers)
	var wg sync.WaitGroup

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			id, created, err := storage.CreateOrGet(ctx, url)
			results <- result{id: id, created: created, err: err}
		}()
	}

	close(start)
	wg.Wait()
	close(results)

	createdCount := 0
	uniqueIDs := make(map[int64]struct{}, workers)

	for res := range results {
		if res.err != nil {
			t.Fatalf("CreateOrGet returned error: %v", res.err)
		}
		if res.created {
			createdCount++
		}
		uniqueIDs[res.id] = struct{}{}
	}

	if len(uniqueIDs) != 1 {
		t.Fatalf("same URL returned different ids: %v", uniqueIDs)
	}
	if createdCount != 1 {
		t.Fatalf("created count = %d, want 1", createdCount)
	}
}
