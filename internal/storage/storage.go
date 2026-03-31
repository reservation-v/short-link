package storage

import "context"

type Storage interface {
	CreateOrGet(ctx context.Context, originalURL string) (id int64, created bool, err error)
	GetByID(ctx context.Context, id int64) (string, error)
}
