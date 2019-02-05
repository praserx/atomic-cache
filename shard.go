package atomiccache

import (
	"sync"
	"time"
)

// Shard structure contains multiple slots for records.
type Shard struct {
	slotSize  uint32
	slotCount uint32
	slotAvail []uint32
	slots     []*Record
	mutex     *sync.RWMutex
}

// NewShard initialize list of records with specified size. List is stored
// in property records and every record has it's own unique id (id is not
// propagated to record instance).
func NewShard(slotCount, slotSize uint32) *Shard {
	shard := &Shard{
		slotSize:  slotSize,
		slotCount: slotCount,
		mutex:     &sync.RWMutex{},
	}

	// Initialize available slots stack
	for i := uint32(0); i < slotCount; i++ {
		shard.slotAvail = append(shard.slotAvail, i)
	}

	// Initialize record list
	for i := uint32(0); i < slotCount; i++ {
		shard.slots = append(shard.slots, NewRecord(slotSize))
	}

	return shard
}

// Set ...
func (s *Shard) Set(data []byte, expire time.Duration) uint32 {
	var index uint32

	s.mutex.Lock() // Lock for writing and reading
	index, s.slotAvail = s.slotAvail[0], s.slotAvail[1:]
	s.mutex.Unlock() // Unlock for writing and reading

	s.slots[index].Set(data)
	go s.FreeAfterExpiration(index, expire) // TODO: It is probably not possible.

	return index
}

// Get ...
func (s *Shard) Get(index uint32) []byte {
	return s.slots[index].Get()
}

// Free ...
func (s *Shard) Free(index uint32) {
	s.slots[index].Free()

	s.mutex.Lock()
	s.slotAvail = append(s.slotAvail, index)
	s.mutex.Unlock()
}

// FreeAfterExpiration ...
func (s *Shard) FreeAfterExpiration(index uint32, expire time.Duration) {
	timer := time.NewTimer(expire)
	<-timer.C

	s.slots[index].Free()

	s.mutex.Lock()
	s.slotAvail = append(s.slotAvail, index)
	s.mutex.Unlock()
}
