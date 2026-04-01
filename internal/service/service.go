package service

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/reservation-v/short-link/internal/codec"
	"github.com/reservation-v/short-link/internal/domain"
	"github.com/reservation-v/short-link/internal/storage"
)

const defaultBaseURL = "http://localhost:8080"

type Service struct {
	storage storage.Storage
	baseURL string
}

type CreateResult struct {
	URL      string
	Code     string
	ShortURL string
	Created  bool
}

func New(st storage.Storage, baseURL string) *Service {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultBaseURL
	}

	return &Service{
		storage: st,
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (s *Service) Create(ctx context.Context, originalURL string) (CreateResult, error) {
	if err := validateOriginalURL(originalURL); err != nil {
		return CreateResult{}, err
	}

	id, created, err := s.storage.CreateOrGet(ctx, originalURL)
	if err != nil {
		return CreateResult{}, err
	}

	code, err := codec.EncodeID(id)
	if err != nil {
		return CreateResult{}, err
	}

	return CreateResult{
		URL:      originalURL,
		Code:     code,
		ShortURL: s.baseURL + "/" + code,
		Created:  created,
	}, nil
}

func (s *Service) Resolve(ctx context.Context, code string) (string, error) {
	id, err := codec.DecodeCode(code)
	if err != nil {
		if errors.Is(err, codec.ErrInvalidCode) {
			return "", domain.ErrInvalidCode
		}
		return "", err
	}

	originalURL, err := s.storage.GetByID(ctx, id)
	if err != nil {
		return "", err
	}

	return originalURL, nil
}

func validateOriginalURL(originalURL string) error {
	if strings.TrimSpace(originalURL) == "" {
		return domain.ErrInvalidURL
	}

	parsedURL, err := url.ParseRequestURI(originalURL)
	if err != nil {
		return domain.ErrInvalidURL
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return domain.ErrInvalidURL
	}

	if parsedURL.Host == "" {
		return domain.ErrInvalidURL
	}

	return nil
}
