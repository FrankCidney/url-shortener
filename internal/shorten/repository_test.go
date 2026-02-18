package shorten

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func newTestData(id string, url string) ShortLink {
	link := ShortLink{
		ID:        id,
		URL:       url,
		CreatedAt: time.Now(),
		Hits:      0,
	}

	return link
}

func TestMemStore_SaveGetList(t *testing.T) {
	store := NewMemStore()
	link := newTestData("abc123", "https://example.com")

	t.Run("save and get", func(t *testing.T) {
		if err := store.Save(link); err != nil {
			t.Fatalf("unexpected error on save: %v", err)
		}

		gotLink, err := store.Get(link.ID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				t.Fatal("expected id to exist")
			} else {
				t.Fatalf("unexpected error on get: %v", err)
			}
		}

		if gotLink.ID != link.ID {
			t.Fatalf("expected id %q, got %q", link.ID, gotLink.ID)
		}

		if gotLink.URL != link.URL {
			t.Fatalf("expected url %q, got %q", link.URL, gotLink.URL)
		}

		if gotLink.Hits != link.Hits {
			t.Fatalf("expected hits %d, got %d", link.Hits, gotLink.Hits)
		}

		if gotLink.CreatedAt.IsZero() {
			t.Fatalf("expected CreatedAt to be set")
		}
	})

	// t.Run("list", func(t *testing.T) {
	// 	store := NewMemStore()
	// 	link1 := newTestData("a", "url1")
	// 	link2 := newTestData("b", "url2")

	// 	if err := store.Save(link1); err != nil {
	// 		t.Fatalf("unexpected error on save: %v", err)
	// 	}

	// 	if err := store.Save(link2); err != nil {
	// 		t.Fatalf("unexpected error on save: %v", err)
	// 	}

	// 	links := store.List()
	// 	if len(links) != 2 {
	// 		t.Fatalf("expected 2 items, got %d", len(links))
	// 	}
	// })
}

func TestMemStore_ConcurrentSaveGet(t *testing.T) {
	store := NewMemStore()
	wg := sync.WaitGroup{}

	n := 1000

	for i := 0; i < n; i++ {
		wg.Add(2)

		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("id%d", i)
			store.Save(newTestData(id, "url"))
		}(i)

		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("id%d", i)
			store.Get(id)
		}(i)
	}

	wg.Wait()
}
