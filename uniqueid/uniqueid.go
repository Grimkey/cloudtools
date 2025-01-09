package uniqueid

import (
	"fmt"
	"sync/atomic"
	"time"
)

const (
	Mask5Bits  = 0x1F
	Mask12Bits = 0xFFF
)

func MaxUint64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

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
	currentID uint64
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

	currentID := makeID(epochMilliseconds, s, m, 0)

	return UniqueIDGen{currentID: currentID}, nil
}

func makeID(epoch uint64, server uint32, machine uint32, inc uint32) uint64 {
	return (epoch << 22) | (uint64(server) << 17) | (uint64(machine) << 12) | uint64(inc)
}

func (id *UniqueIDGen) Next() UniqueID {
	for {
		// Always collect this first to make sure time is consistent.
		now := uint64(time.Now().UnixMilli())
		currentID := UniqueID(atomic.LoadUint64(&id.currentID))

		var newID uint64 = 0

		if now > currentID.Epoch() {
			newID = makeID(now, currentID.Server(), currentID.Machine(), 0)
			if atomic.CompareAndSwapUint64(&id.currentID, uint64(currentID), newID) {
				return UniqueID(newID)
			}
		}

		inc := currentID.Increment() + 1

		// Corner case where this machine has done 4096 its within the same millisecond
		// Without the sleep, I could not figure out how to prevent race conditions
		// just by incrementing the "now" value does not work.
		if inc >= Mask12Bits {
			time.Sleep(time.Millisecond)
			continue
		}

		newID = makeID(MaxUint64(now, currentID.Epoch()), currentID.Server(), currentID.Machine(), inc)

		if atomic.CompareAndSwapUint64(&id.currentID, uint64(currentID), newID) {
			return UniqueID(newID)
		}
	}
}
