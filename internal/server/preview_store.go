package server

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

const janitorInterval = 5 * time.Minute

type PreviewStore struct {
	mu    sync.Mutex
	items map[string]previewItem
	ttl   time.Duration
	now   func() time.Time
}

type previewItem struct {
	html     []byte
	mime     string
	filename string
	expires  time.Time
}

func NewPreviewStore(ttl time.Duration) *PreviewStore {
	return &PreviewStore{
		items: make(map[string]previewItem),
		ttl:   ttl,
		now:   time.Now,
	}
}

func (s *PreviewStore) Put(html []byte, mime, filename string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := uuid.NewString()
	s.items[id] = previewItem{
		html:     append([]byte(nil), html...),
		mime:     mime,
		filename: filename,
		expires:  s.now().Add(s.ttl),
	}

	return id
}

func (s *PreviewStore) Take(id string) (previewItem, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[id]
	if !ok {
		return previewItem{}, false
	}

	delete(s.items, id)
	if s.now().After(item.expires) {
		return previewItem{}, false
	}

	item.html = append([]byte(nil), item.html...)
	return item, true
}

func (s *PreviewStore) janitor(ctx context.Context) {
	ticker := time.NewTicker(janitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.cleanupExpired(now)
		}
	}
}

func (s *PreviewStore) cleanupExpired(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, item := range s.items {
		if now.After(item.expires) {
			delete(s.items, id)
		}
	}
}
