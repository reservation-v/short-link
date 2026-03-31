package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/reservation-v/short-link/internal/domain"
)

func TestStorage_CreateOrGetAndGetByID(t *testing.T) {
	s := New()
	ctx := context.Background()

	id, created, err := s.CreateOrGet(ctx, "https://example.com/some/path")
	if err != nil {
		t.Fatalf("CreateOrGet returned error: %v", err)
	}
	if !created {
		t.Fatalf("CreateOrGet created = false, want true")
	}
	if id != 1 {
		t.Fatalf("CreateOrGet id = %d, want 1", id)
	}

	url, err := s.GetByID(ctx, id)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if url != "https://example.com/some/path" {
		t.Fatalf("GetByID url = %q, want %q", url, "https://example.com/some/path")
	}
}

func TestStorage_CreateOrGet_DuplicateURL(t *testing.T) {
	s := New()
	ctx := context.Background()
	url := "https://example.com"

	id1, created1, err := s.CreateOrGet(ctx, url)
	if err != nil {
		t.Fatalf("first CreateOrGet returned error: %v", err)
	}
	if !created1 {
		t.Fatalf("first CreateOrGet created = false, want true")
	}

	id2, created2, err := s.CreateOrGet(ctx, url)
	if err != nil {
		t.Fatalf("second CreateOrGet returned error: %v", err)
	}
	if created2 {
		t.Fatalf("second CreateOrGet created = true, want false")
	}
	if id1 != id2 {
		t.Fatalf("duplicate URL returned different ids: %d vs %d", id1, id2)
	}
}

func TestStorage_GetByID_NotFound(t *testing.T) {
	s := New()

	_, err := s.GetByID(context.Background(), 42)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("GetByID error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestStorage_ContextCanceled(t *testing.T) {
	s := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	id, created, err := s.CreateOrGet(ctx, "https://example.com")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("CreateOrGet error = %v, want %v", err, context.Canceled)
	}
	if id != 0 {
		t.Fatalf("CreateOrGet id = %d, want 0", id)
	}
	if created {
		t.Fatalf("CreateOrGet created = true, want false")
	}

	_, err = s.GetByID(ctx, 1)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("GetByID error = %v, want %v", err, context.Canceled)
	}
}

func TestStorage_CreateOrGet_ConcurrentSameURL(t *testing.T) {
	s := New()
	ctx := context.Background()
	url := "https://example.com/concurrent"
	const workers = 100

	type result struct {
		id      int64
		created bool
		err     error
	}

	results := make(chan result, workers)
	wg := &sync.WaitGroup{}

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, created, err := s.CreateOrGet(ctx, url)
			results <- result{id: id, created: created, err: err}
		}()
	}

	wg.Wait()
	close(results)

	createdCount := 0
	uniqueIDs := make(map[int64]struct{})

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

func TestStorage_CreateOrGet_ConcurrentDifferentURLs(t *testing.T) {
	s := New()
	ctx := context.Background()
	const workers = 100

	type result struct {
		id      int64
		url     string
		created bool
		err     error
	}

	results := make(chan result, workers)
	var wg sync.WaitGroup

	for i := range workers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			url := fmt.Sprintf("https://example.com/%d", i)
			id, created, err := s.CreateOrGet(ctx, url)
			results <- result{id: id, url: url, created: created, err: err}
		}(i)
	}

	wg.Wait()
	close(results)

	uniqueIDs := make(map[int64]string, workers)
	for res := range results {
		if res.err != nil {
			t.Fatalf("CreateOrGet returned error: %v", res.err)
		}
		if !res.created {
			t.Fatalf("CreateOrGet created = false for new URL %q", res.url)
		}
		if _, exists := uniqueIDs[res.id]; exists {
			t.Fatalf("duplicate id generated for different URLs: %d", res.id)
		}
		uniqueIDs[res.id] = res.url
	}

	if len(uniqueIDs) != workers {
		t.Fatalf("unique id count = %d, want %d", len(uniqueIDs), workers)
	}
}
