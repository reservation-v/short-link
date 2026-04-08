package service

import (
	"context"
	"errors"
	"testing"

	"github.com/reservation-v/short-link/internal/domain"
)

type stubStorage struct {
	createOrGet func(ctx context.Context, originalURL string) (int64, bool, error)
	getByID     func(ctx context.Context, id int64) (string, error)
}

func (m *stubStorage) CreateOrGet(ctx context.Context, originalURL string) (int64, bool, error) {
	if m.createOrGet == nil {
		return 0, false, errors.New("unexpected CreateOrGet call")
	}
	return m.createOrGet(ctx, originalURL)
}

func (m *stubStorage) GetByID(ctx context.Context, id int64) (string, error) {
	if m.getByID == nil {
		return "", errors.New("unexpected GetByID call")
	}
	return m.getByID(ctx, id)
}

func TestServiceCreate_SuccessNew(t *testing.T) {
	st := &stubStorage{
		createOrGet: func(ctx context.Context, originalURL string) (int64, bool, error) {
			if originalURL != "https://example.com/path" {
				t.Fatalf("originalURL = %q, want %q", originalURL, "https://example.com/path")
			}
			return 1, true, nil
		},
	}
	svc := New(st, "http://localhost:8080")

	got, err := svc.Create(context.Background(), "https://example.com/path")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if got.URL != "https://example.com/path" {
		t.Fatalf("URL = %q, want %q", got.URL, "https://example.com/path")
	}
	if got.Code != "aaaaaaaaab" {
		t.Fatalf("Code = %q, want %q", got.Code, "aaaaaaaaab")
	}
	if got.ShortURL != "http://localhost:8080/aaaaaaaaab" {
		t.Fatalf("ShortURL = %q, want %q", got.ShortURL, "http://localhost:8080/aaaaaaaaab")
	}
	if !got.Created {
		t.Fatalf("Created = false, want true")
	}
}

func TestServiceCreate_SuccessExisting(t *testing.T) {
	st := &stubStorage{
		createOrGet: func(ctx context.Context, originalURL string) (int64, bool, error) {
			return 1, false, nil
		},
	}
	svc := New(st, "http://localhost:8080")

	got, err := svc.Create(context.Background(), "https://example.com/path")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if got.Created {
		t.Fatalf("Created = true, want false")
	}
}

func TestServiceCreate_InvalidURL(t *testing.T) {
	tests := []string{
		"",
		"   ",
		"not-a-url",
		"ftp://example.com/path",
		"http:///path-only",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			called := false
			st := &stubStorage{
				createOrGet: func(ctx context.Context, originalURL string) (int64, bool, error) {
					called = true
					return 0, false, nil
				},
			}
			svc := New(st, "http://localhost:8080")

			_, err := svc.Create(context.Background(), input)
			if !errors.Is(err, domain.ErrInvalidURL) {
				t.Fatalf("Create error = %v, want %v", err, domain.ErrInvalidURL)
			}
			if called {
				t.Fatalf("storage.CreateOrGet was called for invalid URL %q", input)
			}
		})
	}
}

func TestServiceCreate_StorageError(t *testing.T) {
	storageErr := errors.New("storage failure")
	st := &stubStorage{
		createOrGet: func(ctx context.Context, originalURL string) (int64, bool, error) {
			return 0, false, storageErr
		},
	}
	svc := New(st, "http://localhost:8080")

	_, err := svc.Create(context.Background(), "https://example.com/path")
	if !errors.Is(err, storageErr) {
		t.Fatalf("Create error = %v, want %v", err, storageErr)
	}
}

func TestServiceCreate_BaseURLHandling(t *testing.T) {
	st := &stubStorage{
		createOrGet: func(ctx context.Context, originalURL string) (int64, bool, error) {
			return 1, true, nil
		},
	}

	svcDefault := New(st, "")
	resDefault, err := svcDefault.Create(context.Background(), "https://example.com/path")
	if err != nil {
		t.Fatalf("Create with default baseURL returned error: %v", err)
	}
	if resDefault.ShortURL != "http://localhost:8081/aaaaaaaaab" {
		t.Fatalf("default ShortURL = %q, want %q", resDefault.ShortURL, "http://localhost:8081/aaaaaaaaab")
	}

	svcTrimmed := New(st, "http://localhost:8080/")
	resTrimmed, err := svcTrimmed.Create(context.Background(), "https://example.com/path")
	if err != nil {
		t.Fatalf("Create with trailing-slash baseURL returned error: %v", err)
	}
	if resTrimmed.ShortURL != "http://localhost:8080/aaaaaaaaab" {
		t.Fatalf("trimmed ShortURL = %q, want %q", resTrimmed.ShortURL, "http://localhost:8080/aaaaaaaaab")
	}
}

func TestServiceCreate_ContextCanceled(t *testing.T) {
	st := &stubStorage{
		createOrGet: func(ctx context.Context, originalURL string) (int64, bool, error) {
			return 0, false, ctx.Err()
		},
	}
	svc := New(st, "http://localhost:8080")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.Create(ctx, "https://example.com/path")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Create error = %v, want %v", err, context.Canceled)
	}
}

func TestServiceResolve_Success(t *testing.T) {
	st := &stubStorage{
		getByID: func(ctx context.Context, id int64) (string, error) {
			if id != 1 {
				t.Fatalf("id = %d, want 1", id)
			}
			return "https://example.com/path", nil
		},
	}
	svc := New(st, "http://localhost:8080")

	got, err := svc.Resolve(context.Background(), "aaaaaaaaab")
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if got != "https://example.com/path" {
		t.Fatalf("Resolve url = %q, want %q", got, "https://example.com/path")
	}
}

func TestServiceResolve_InvalidCode(t *testing.T) {
	called := false
	st := &stubStorage{
		getByID: func(ctx context.Context, id int64) (string, error) {
			called = true
			return "", nil
		},
	}
	svc := New(st, "http://localhost:8080")

	_, err := svc.Resolve(context.Background(), "bad")
	if !errors.Is(err, domain.ErrInvalidCode) {
		t.Fatalf("Resolve error = %v, want %v", err, domain.ErrInvalidCode)
	}
	if called {
		t.Fatalf("storage.GetByID was called for invalid code")
	}
}

func TestServiceResolve_NotFound(t *testing.T) {
	st := &stubStorage{
		getByID: func(ctx context.Context, id int64) (string, error) {
			return "", domain.ErrNotFound
		},
	}
	svc := New(st, "http://localhost:8080")

	_, err := svc.Resolve(context.Background(), "aaaaaaaaab")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Resolve error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestServiceResolve_StorageError(t *testing.T) {
	storageErr := errors.New("storage failure")
	st := &stubStorage{
		getByID: func(ctx context.Context, id int64) (string, error) {
			return "", storageErr
		},
	}
	svc := New(st, "http://localhost:8080")

	_, err := svc.Resolve(context.Background(), "aaaaaaaaab")
	if !errors.Is(err, storageErr) {
		t.Fatalf("Resolve error = %v, want %v", err, storageErr)
	}
}

func TestServiceResolve_ContextCanceled(t *testing.T) {
	st := &stubStorage{
		getByID: func(ctx context.Context, id int64) (string, error) {
			return "", ctx.Err()
		},
	}
	svc := New(st, "http://localhost:8080")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.Resolve(ctx, "aaaaaaaaab")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Resolve error = %v, want %v", err, context.Canceled)
	}
}
