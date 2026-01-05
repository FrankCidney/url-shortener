package shorten

import "sync"

type Store interface {
	Save(id string, url string) error
	Get(id string) (string, bool)
	List() map[string]string
}

type MemStore struct {
	mu sync.RWMutex
	data map[string]string
}

func NewMemStore() *MemStore {
	return &MemStore{
		data: make(map[string]string),
	}
}

func (store *MemStore) Save(id string, url string) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	
	store.data[id] = url

	return nil
}

func (store *MemStore) Get(id string) (string, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	value, exists := store.data[id]
	return value, exists
}

func (store *MemStore) List() map[string]string {
	store.mu.RLock()
	defer store.mu.RUnlock()

	// return a snapshot, not the internal map
	copy := make(map[string]string, len(store.data))

	for k, v := range store.data {
		copy[k] = v
	}

	return copy
}