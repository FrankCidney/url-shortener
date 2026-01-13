package shorten

import (
	"fmt"
	"testing"
)

func TestBase62Generator_Next(t *testing.T) {
	generator := NewBase62Generator()

	const n = 1000
	seen := make(map[string]bool, n)
	
	for i := 0; i < n; i++ {
		id, err := generator.Next("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if id == "" {
			t.Fatal("expected non-empty id")
		}
		seen[id] = true
	}
}

// Test: Same input -> Same output
func TestHashGenerator_Deterministic(t *testing.T) {
	generator := NewHashGenerator(8)

	url := "https://example.com"

	id1, err := generator.Next(url)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	id2, err := generator.Next(url)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id1 != id2 {
		t.Fatalf("expected same id for same input, got %s and %s", id1, id2)
	}
}

// Test: Different inputs -> Usually different outputs
func TestHashGenerator_DifferentInputs(t *testing.T) {
	generator := NewHashGenerator(8)

	url1 := "https://example.com"
	url2 := "https://example.com/other"

	id1, err := generator.Next(url1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	id2, err := generator.Next(url2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	} 

	if id1 == id2 {
		t.Fatalf("unexpected collision: %s", id1)
	}
}

// Test: Truncation length respected
func TestHashGenerator_Length(t *testing.T) {
	tests := []int{6, 8, 10}

	url := "https://example.com"

	for _, length := range tests {
		t.Run(fmt.Sprintf("length=%d", length), func(t *testing.T) {
			generator := NewHashGenerator(length)

			id, err := generator.Next(url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(id) != length {
				t.Fatalf("expected id length %d, got %d", length, len(id))
			}
		})
	}
}