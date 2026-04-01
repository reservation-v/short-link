package http

import (
	"context"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/reservation-v/short-link/internal/service"
)

func TestRouter_MethodNotAllowed(t *testing.T) {
	svc := &stubService{
		createFn: func(ctx context.Context, originalURL string) (service.CreateResult, error) {
			t.Fatalf("service.Create should not be called")
			return service.CreateResult{}, nil
		},
		resolveFn: func(ctx context.Context, code string) (string, error) {
			t.Fatalf("service.Resolve should not be called")
			return "", nil
		},
	}

	router := NewRouter(NewHandler(svc))

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   "GET on /links",
			method: stdhttp.MethodGet,
			path:   "/links",
		},
		{
			name:   "POST on /links/{code}",
			method: stdhttp.MethodPost,
			path:   "/links/aaaaaaaaab",
			body:   `{"url":"https://example.com"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyReader *strings.Reader
			if tt.body == "" {
				bodyReader = strings.NewReader("")
			} else {
				bodyReader = strings.NewReader(tt.body)
			}

			req := httptest.NewRequest(tt.method, tt.path, bodyReader)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			if rr.Code != stdhttp.StatusMethodNotAllowed {
				t.Fatalf("status = %d, want %d", rr.Code, stdhttp.StatusMethodNotAllowed)
			}
		})
	}
}

func TestRouter_NotFound(t *testing.T) {
	svc := &stubService{
		createFn: func(ctx context.Context, originalURL string) (service.CreateResult, error) {
			t.Fatalf("service.Create should not be called")
			return service.CreateResult{}, nil
		},
		resolveFn: func(ctx context.Context, code string) (string, error) {
			t.Fatalf("service.Resolve should not be called")
			return "", nil
		},
	}

	router := NewRouter(NewHandler(svc))

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "unknown path",
			method: stdhttp.MethodGet,
			path:   "/unknown",
		},
		{
			name:   "links trailing slash",
			method: stdhttp.MethodGet,
			path:   "/links/",
		},
		{
			name:   "links trailing slash post",
			method: stdhttp.MethodPost,
			path:   "/links/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			if rr.Code != stdhttp.StatusNotFound {
				t.Fatalf("status = %d, want %d", rr.Code, stdhttp.StatusNotFound)
			}
		})
	}
}
