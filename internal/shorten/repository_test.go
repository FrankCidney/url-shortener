package shorten
// package shorten

// import (
// 	"fmt"
// 	"sync"
// 	"testing"
// )

// func TestMemStore_SaveGetList(t *testing.T) {
// 	t.Run("save and get", func(t *testing.T) {
// 		store := NewMemStore()
// 		id := "abc123"
// 		url := "https://example.com"

// 		if err := store.Save(id, url); err != nil {
// 			t.Fatalf("unexpected error on save: %v", err)
// 		}

// 		got, ok := store.Get(id)
// 		if !ok {
// 			t.Fatal("expected id to exist")
// 		}
// 		if got != url {
// 			t.Fatalf("expected url %q, got %q", url, got)
// 		}
// 	})

// 	t.Run("list", func(t *testing.T) {
// 		store := NewMemStore()
// 		store.Save("a", "url1")
// 		store.Save("b", "url2")

// 		m := store.List()
// 		if len(m) != 2 {
// 			t.Fatalf("expected 2 items, got %d", len(m))
// 		}

// 		// Make sure modifying returned map doesn't affect internal state
// 		m["c"] = "url3"
// 		if _, ok := store.Get("c"); ok {
// 			t.Fatal("modifying List result changed store")
// 		}
// 	})
// }

// func TestMemStore_ConcurrentSaveGet(t *testing.T) {
// 	store := NewMemStore()
// 	wg := sync.WaitGroup{}

// 	n := 1000

// 	for i := 0; i < n; i++ {
// 		wg.Add(2)

// 		go func(i int) {
// 			defer wg.Done()
// 			id := fmt.Sprintf("id%d", i)
// 			store.Save(id, "url")
// 		}(i)

// 		go func(i int) {
// 			defer wg.Done()
// 			id := fmt.Sprintf("id%d", i)
// 			store.Get(id)
// 		}(i)
// 	}

// 	wg.Wait()
// }