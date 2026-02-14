package shorten

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

// todo: Find out: Should these (the error definitions below) be moved to a separate file specifically for errors? Is that better design than having them here?
var ErrNotFound = errors.New("not found")
var ErrInvalidURL = errors.New("invalid url")
var ErrTooManyCollisions = errors.New("could not generate unique id, too many collisions")

type Shortener struct {
	store Store
	ids   IDGenerator
}

func NewShortener(store Store, ids IDGenerator) *Shortener {
	return &Shortener{
		store: store,
		ids:   ids,
	}
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

// Create generates a Short ID and saves it along with the associated URL
// It also initialises a hit counter and saves the time of creation (CreatedAt)
func (s *Shortener) Create(url string) (ShortLink, error) {
	// Validate the URL
	if err := validateURL(url); err != nil {
		return ShortLink{}, fmt.Errorf("%w: %s", ErrInvalidURL, err)
	}

	const maxAttempts = 10

	for attempt := 0; attempt < maxAttempts; attempt++ {
		input := url
		if attempt > 0 {
			input = fmt.Sprintf("%s#%d", url, attempt)
		}

		id, err := s.ids.Next(input)
		if err != nil {
			return ShortLink{}, err
		}

		// // check collision
		// if _, exists := s.store.Get(id); exists {
		// 	// collision -> try again
		// 	continue
		// }

		link := ShortLink{
			ID:        id,
			URL:       url,
			Hits:      0,
			CreatedAt: time.Now(),
		}

		if err := s.store.Save(link); err != nil {
			if errors.Is(err, ErrDuplicateID) {
				continue // collision -> retry
			}
			return ShortLink{}, err
		}

		return link, nil
	}

	return ShortLink{}, ErrTooManyCollisions 
}

// Resolve returns the URL associated with the given id. It also increments hits
func (s *Shortener) Resolve(id string) (string, error) {
	link, ok := s.store.Get(id)
	if !ok {
		return "", ErrNotFound
	}

	s.store.IncrementHits(id)

	return link.URL, nil
}

// Stats returns metadata for an ID, including the Short ID itself, the associated URL, hit count, and time of creation of the ID
func (s *Shortener) Stats(id string) (ShortLink, error) {
	link, ok := s.store.Get(id)
	if !ok {
		return ShortLink{}, ErrNotFound
	}

	return link, nil
}
