package shorten

type Store interface {
	Save(link ShortLink) error
	Get(id string) (ShortLink, error)
	// List() []ShortLink
	IncrementHits(id string) error
}
