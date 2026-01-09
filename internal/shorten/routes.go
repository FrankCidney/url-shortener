package shorten

import "net/http"

func RegisterRoutes(mux *http.ServeMux, shortener *Shortener) {
	handler := NewHandler(shortener)

	mux.HandleFunc("/shorten", handler.HandleShorten)
	mux.HandleFunc("/stats/", handler.HandleStats)
	mux.HandleFunc("/", handler.HandleRedirect)
}