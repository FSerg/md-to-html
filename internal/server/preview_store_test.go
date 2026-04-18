package server

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestPreviewStore_OneShot(t *testing.T) {
	t.Parallel()

	store := NewPreviewStore(time.Hour)
	id := store.Put([]byte("<h1>Hello</h1>"), "text/html; charset=utf-8", "hello.html")

	item, ok := store.Take(id)
	if !ok {
		t.Fatalf("expected first take to succeed")
	}
	if got := string(item.html); got != "<h1>Hello</h1>" {
		t.Fatalf("unexpected html: %q", got)
	}

	if _, ok := store.Take(id); ok {
		t.Fatalf("expected second take to miss")
	}
}

func TestPreviewStore_TTL(t *testing.T) {
	t.Parallel()

	store := NewPreviewStore(10 * time.Millisecond)
	id := store.Put([]byte("expired"), "text/html; charset=utf-8", "expired.html")

	time.Sleep(30 * time.Millisecond)
	store.cleanupExpired(time.Now())

	if _, ok := store.Take(id); ok {
		t.Fatalf("expected expired item to be removed")
	}
}

func TestPreviewStore_Concurrent(t *testing.T) {
	t.Parallel()

	store := NewPreviewStore(time.Hour)
	var wg sync.WaitGroup

	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := store.Put([]byte("payload"), "text/html; charset=utf-8", "payload.html")
			store.Take(id)
		}()
	}

	wg.Wait()
}

func TestPreviewStore_JanitorStopsWithContext(t *testing.T) {
	t.Parallel()

	store := NewPreviewStore(time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		store.janitor(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("janitor did not stop after context cancellation")
	}
}
