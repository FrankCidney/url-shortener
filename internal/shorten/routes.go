package shorten

import "net/http"

func RegisterRoutes(mux *http.ServeMux, shortener *Shortener) {
	handler := NewHandler(shortener)

	mux.HandleFunc("POST /shorten", handler.HandleShorten)
	mux.HandleFunc("GET /stats/", handler.HandleStats)
	mux.HandleFunc("GET /", handler.HandleRedirect)
}