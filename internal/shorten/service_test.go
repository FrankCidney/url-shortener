package shorten

import (
	"errors"
	"testing"
)

// Sequence Generator is preloaded with IDs.
// Each call to Next() returns the next  ID, so it's fully deterministic (i.e., you always know which ID will be next, so you can test for
// very specific behaviour)
type SequenceGenerator struct {
	ids []string
	i int
}

func NewSequenceGenerator(ids ...string) *SequenceGenerator {
	return &SequenceGenerator{ids: ids}
}

func (g *SequenceGenerator) Next(_ string) (string, error) {
	if g.i >= len(g.ids) {
		return "", errors.New("no more ids")
	}

	id := g.ids[g.i]
	g.i++
	return id, nil
}

// Mock Generator always returns the exact same ID. Therefore, every call to Next() except for the first one, will generate a collision
type MockGenerator struct {
	id string
}

func (mg *MockGenerator) Next(_ string) (string, error) {
	return "fake-id", nil
}

func NewMockGenerator() *MockGenerator {
	return &MockGenerator{}
}

// Helper: Sets up a shortener service instance to use for the tests
func newTestShortener(t *testing.T, generator IDGenerator) *Shortener {
	t.Helper()

	store := NewMemStore()
	return NewShortener(store, generator)
}

func TestCreate(t *testing.T) {
	// Test Cases
	// Test: Successfully created short link
	t.Run("Valid URL", func(t *testing.T) {
		// Setup
		shortener := newTestShortener(t, NewBase62Generator())
		url := "https://example.com"

		link, err := shortener.Create(url)
		if err != nil {
			t.Fatalf("expected nil err, got %v", err)
		}

		// Check that ID was generated
		if link.ID == "" {
			t.Error(`expected id, got ""`)
		}

		// Check that URL was stored
		if link.URL != url {
			t.Errorf("expected url %s, got %s", url, link.URL)
		}

		if link.Hits != 0 {
			t.Fatalf("expected hits=0, got hits=%d", link.Hits)
		}

		if link.CreatedAt.IsZero() {
			t.Fatalf("expected CreatedAt to be set")
		}
	})

	// Test: Invalid URL
	t.Run("Invalid URL", func(t *testing.T) {
		// setup
		shortener := newTestShortener(t, NewBase62Generator())

		tests := []struct{
			name string
			url string
		} {
			{"empty url", ""},
			{"missing scheme", "example.com"},
			{"missing host", "https://"},
			{"relative path", "/path/to/resource"},
			{"invalid URL scheme", "ftp://example.com"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := shortener.Create(tt.url)

				if err == nil {
					t.Fatalf("expected error for url %q, got nil", tt.url)
				}

				if !errors.Is(err, ErrInvalidURL) {
					t.Fatalf("expected ErrInvalidURL, got %v", err)
				}
			})
		}
	})
	
	// Test: Collision
	t.Run("Collision", func(t *testing.T) {
		// First ID collides, second succeeds.
		// This is the typical expected behaviour when there is a collision. Retries should fix it.
		generator := NewSequenceGenerator("id1", "id1", "id2")
		shortener := newTestShortener(t, generator)

		url := "https://example.com"

		// Call Create() twice to generate a collision
		link, err := shortener.Create(url)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		link, err = shortener.Create(url)	
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		if link.ID != "id2" {
			t.Fatalf("expected final id to be %q, got %q", "id2", link.ID)
		}
	})

	// Test: Too many collisions
	t.Run("Too many collisions", func(t *testing.T) {	
		shortener := newTestShortener(t, NewMockGenerator())
		url := "https://example.com"

		// Create ID
		_, err := shortener.Create(url)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		// Create the ID again
		// Mock generator returns the same id every time, so it will generate collisions infinitely - The expected behaviour is that you run
		// out of retry attempts
		_, err = shortener.Create(url)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, ErrTooManyCollisions) {
			t.Fatalf("expected ErrTooManyCollisions, got %v", err)
		}
	})
}

// -----------------------------------------------------------
func TestResolve(t *testing.T) {
	// Test: ID found
	t.Run("Short ID exists", func(t *testing.T) {
		shortener := newTestShortener(t, NewHashGenerator(8))
		url := "https://example.com"

		// 1) Create ID
		link, err := shortener.Create(url)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		// 2) Resolve
		gotURL, err := shortener.Resolve(link.ID)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		// Check that you got back the correct URL
		if url != gotURL {
			t.Errorf("expected url %q, got %q", url, gotURL)
		}

		// Check whether hit count was incremented
		link, err = shortener.Stats(link.ID)
		if err != nil {
			t.Fatalf("stats failed: %v", err)
		}

		if link.Hits != 1 {
			t.Fatalf("expected hits = 1, got %d", link.Hits)
		}
	})

	// Test: ID not found
	t.Run("Short ID not found", func(t *testing.T) {
		shortener := newTestShortener(t, NewHashGenerator(8))
		_, err := shortener.Resolve("fake-id")
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected error: %v; got %v", ErrNotFound, err)
		}
	})
}

func TestStats(t *testing.T) {
	t.Run("Happy path: ID found", func(t *testing.T) {
		shortener := newTestShortener(t, NewHashGenerator(8))
		url := "https://example.com"

		// 1) Create ID
		link, err := shortener.Create(url)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		// 2) Resolve thrice, so expected hits == 3
		for i := 0; i < 3; i++ {
			if _, err := shortener.Resolve(link.ID); err != nil {
				t.Fatalf("setup resolve failed: %v", err)
			}
		}
		
		link, err = shortener.Stats(link.ID)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		if link.URL != url {
			t.Errorf("expected url = %q, got %q", url, link.URL)
		}

		if link.Hits != 3 {
			t.Errorf("expected hit count = 3, got %d", link.Hits)
		}
	})

	t.Run("Short ID not found", func(t *testing.T) {
		shortener := newTestShortener(t, NewHashGenerator(8))

		_, err := shortener.Stats("fake-id")
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected error: %v; got %v", ErrNotFound, err)
		}
	})
}
