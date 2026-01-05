package shared

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const apiKeyHeader = "X-API-Key"

// this is for creating unique request ID keys for request context,
// to prevent collisions with other middleware
type ctxKeyRequestID struct{}

// wrap responseWriter so we can record response status and response size
type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.status = code
	lw.ResponseWriter.WriteHeader(code)
}

func (lw *loggingResponseWriter) Write(b []byte) (int, error) {
	if lw.status == 0 {
		lw.status = http.StatusOK
	}
	n, err := lw.ResponseWriter.Write(b)
	lw.bytes += n
	return n, err
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lw := &loggingResponseWriter{
			ResponseWriter: w,
		}

		next.ServeHTTP(lw, r)

		duration := time.Since(start)

		reqID, _ := r.Context().Value(ctxKeyRequestID{}).(string)

		log.Printf(
			`request_id=%s method=%s path=%s status=%d bytes=%d duration=%s`,
			reqID,
			r.Method,
			r.URL.Path,
			lw.status,
			lw.bytes,
			duration,
		)
	})
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()

		ctx := context.WithValue(r.Context(), ctxKeyRequestID{}, id)
		r = r.WithContext(ctx)

		w.Header().Set("X-Request-ID", id)
		
		next.ServeHTTP(w, r)
	})
}

func Auth(expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get(apiKeyHeader)

			if key == "" || key != expectedKey {
				w.Header().Set("Content-Type", "appication/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "unauthorized",
				})
				return
			}

			// if key is valid, continue
			next.ServeHTTP(w, r)
		})
	}
}

func Chain(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}
