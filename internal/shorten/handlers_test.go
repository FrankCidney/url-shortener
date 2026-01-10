package shorten
// package shorten

// import (
// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"strings"
// 	"testing"
// )

// func TestHandleShorten(t *testing.T) {
// 	t.Run("happy path", func(t *testing.T) {
// 		body := `{"url":"https://example.com"}`
// 		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
// 		req.Header.Set("Content-Type", "application/json")

// 		rr := httptest.NewRecorder()
// 		HandleShorten(rr, req)

// 		if status := rr.Code; status != http.StatusOK {
// 			t.Errorf("expected status %d, got %d", http.StatusOK, status)
// 		}

// 		var resp shortenResponse
// 		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
// 			t.Fatalf("failed to decode response: %v", err)
// 		}

// 		if resp.Short == "" {
// 			t.Error("expected short code, got empty string")
// 		}
// 		if resp.URL != "https://example.com" {
// 			t.Errorf("expected URL https://example.com, got %s", resp.URL)
// 		}
// 	})

// 	t.Run("wrong HTTP method", func(t *testing.T) {
// 		req := httptest.NewRequest(http.MethodGet, "/shorten", nil)
// 		rr := httptest.NewRecorder()

// 		HandleShorten(rr, req)

// 		if status := rr.Code; status != http.StatusMethodNotAllowed {
// 			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, status)
// 		}

// 		if allow := rr.Header().Get("Allow"); allow != http.MethodPost {
// 			t.Errorf("expected Allow header %s, got %s", http.MethodPost, allow)
// 		}
// 	})

// 	// syntax error in request body JSON
// 	t.Run("malformed JSON", func(t *testing.T) {
// 		body := `{"url": "https://example.com}` // no closing quote for value
// 		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
// 		req.Header.Set("Content-Type", "application/json")

// 		rr := httptest.NewRecorder()
// 		HandleShorten(rr, req)

// 		if status := rr.Code; status != http.StatusBadRequest {
// 			t.Errorf("expected status %d, got %d", http.StatusBadRequest, status)
// 		}
// 	})

// 	// DisallowUnknownFields error
// 	t.Run("unknown fields in JSON", func(t *testing.T) {
// 		body := `{"url":"https://example.com", "extra":"nope"}`
// 		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
// 		rr := httptest.NewRecorder()

// 		HandleShorten(rr, req)

// 		if status := rr.Code; status != http.StatusBadRequest {
// 			t.Errorf("expected status %d, got %d", http.StatusBadRequest, status)
// 		}
// 	})

// 	t.Run("mutiple JSON objects", func(t *testing.T) {
// 		body := `{"url":"https://example.com"}{"url":"https://another.com"}`
// 		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
// 		rr := httptest.NewRecorder()

// 		HandleShorten(rr, req)

// 		if status := rr.Code; status != http.StatusBadRequest {
// 			t.Errorf("expected status %d, got %d", http.StatusBadRequest, status)
// 		}
// 	})

// 	t.Run("request body too large", func(t *testing.T) {
// 		// Create a valid JSON body larger than maxBodyBytes (1MB)
// 		// Use a very long URL string to exceed the limit

// 		longURL := "https://example.com/" + strings.Repeat("a", maxBodyBytes)
// 		body := `{"url":"` + longURL + `"}`

// 		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
// 		rr := httptest.NewRecorder()

// 		HandleShorten(rr, req)

// 		if status := rr.Code; status != http.StatusRequestEntityTooLarge {
// 			t.Errorf("expected status %d, got %d", http.StatusRequestEntityTooLarge, status)
// 		}
// 	})

// 	t.Run("empty URL", func(t *testing.T) {
// 		body := `"url":""`
// 		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
// 		rr := httptest.NewRecorder()

// 		HandleShorten(rr, req)

// 		if status := rr.Code; status != http.StatusBadRequest {
// 			t.Errorf("expected status %d, got %d", http.StatusBadRequest, status)
// 		}
// 	})

// 	t.Run("invalid URL scheme", func(t *testing.T) {
// 		body := `{"url":"ftp://example.com"}`
// 		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
// 		rr := httptest.NewRecorder()

// 		HandleShorten(rr, req)

// 		if status := rr.Code; status != http.StatusBadRequest {
// 			t.Errorf("expected status %d, got %d", http.StatusBadRequest, status)
// 		}
// 	})

// 	t.Run("URL missing host", func(t *testing.T) {
// 		body := `{"url":"https://"}`
// 		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
// 		rr := httptest.NewRecorder()

// 		HandleShorten(rr, req)

// 		if status := rr.Code; status != http.StatusBadRequest {
// 			t.Errorf("expected status %d, got %d", http.StatusBadRequest, status)
// 		}
// 	})

// 	t.Run("relative URL rejected", func(t *testing.T) {
// 		body := `{"url":"/path/to/resource"}`
// 		req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(body))
// 		rr := httptest.NewRecorder()

// 		HandleShorten(rr, req)

// 		if status := rr.Code; status != http.StatusBadRequest {
// 			t.Errorf("expected status %d, got %d", http.StatusBadRequest, status)
// 		}
// 	})
// 	// unexpected EOF

// 	// do the request fields match the expected fields?
// 	// unmarshalTypeErr error - i.e., do the request field types match the expected field types?

// 	// is the url structure correct? i.e. valid url
// 	// does it have a scheme?
// 	// does it have a host?

// 	// is the url scheme allowed?
// }

// func TestValidateURL(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		url     string
// 		wantErr bool
// 	}{
// 		{"valid https", "https://example.com", false},
// 		{"valid http", "http://example.com", false},
// 		{"valid with path", "https://example.com/path", false},
// 		{"empty url", "", true},
// 		{"missing scheme", "example.com", true},
// 		{"missing host", "https://", true},
// 		{"relative path", "/path/to/resource", true},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			err := validateURL(tt.url)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("validateURL() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }
