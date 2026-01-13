package shorten

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestGenerator() IDGenerator {
	return NewBase62Generator()
}

func TestHandleShorten(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		shortener := newTestShortener(t, newTestGenerator())
		handler := NewHandler(shortener)

		body := `{"url":"https://example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler.HandleShorten(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, status)
		}

		var resp shortenResponse
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Short == "" {
			t.Error("expected short code, got empty string")
		}
		if resp.URL != "https://example.com" {
			t.Errorf("expected URL https://example.com, got %s", resp.URL)
		}
	})

	t.Run("wrong HTTP method", func(t *testing.T) {
		shortener := newTestShortener(t, newTestGenerator())
		handler := NewHandler(shortener)

		req := httptest.NewRequest(http.MethodGet, "/shorten", nil)
		rr := httptest.NewRecorder()

		handler.HandleShorten(rr, req)

		if status := rr.Code; status != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, status)
		}

		if allow := rr.Header().Get("Allow"); allow != http.MethodPost {
			t.Errorf("expected Allow header %s, got %s", http.MethodPost, allow)
		}
	})

	// syntax error in request body JSON
	t.Run("malformed JSON", func(t *testing.T) {
		shortener := newTestShortener(t, newTestGenerator())
		handler := NewHandler(shortener)

		body := `{"url": "https://example.com}` // no closing quote for value
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler.HandleShorten(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, status)
		}
	})

	// DisallowUnknownFields error
	t.Run("unknown fields in JSON", func(t *testing.T) {
		shortener := newTestShortener(t, newTestGenerator())
		handler := NewHandler(shortener)

		body := `{"url":"https://example.com", "extra":"nope"}`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
		rr := httptest.NewRecorder()

		handler.HandleShorten(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, status)
		}
	})

	t.Run("mutiple JSON objects", func(t *testing.T) {
		shortener := newTestShortener(t, newTestGenerator())
		handler := NewHandler(shortener)

		body := `{"url":"https://example.com"}{"url":"https://another.com"}`
		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
		rr := httptest.NewRecorder()

		handler.HandleShorten(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, status)
		}
	})

	t.Run("request body too large", func(t *testing.T) {
		shortener := newTestShortener(t, newTestGenerator())
		handler := NewHandler(shortener)

		// Create a valid JSON body larger than maxBodyBytes (1MB)
		// Use a very long URL string to exceed the limit

		longURL := "https://example.com/" + strings.Repeat("a", maxBodyBytes)
		body := `{"url":"` + longURL + `"}`

		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
		rr := httptest.NewRecorder()

		handler.HandleShorten(rr, req)

		if status := rr.Code; status != http.StatusRequestEntityTooLarge {
			t.Errorf("expected status %d, got %d", http.StatusRequestEntityTooLarge, status)
		}
	})
}

func TestHandleRedirect_OK(t *testing.T) {
	shortener := newTestShortener(t, newTestGenerator())
	handler := NewHandler(shortener)
	
	url := "https://example.com"

	// setup
	link, err := shortener.Create(url)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/"+link.ID, nil)
	rr := httptest.NewRecorder()

	handler.HandleRedirect(rr, req)

	res := rr.Result()
	res.Body.Close()

	if res.StatusCode != http.StatusFound {
		t.Fatalf("expected status 302, got %d", res.StatusCode)
	}

	loc := res.Header.Get("Location")
	if loc != link.URL {
		t.Fatalf("expected redirect to %q, got %q", link.URL, loc)
	}
}

func TestHandleRedirect_NotFound(t *testing.T) {
	shortener := newTestShortener(t, newTestGenerator())
	handler := NewHandler(shortener)

	req := httptest.NewRequest(http.MethodGet, "/doesnotexist", nil)
	rr := httptest.NewRecorder()

	handler.HandleRedirect(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, res.StatusCode)
	}
}

func TestHandleRedirect_MissingID(t *testing.T) {
	shortener := newTestShortener(t, newTestGenerator())
	handler := NewHandler(shortener)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.HandleRedirect(rr, req)

	res := rr.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, res.StatusCode)
	}
}

func TestHandleStats(t *testing.T) {
	// Test: Happy path
	t.Run("OK", func(t *testing.T) {
		shortener := newTestShortener(t, newTestGenerator())
		handler := NewHandler(shortener)

		url := "https://example.com"

		link, err := shortener.Create(url)
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		// increment hits
		for i := 0; i < 3; i++ {
			_, err := shortener.Resolve(link.ID)
			if err != nil {
				t.Fatalf("setup resolve failed: %v", err)
			}
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/stats/%s", link.ID), nil)
		rr := httptest.NewRecorder()

		handler.HandleStats(rr, req)

		res := rr.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, res.StatusCode)
		}

		var body statsResponse
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if body.Short != link.ID {
			t.Errorf("expected id %q, got %q", link.ID, body.Short)
		}

		if body.URL != link.URL {
			t.Errorf("expected url %q, got %q", link.URL, body.URL)
		}

		if body.Hits != 2 {
			t.Errorf("expected hits 3, got %d", body.Hits)
		}

		if _, err := time.Parse(time.RFC3339, body.CreatedAt); err != nil {
			t.Fatalf("createdAt is not valid RFC3339: %v", err)
		}
	})

	// Test: Missing id
	t.Run("Missing ID", func(t *testing.T) {
		shortener := newTestShortener(t, newTestGenerator())
		handler := NewHandler(shortener)	

		req := httptest.NewRequest(http.MethodGet, "/stats", nil)
		rr := httptest.NewRecorder()

		handler.HandleStats(rr, req)
		res := rr.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", res.StatusCode)
		}
	})

	// Test: ID not found
	t.Run("NOT FOUND", func(t *testing.T) {
		shortener := newTestShortener(t, newTestGenerator())
		handler := NewHandler(shortener)	

		req := httptest.NewRequest(http.MethodGet, "/stats/madethisup", nil)
		rr := httptest.NewRecorder()

		handler.HandleStats(rr, req)

		res := rr.Result()
		res.Body.Close()

		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", res.StatusCode)
		}
	})	
}