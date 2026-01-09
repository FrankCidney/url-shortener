package shorten

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxBodyBytes = 1 << 20 // 1MB

type Handler struct {
	service *Shortener
}

func NewHandler(s *Shortener) *Handler {
	return &Handler{service: s}
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Short string `json:"short"`
	URL   string `json:"url"`
}

type apiError struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, apiError{Error: msg})
}

func validateURL(raw string) error {
	// make sure URL is not empty
	if raw == "" {
		return errors.New("url is required")
	}

	// parseable i.e., no syntax errors
	u, err := url.Parse(raw)
	if err != nil {
		return errors.New("invalid url")
	}

	// url.Parse allows relative paths; so here we require scheme and host, + reject unwanted schemes
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("url scheme must be http or https")
	}
	if u.Host == "" {
		return errors.New("url missing host")
	}
	return nil
}

func (h *Handler) HandleShorten(w http.ResponseWriter, r *http.Request) {
	// 1) method
	// This one's basically redundant since Go already blocks other methods when you indicate the method in your path string when
	// registering a handler. 
	// But I'll leave it in for now in case say someone forgets and registers a handler without indicating it should only accept POST requests.
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 2) limit body size to prevent DoS attacks
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)

	// 3) disallow unknown fields
	dec.DisallowUnknownFields()

	var req shortenRequest
	if err := dec.Decode(&req); err != nil {
		var syntaxErr *json.SyntaxError
		var unmarshalTypeErr *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxErr):
			writeError(w, http.StatusBadRequest, "malformed JSON")
		case errors.Is(err, io.ErrUnexpectedEOF):
			writeError(w, http.StatusBadRequest, "malformed JSON")
		case errors.As(err, &unmarshalTypeErr):
			writeError(w, http.StatusBadRequest, "invalid JSON field type")
		case err.Error() == "http: request body too large":
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
		default:
			// includes errors from DisallowUnknownFields()
			writeError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	// 4) make sure there's no extra JSON after the first object
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "multiple JSON values in body")
		return
	}

	// 5) validate the URL
	if err := validateURL(req.URL); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// 6) generate short code
	link, err := h.service.Create(req.URL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := shortenResponse{
		Short: link.ID,
		URL:   link.URL,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	url, err := h.service.Resolve(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "short link not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
	}

	http.Redirect(w, r, url, http.StatusFound) // 302 redirect
}

func (h *Handler) HandleStats(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/stats/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "misssing id")
		return
	}

	link, err := h.service.Stats(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "short link not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := struct {	
		URL       string `json:"url"`
		Short        string `json:"short"`
		Hits      int64 `json:"hits"`
		CreatedAt string `json:"createdAt"`
	} {
		URL: link.URL,
		Short: link.ID,
		Hits: link.Hits,
		CreatedAt: link.CreatedAt.Format(time.RFC3339),
	}

	writeJSON(w, http.StatusOK, resp)
}