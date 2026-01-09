package shorten

import "sync"

// type Store interface {
// 	Save(id string, url string) error
// 	Get(id string) (string, bool)
// 	List() map[string]string
// }

type Store interface {
	Save(link ShortLink) error
	Get(id string) (ShortLink, bool)
	List() []ShortLink
	IncrementHits(id string)
}

type MemStore struct {
	mu sync.RWMutex
	data map[string]ShortLink
}

func NewMemStore() *MemStore {
	return &MemStore{
		data: make(map[string]ShortLink),
	}
}

func (store *MemStore) Save(link ShortLink) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	
	store.data[link.ID] = link

	return nil
}

func (store *MemStore) Get(id string) (ShortLink, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	value, exists := store.data[id]
	return value, exists
}

func (store *MemStore) List() []ShortLink {
	store.mu.RLock()
	defer store.mu.RUnlock()

	// return a snapshot, not the internal map
	// copy := make(map[string]ShortLink, len(store.data))
	copy := make([]ShortLink, len(store.data))

	for _, v := range store.data {
		copy = append(copy, v)
	}

	return copy
}

func (store *MemStore) IncrementHits(id string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	link, ok := store.data[id]
	if !ok {
		return
	}

	link.Hits++
	store.data[id] = link
}