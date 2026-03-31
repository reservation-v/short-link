package memory

import (
	"context"
	"sync"

	"github.com/reservation-v/short-link/internal/domain"
)

type Storage struct {
	mu      sync.RWMutex
	nextID  int64
	urlToID map[string]int64
	idToURL map[int64]string
}

func New() *Storage {
	return &Storage{
		nextID:  1,
		urlToID: make(map[string]int64),
		idToURL: make(map[int64]string),
	}
}

func (s *Storage) CreateOrGet(ctx context.Context, originalURL string) (id int64, created bool, err error) {
	select {
	case <-ctx.Done():
		return 0, false, ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if existingID, ok := s.urlToID[originalURL]; ok {
		return existingID, false, nil
	}

	id = s.nextID
	s.nextID++
	s.urlToID[originalURL] = id
	s.idToURL[id] = originalURL

	return id, true, nil
}

func (s *Storage) GetByID(ctx context.Context, id int64) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	url, ok := s.idToURL[id]
	if !ok {
		return "", domain.ErrNotFound
	}

	return url, nil
}
