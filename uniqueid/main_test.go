package main

import (
	"sync"
	"testing"
)

func TestUniqueID_NoDuplicates(t *testing.T) {
	id, err := NewID(1, 1)
	if err != nil {
		t.Fatal(err)
	}
	numIterations := 8190
	numGoroutines := 10 // Number of goroutines to use
	iterationsPerGoroutine := numIterations / numGoroutines

	// Channel to collect results
	results := make(chan uint64, numIterations)

	// Use a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Start goroutines
	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterationsPerGoroutine; i++ {
				newID := id.Next()
				results <- newID
			}
		}()
	}

	// Close the channel after all goroutines finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results into a slice
	var ids []uint64
	for newID := range results {
		ids = append(ids, newID)
	}

	// Check for duplicates
	seen := make(map[uint64]bool)
	for _, id := range ids {
		if seen[id] {
			t.Fatalf("Duplicate ID detected: %d (%b)", id, id)
		}
		seen[id] = true
	}

	// Ensure we got the expected number of IDs
	if len(ids) != numIterations {
		t.Fatalf("Expected %d IDs, but got %d", numIterations, len(ids))
	}
}
