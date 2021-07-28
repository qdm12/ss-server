package filter

import (
	"hash/fnv"
	"sync"

	"github.com/riobard/go-bloom"
)

const (
	saltFilterCapacity          = 1e6
	saltFilterFalsePositiveRate = 1e-6
	saltFilterSlotsNumber       = 10
)

var _ SaltFilter = (*BloomRing)(nil)

// SaltFilter is used to mitigate replay attacks by detecting repeated salts.
type SaltFilter interface {
	AddSalt(b []byte)
	IsSaltRepeated(b []byte) bool
}

func NewBloomRing() *BloomRing {
	br := &BloomRing{
		slotCapacity: saltFilterCapacity / saltFilterSlotsNumber,
		slotCount:    saltFilterSlotsNumber,
		slots:        make([]bloom.Filter, saltFilterSlotsNumber),
	}
	for i := 0; i < saltFilterSlotsNumber; i++ {
		br.slots[i] = bloom.New(br.slotCapacity, saltFilterFalsePositiveRate, doubleFNV)
	}
	return br
}

// Double FNV as the Bloom Filter hash.
func doubleFNV(b []byte) (uint64, uint64) {
	hx := fnv.New64()
	_, _ = hx.Write(b)
	x := hx.Sum64()
	hy := fnv.New64a()
	_, _ = hy.Write(b)
	y := hy.Sum64()
	return x, y
}

type BloomRing struct {
	slotCapacity int
	slotPosition int
	slotCount    int
	entryCounter int
	slots        []bloom.Filter
	mu           sync.RWMutex
}

func (r *BloomRing) AddSalt(salt []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	slot := r.slots[r.slotPosition]
	if r.entryCounter > r.slotCapacity {
		// Move to next slot and reset
		r.slotPosition = (r.slotPosition + 1) % r.slotCount
		slot = r.slots[r.slotPosition]
		slot.Reset()
		r.entryCounter = 0
	}
	r.entryCounter++
	slot.Add(salt)
}

func (r *BloomRing) IsSaltRepeated(salt []byte) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.slots {
		if s.Test(salt) {
			return true
		}
	}
	return false
}
