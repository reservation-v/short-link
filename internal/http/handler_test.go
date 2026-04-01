package http

import (
	"context"
	"encoding/json"
	"errors"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/reservation-v/short-link/internal/domain"
	"github.com/reservation-v/short-link/internal/service"
)

type stubService struct {
	createFn  func(ctx context.Context, originalURL string) (service.CreateResult, error)
	resolveFn func(ctx context.Context, code string) (string, error)
}

func (s *stubService) Create(ctx context.Context, originalURL string) (service.CreateResult, error) {
	if s.createFn == nil {
		return service.CreateResult{}, errors.New("unexpected Create call")
	}
	return s.createFn(ctx, originalURL)
}

func (s *stubService) Resolve(ctx context.Context, code string) (string, error) {
	if s.resolveFn == nil {
		return "", errors.New("unexpected Resolve call")
	}
	return s.resolveFn(ctx, code)
}

func TestCreateLink_Created(t *testing.T) {
	svc := &stubService{
		createFn: func(ctx context.Context, originalURL string) (service.CreateResult, error) {
			if originalURL != "https://example.com/some/path" {
				t.Fatalf("originalURL = %q, want %q", originalURL, "https://example.com/some/path")
			}
			return service.CreateResult{
				URL:      "https://example.com/some/path",
				Code:     "aaaaaaaaab",
				ShortURL: "http://localhost:8080/aaaaaaaaab",
				Created:  true,
			}, nil
		},
	}

	router := NewRouter(NewHandler(svc))
	req := httptest.NewRequest(stdhttp.MethodPost, "/links", strings.NewReader(`{"url":"https://example.com/some/path"}`))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != stdhttp.StatusCreated {
		t.Fatalf("status = %d, want %d", rr.Code, stdhttp.StatusCreated)
	}
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("content-type = %q, want %q", rr.Header().Get("Content-Type"), "application/json")
	}

	var got createLinkResponse
	decodeResponseBody(t, rr.Body.String(), &got)
	if got.URL != "https://example.com/some/path" {
		t.Fatalf("url = %q, want %q", got.URL, "https://example.com/some/path")
	}
	if got.Code != "aaaaaaaaab" {
		t.Fatalf("code = %q, want %q", got.Code, "aaaaaaaaab")
	}
	if got.ShortURL != "http://localhost:8080/aaaaaaaaab" {
		t.Fatalf("short_url = %q, want %q", got.ShortURL, "http://localhost:8080/aaaaaaaaab")
	}
}

func TestCreateLink_Existing(t *testing.T) {
	svc := &stubService{
		createFn: func(ctx context.Context, originalURL string) (service.CreateResult, error) {
			return service.CreateResult{
				URL:      "https://example.com/some/path",
				Code:     "aaaaaaaaab",
				ShortURL: "http://localhost:8080/aaaaaaaaab",
				Created:  false,
			}, nil
		},
	}

	router := NewRouter(NewHandler(svc))
	req := httptest.NewRequest(stdhttp.MethodPost, "/links", strings.NewReader(`{"url":"https://example.com/some/path"}`))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != stdhttp.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, stdhttp.StatusOK)
	}
}

