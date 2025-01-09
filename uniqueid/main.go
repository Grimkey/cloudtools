package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	Mask5Bits  = 0x1F
	Mask12Bits = 0xFFF
)

type UniqueID uint64

func (id UniqueID) Epoch() uint64 {
	return uint64(id >> 22)
}

func (id UniqueID) Server() uint32 {
	return uint32(id >> 17 & Mask5Bits)
}

func (id UniqueID) Machine() uint32 {
	return uint32(id >> 12 & Mask5Bits)
}

func (id UniqueID) Increment() uint32 {
	return uint32(id & Mask12Bits)
}

func (id UniqueID) String() string {
	return fmt.Sprintf("%d (%b)", id, id)
}

type UniqueIDGen struct {
	currentID UniqueID
	epochMS   uint64
	server    uint32
	machine   uint32
	inc       uint32
}

func NewID(server int, machine int) (UniqueIDGen, error) {
	var epochMilliseconds uint64 = uint64(time.Now().UnixMilli())

	s := uint32(server)
	if s != (s & Mask5Bits) {
		return UniqueIDGen{}, fmt.Errorf("server must be between 0 and 31")
	}

	m := uint32(machine)
	if m != (m & Mask5Bits) {
		return UniqueIDGen{}, fmt.Errorf("machine must be between 0 and 31")
	}

	return UniqueIDGen{epochMS: epochMilliseconds, server: s, machine: m, inc: 0}, nil
}

func (id *UniqueIDGen) Next() uint64 {
	// Always collect this first to make sure time is consistent.
	now := uint64(time.Now().UnixMilli())

	// It is important to increment before we check the time, otherwise we could get into a race condition
	// where the time has been changed and the inc is reset, and we will reissue a previous ID
	var inc uint32 = atomic.AddUint32(&id.inc, 1)

	currentEpoch := atomic.LoadUint64(&id.epochMS)

	// Go can create A LOT of IDs in a millisecond, so to guarantee uniqueness we create ids in the future if we run out
	for {
		if now > currentEpoch {
			// Try to update the epochMS to the new timestamp
			if atomic.CompareAndSwapUint64(&id.epochMS, currentEpoch, now) {
				// Successfully swapped, reset inc to 0
				atomic.StoreUint32(&id.inc, 0)
				inc = 0
			} else {
				inc = atomic.AddUint32(&id.inc, 1)
			}
		}

		if inc <= Mask12Bits {
			break
		} else {
			now += 1
		}
	}

	// Construct the unique ID
	return (atomic.LoadUint64(&id.epochMS) << 22) | (uint64(id.server) << 17) | (uint64(id.machine) << 12) | uint64(inc)
}

func main() {
	id, err := NewID(1, 1)
	if err != nil {
		fmt.Println(err)
		return
	}

	numIterations := 8191
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

	// Collect and print results
	for newID := range results {
		fmt.Printf("%d (%b)\n", newID, newID)
	}
}
