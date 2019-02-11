package atomiccache

import (
	"sync"
	"time"
)

// Shard structure contains multiple slots for records.
type Shard struct {
	sync.RWMutex
	slotSize  uint32
	slotCount uint32
	slotAvail []uint32
	slots     []*Record
	tasks     sync.Map
}

// NewShard initialize list of records with specified size. List is stored
// in property records and every record has it's own unique id (id is not
// propagated to record instance).
func NewShard(slotCount, slotSize uint32) *Shard {
	shard := &Shard{
		slotSize:  slotSize,
		slotCount: slotCount,
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

// Set store data as a record and decrease slotAvail count. On output it return
// index of used slot.
func (s *Shard) Set(data []byte, expire time.Duration) uint32 {
	var index uint32

	s.Lock() // Lock for writing and reading
	index, s.slotAvail = s.slotAvail[0], s.slotAvail[1:]
	s.Unlock() // Unlock for writing and reading

	// Set data
	s.slots[index].Set(data)

	// Run expiration task
	go s.FreeAfterExpiration(index, expire)

	return index
}

// Get returns bytes from shard memory based on index. If array on output is
// empty, then record is not exists.
func (s *Shard) Get(index uint32) []byte {
	return s.slots[index].Get()
}

// Free empty memory specified by index on input and increase slot counter.
func (s *Shard) Free(index uint32) {
	s.slots[index].Free()

	s.Lock()
	s.slotAvail = append(s.slotAvail, index)
	s.Unlock()
}

// FreeAfterExpiration frees memory after destinated time. It requires index and
// expiration time on input.
func (s *Shard) FreeAfterExpiration(index uint32, expire time.Duration) {
	currentTime := time.Now()

	s.tasks.Store(currentTime, true)

	timer := time.NewTimer(expire)
	<-timer.C

	s.slots[index].Free()

	s.Lock()
	s.slotAvail = append(s.slotAvail, index)
	s.Unlock()

	s.tasks.Delete(currentTime)
}
