package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/reservation-v/short-link/internal/domain"
)

const (
	insertLinkQuery = `
		INSERT INTO links (original_url)
		VALUES ($1)
		ON CONFLICT (original_url) DO NOTHING
		RETURNING id
	`
	selectLinkIDByURLQuery = `
		SELECT id
		FROM links
		WHERE original_url = $1
	`
	selectOriginalURLByIDQuery = `
		SELECT original_url
		FROM links
		WHERE id = $1
	`
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Storage {
	return &Storage{pool: pool}
}

func (s *Storage) CreateOrGet(ctx context.Context, originalURL string) (id int64, created bool, err error) {
	err = s.pool.QueryRow(ctx, insertLinkQuery, originalURL).Scan(&id)
	switch {
	case err == nil:
		return id, true, nil
	case !errors.Is(err, pgx.ErrNoRows):
		return 0, false, err
	}

	err = s.pool.QueryRow(ctx, selectLinkIDByURLQuery, originalURL).Scan(&id)
	if err != nil {
		return 0, false, err
	}

	return id, false, nil
}

func (s *Storage) GetByID(ctx context.Context, id int64) (string, error) {
	var originalURL string

	err := s.pool.QueryRow(ctx, selectOriginalURLByIDQuery, id).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrNotFound
		}

		return "", err
	}

	return originalURL, nil
}
