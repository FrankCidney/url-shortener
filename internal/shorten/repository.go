package shorten

import "errors"

var (
	ErrDuplicateID = errors.New("duplicate short id")
	ErrNotFound = errors.New("link not found")
)

type Store interface {
	Save(link ShortLink) error
	Get(id string) (ShortLink, error)
	// List() []ShortLink
	IncrementHits(id string) error
}
