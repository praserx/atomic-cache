package atomiccache

import (
	"sync"
)

// Shard structure contains multiple slots for records.
type Shard struct {
	sync.RWMutex
	slotAvail []int
	slots     []*Record
}

// NewShard initialize list of records with specified size. List is stored
// in property records and every record has it's own unique id (id is not
// propagated to record instance). Argument slotCount represents number of
// records in shard and slotSize represents size of one record.
func NewShard(slotCount, slotSize int) *Shard {
	shard := &Shard{}

	// Initialize available slots stack
	for i := 0; i < slotCount; i++ {
		shard.slotAvail = append(shard.slotAvail, i)
	}

	// Initialize record list
	for i := 0; i < slotCount; i++ {
		shard.slots = append(shard.slots, NewRecord(slotSize))
	}

	return shard
}

// Set store data as a record and decrease slotAvail count. On output it return
// index of used slot.
func (s *Shard) Set(data []byte) (i int) {
	s.Lock() // Lock for writing and reading
	i, s.slotAvail = s.slotAvail[0], s.slotAvail[1:]
	s.slots[i].Set(data)
	s.Unlock() // Unlock for writing and reading
	return
}

// Seti updates data in shard memory based on index. To preserve performance,
// it does not check if index is valid. It is responsibility of caller to ensure
// that index is valid and within bounds of shard.
func (s *Shard) Seti(i int, data []byte) {
	s.Lock() // Lock for writing and reading
	s.slots[i].Set(data)
	s.Unlock() // Unlock for writing and reading
}

// Get returns bytes from shard memory based on index. If array on output is
// empty, then record is not exists.
func (s *Shard) Get(index int) (v []byte) {
	s.RLock()
	v = s.slots[index].Get()
	s.RUnlock()
	return
}

// Free empty memory specified by index on input and increase slot counter.
func (s *Shard) Free(index int) {
	s.Lock()
	s.slots[index].Free()
	s.slotAvail = append(s.slotAvail, index)
	s.Unlock()
}

// GetSlotsAvail returns number of available memory slots of shard.
func (s *Shard) GetSlotsAvail() (cnt int) {
	s.RLock()
	cnt = len(s.slotAvail)
	s.RUnlock()
	return
}

// IsEmpty return true if shard has no record registered. Otherwise return
// false.
func (s *Shard) IsEmpty() (result bool) {
	s.RLock()
	if len(s.slotAvail) == len(s.slots) {
		result = true
	}
	s.RUnlock()
	return
}
