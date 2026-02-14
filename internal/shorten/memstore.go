package shorten

import "sync"

type MemStore struct {
	mu   sync.RWMutex
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

	if _, exists := store.data[link.ID]; exists {
		return ErrDuplicateID
	}
	store.data[link.ID] = link

	return nil
}

func (store *MemStore) Get(id string) (ShortLink, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	link, exists := store.data[id]
	if !exists {
		return ShortLink{}, ErrNotFound
	}

	return link, nil
}

// func (store *MemStore) List() []ShortLink {
// 	store.mu.RLock()
// 	defer store.mu.RUnlock()

// 	// return a snapshot, not the internal map
// 	copy := make([]ShortLink, 0, len(store.data))

// 	for _, v := range store.data {
// 		copy = append(copy, v)
// 	}

// 	return copy
// }

func (store *MemStore) IncrementHits(id string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	link, ok := store.data[id]
	if !ok {
		return ErrNotFound
	}

	link.Hits++
	store.data[id] = link
	return nil
}
