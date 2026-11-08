package store

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/SaraAbreu/postboard/api/internal/models"
)

// MemoryStore guarda los posts en memoria. Útil para tests y para probar sin base de datos.
type MemoryStore struct {
	mu     sync.Mutex
	posts  map[int64]models.Post
	nextID int64
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{posts: map[int64]models.Post{}, nextID: 1}
}

func (s *MemoryStore) List(_ context.Context) ([]models.Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]models.Post, 0, len(s.posts))
	for _, p := range s.posts {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out, nil
}

func (s *MemoryStore) Get(_ context.Context, id int64) (models.Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.posts[id]
	if !ok {
		return models.Post{}, ErrNotFound
	}
	return p, nil
}

func (s *MemoryStore) Create(_ context.Context, in models.PostInput) (models.Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	p := models.Post{ID: s.nextID, Title: in.Title, Content: in.Content, Platform: in.Platform,
		Status: in.Status, ScheduledAt: in.ScheduledAt, CreatedAt: now, UpdatedAt: now}
	s.posts[p.ID] = p
	s.nextID++
	return p, nil
}

func (s *MemoryStore) Update(_ context.Context, id int64, in models.PostInput) (models.Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.posts[id]
	if !ok {
		return models.Post{}, ErrNotFound
	}
	p.Title, p.Content, p.Platform, p.Status, p.ScheduledAt = in.Title, in.Content, in.Platform, in.Status, in.ScheduledAt
	p.UpdatedAt = time.Now().UTC()
	s.posts[id] = p
	return p, nil
}

func (s *MemoryStore) Delete(_ context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.posts[id]; !ok {
		return ErrNotFound
	}
	delete(s.posts, id)
	return nil
}
