package shorten

import (
	"errors"
	"fmt"
	"time"
)

var ErrNotFound = errors.New("not found")

type Shortener struct {
	store Store
	ids IDGenerator
}

func NewShortener(store Store, ids IDGenerator) *Shortener {
	return &Shortener{
		store: store,
		ids: ids,
	}
}

func (s *Shortener) Create(url string) (ShortLink, error) {
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

		// check collision
		if _, exists := s.store.Get(id); exists {
			// collision -> try again
			continue
		}

		link := ShortLink{
			ID: id,
			URL: url,
			CreatedAt: time.Now(),
			Hits: 0,
		}

		if err := s.store.Save(link); err != nil {
			return ShortLink{}, err
		}

		return link, nil
	}

	return ShortLink{}, errors.New("could not generate unique id")
}

// Resolve gets the URL associated with the given short id, and increments hits
func (s *Shortener) Resolve(id string) (string, error) {
	link, ok := s.store.Get(id)
	if !ok {
		return "", ErrNotFound
	}

	s.store.IncrementHits(id)

	return link.URL, nil
}

func (s *Shortener) Stats(id string) (ShortLink, error) {
	link, ok := s.store.Get(id)
	if !ok {
		return ShortLink{}, ErrNotFound
	}

	return link, nil
}