func TestCreateLink_InvalidRequest(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "malformed-json",
			body: `{"url":"https://example.com"`,
		},
		{
			name: "unknown-field",
			body: `{"url":"https://example.com","extra":"x"}`,
		},
		{
			name: "trailing-json",
			body: `{"url":"https://example.com"}{"url":"https://another.com"}`,
		},
		{
			name: "body-too-large",
			body: `{"url":"https://example.com/` + strings.Repeat("a", maxCreateLinkRequestBody) + `"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			svc := &stubService{
				createFn: func(ctx context.Context, originalURL string) (service.CreateResult, error) {
					called = true
					return service.CreateResult{}, nil
				},
			}

			router := NewRouter(NewHandler(svc))
			req := httptest.NewRequest(stdhttp.MethodPost, "/links", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			if rr.Code != stdhttp.StatusBadRequest {
				t.Fatalf("status = %d, want %d", rr.Code, stdhttp.StatusBadRequest)
			}

			var got errorResponse
			decodeResponseBody(t, rr.Body.String(), &got)
			if got.Error != "invalid request" {
				t.Fatalf("error = %q, want %q", got.Error, "invalid request")
			}
			if called {
				t.Fatalf("service.Create was called for invalid request")
			}
		})
	}
}

func TestCreateLink_InvalidURL(t *testing.T) {
	svc := &stubService{
		createFn: func(ctx context.Context, originalURL string) (service.CreateResult, error) {
			return service.CreateResult{}, domain.ErrInvalidURL
		},
	}

	router := NewRouter(NewHandler(svc))
	req := httptest.NewRequest(stdhttp.MethodPost, "/links", strings.NewReader(`{"url":"not-a-url"}`))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != stdhttp.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, stdhttp.StatusBadRequest)
	}

	var got errorResponse
	decodeResponseBody(t, rr.Body.String(), &got)
	if got.Error != "invalid request" {
		t.Fatalf("error = %q, want %q", got.Error, "invalid request")
	}
}

func TestCreateLink_InternalError(t *testing.T) {
	svc := &stubService{
		createFn: func(ctx context.Context, originalURL string) (service.CreateResult, error) {
			return service.CreateResult{}, errors.New("db is down")
		},
	}

	router := NewRouter(NewHandler(svc))
	req := httptest.NewRequest(stdhttp.MethodPost, "/links", strings.NewReader(`{"url":"https://example.com/some/path"}`))
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != stdhttp.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, stdhttp.StatusInternalServerError)
	}

	var got errorResponse
	decodeResponseBody(t, rr.Body.String(), &got)
	if got.Error != "internal error" {
		t.Fatalf("error = %q, want %q", got.Error, "internal error")
	}
}

func TestResolveLink_Success(t *testing.T) {
	svc := &stubService{
		resolveFn: func(ctx context.Context, code string) (string, error) {
			if code != "aaaaaaaaab" {
				t.Fatalf("code = %q, want %q", code, "aaaaaaaaab")
			}
			return "https://example.com/some/path", nil
		},
	}

	router := NewRouter(NewHandler(svc))
	req := httptest.NewRequest(stdhttp.MethodGet, "/links/aaaaaaaaab", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != stdhttp.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, stdhttp.StatusOK)
	}

	var got resolveLinkResponse
	decodeResponseBody(t, rr.Body.String(), &got)
	if got.URL != "https://example.com/some/path" {
		t.Fatalf("url = %q, want %q", got.URL, "https://example.com/some/path")
	}
}

func TestResolveLink_Errors(t *testing.T) {
	tests := []struct {
		name       string
		serviceErr error
		wantStatus int
		wantError  string
	}{
		{
			name:       "invalid-code",
			serviceErr: domain.ErrInvalidCode,
			wantStatus: stdhttp.StatusBadRequest,
			wantError:  "invalid code",
		},
		{
			name:       "not-found",
			serviceErr: domain.ErrNotFound,
			wantStatus: stdhttp.StatusNotFound,
			wantError:  "not found",
		},
		{
			name:       "internal-error",
			serviceErr: errors.New("storage failed"),
			wantStatus: stdhttp.StatusInternalServerError,
			wantError:  "internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &stubService{
				resolveFn: func(ctx context.Context, code string) (string, error) {
					return "", tt.serviceErr
				},
			}

			router := NewRouter(NewHandler(svc))
			req := httptest.NewRequest(stdhttp.MethodGet, "/links/aaaaaaaaab", nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tt.wantStatus)
			}

			var got errorResponse
			decodeResponseBody(t, rr.Body.String(), &got)
			if got.Error != tt.wantError {
				t.Fatalf("error = %q, want %q", got.Error, tt.wantError)
			}
		})
	}
}

func decodeResponseBody(t *testing.T, body string, out any) {
	t.Helper()
	if err := json.Unmarshal([]byte(body), out); err != nil {
		t.Fatalf("failed to decode response body %q: %v", body, err)
	}
}
