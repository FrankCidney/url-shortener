package shared

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testHandler(called *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
}

func captureLogs(t *testing.T) (*bytes.Buffer, func()) {
	t.Helper()

	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)

	return &buf, func() {
		log.SetOutput(orig)
	}
}

func TestAuth_MissingKey(t *testing.T) {
	var called bool

	handler := Chain(
		testHandler(&called),
		RequestID,
		Logging,
		Auth("secret123"),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}

	if called {
		t.Fatal("handler should not have been called")
	}
}

func TestAuth_ValidKey(t *testing.T) {
	var called bool

	handler := Chain(
		testHandler(&called),
		RequestID,
		Logging,
		Auth("secret123"),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "secret123")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	if !called {
		t.Fatal("handler shoud have been called")
	}
}

func TestRequestID_HeaderAdded(t *testing.T) {
	var called bool

	handler := Chain(
		testHandler(&called),
		RequestID,
		Logging,
		Auth("secret123"),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-API-Key", "secret123")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	id := rr.Header().Get("X-Request-ID")
	if id == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}
}

func TestLogging_WritesLog_OnUnauthorized(t *testing.T) {
	var called bool

	logBuf, restore := captureLogs(t)
	defer restore()

	handler := Chain(
		testHandler(&called),
		RequestID,
		Logging,
		Auth("secret123"),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	logs := logBuf.String()

	if logs == "" {
		t.Fatal("expected logs to be written")
	}

	if !strings.Contains(logs, "status=401") {
		t.Fatalf("expected log to contain status=401, got: %s", logs)
	}
}
