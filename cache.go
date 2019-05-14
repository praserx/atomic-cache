package atomiccache

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cespare/xxhash"
)

// Internal cache errors
var (
	ErrNotFound  = errors.New("Record not found")
	ErrDataLimit = errors.New("Can't create new record, it violates data limit")
)

// AtomicCache structure represents whole cache memory.
type AtomicCache struct {
	// RWMutex is used for access to shards array.
	sync.RWMutex
	// // Lookup table with hashed values as key and LookupRecord as conent.
	// lookup sync.Map
	// Lookup
	lookup map[uint64]LookupRecord

	// Array of pointers to shard objects.
	shards []*Shard
	// Array of shard indexes which are currently active.
	shardsActive []uint32
	// Array of shard indexes which are currently available for new allocation.
	shardsAvail []uint32

	// Size of byte array used for memory allocation.
	RecordSize uint32
	// Maximum records per shard.
	MaxRecords uint32
	// Maximum shards for allocation.
	MaxShards uint32

	// Garbage collector starter (run garbage collection every X memory sets).
	GcStarter uint32
	// Garbage collector counter for starter.
	GcCounter uint32
}

// LookupRecord represents item in lookup table. One record contains index of
// shard and record. So we can determine which shard access and which record of
// shard to get. Record also contains expiration time.
type LookupRecord struct {
	RecordIndex uint32
	ShardIndex  uint32
	Expiration  time.Time
}

// New initialize whole cache memory with one allocated shard.
func New(opts ...Option) *AtomicCache {
	var options = &Options{
		RecordSize: 4096,
		MaxRecords: 4096,
		MaxShards:  128,
		GcStarter:  5000,
	}

	for _, opt := range opts {
		opt(options)
	}

	// Init cache structure
	cache := &AtomicCache{}

	// Init lookup table
	cache.lookup = make(map[uint64]LookupRecord)

	// Init shards list with nil values
	cache.shards = make([]*Shard, options.MaxShards, options.MaxShards)

	// Create shard available indexes
	for i := uint32(0); i < options.MaxShards; i++ {
		cache.shardsAvail = append(cache.shardsAvail, i)
	}

	// Create start shard
	var shardIndex uint32
	shardIndex, cache.shardsAvail = cache.shardsAvail[0], cache.shardsAvail[1:]
	cache.shardsActive = append(cache.shardsActive, shardIndex)
	cache.shards[shardIndex] = NewShard(options.MaxRecords, options.RecordSize)

	// Define setup values
	cache.RecordSize = options.RecordSize
	cache.MaxRecords = options.MaxRecords
	cache.MaxShards = options.MaxShards
	cache.GcStarter = options.GcStarter

	return cache
}

// Set store data to cache memory. If key/record is already in memory, then data
// are replaced. If not, it checks if there are some allocated shard with empty
// space for data. If there is no empty space, new shard is allocated. Otherwise
// some valid record (FIFO queue) is deleted and new one is stored.
func (a *AtomicCache) Set(key []byte, data []byte, expire time.Duration) error {
	if len(data) > int(a.RecordSize) {
		return ErrDataLimit
	}

	hash := xxhash.Sum64(key)

	a.Lock()
	if val, ok := a.lookup[hash]; ok {
		a.shards[val.ShardIndex].Free(val.RecordIndex)
		val.RecordIndex = a.shards[val.ShardIndex].Set(data)
		a.lookup[hash] = LookupRecord{ShardIndex: val.ShardIndex, RecordIndex: val.RecordIndex, Expiration: a.getExprTime(expire)}
	} else {
		if si, ok := a.getShard(); ok {
			ri := a.shards[si].Set(data)
			a.lookup[hash] = LookupRecord{ShardIndex: si, RecordIndex: ri, Expiration: a.getExprTime(expire)}
		} else if si, ok := a.getEmptyShard(); ok {
			a.shards[si] = NewShard(a.MaxRecords, a.RecordSize)
			ri := a.shards[si].Set(data)
			a.lookup[hash] = LookupRecord{ShardIndex: si, RecordIndex: ri, Expiration: a.getExprTime(expire)}
		} else {
			for k, v := range a.lookup {
				delete(a.lookup, k)
				a.shards[v.ShardIndex].Free(v.RecordIndex)
				v.RecordIndex = a.shards[v.ShardIndex].Set(data)
				a.lookup[hash] = LookupRecord{ShardIndex: v.ShardIndex, RecordIndex: v.RecordIndex, Expiration: a.getExprTime(expire)}
				break
			}
		}
	}
	a.Unlock()

	if atomic.AddUint32(&a.GcCounter, 1) == a.GcStarter {
		atomic.StoreUint32(&a.GcCounter, 0)
		go a.collectGarbage()
	}

	return nil
}

// Get returns list of bytes if record is present in cache memory. If record is
// not found, then error is returned and list is nil.
func (a *AtomicCache) Get(key []byte) ([]byte, error) {
	var result []byte
	var hit = false

	a.RLock()
	if val, ok := a.lookup[xxhash.Sum64(key)]; ok {
		if a.shards[val.ShardIndex] != nil && time.Now().Before(val.Expiration) {
			result = a.shards[val.ShardIndex].Get(val.RecordIndex)
			hit = true
		}
	}
	a.RUnlock()

	if hit {
		return result, nil
	}

	return nil, ErrNotFound
}

// releaseShard release shard if there is no record in memory. It returns true
// if shard was released.
// This method is not thread safe and additional locks are required.
func (a *AtomicCache) releaseShard(shard uint32) bool {
	if a.shards[shard].IsEmpty() == true {
		a.shards[shard] = nil
		return true
	}

	return false
}

// getShard return index of shard which have some available space for new
// record. If there is no shard with available space, then false is returned as
// a second value.
// This method is not thread safe and additional locks are required.
func (a *AtomicCache) getShard() (uint32, bool) {
	for _, shardIndex := range a.shardsActive {
		if a.shards[shardIndex].GetSlotsAvail() != 0 {
			return shardIndex, true
		}
	}

	return 0, false
}

// getEmptyShard return index of shard that can be used for new shard
// allocation. If there is no left index, then false is returned as a second
// value.
// This method is not thread safe and additional locks are required.
func (a *AtomicCache) getEmptyShard() (uint32, bool) {
	if len(a.shardsAvail) == 0 {
		return 0, false
	}

	var shardIndex uint32
	shardIndex, a.shardsAvail = a.shardsAvail[0], a.shardsAvail[1:]

	return shardIndex, true
}

// getExprTime return expiration time based on duration. If duration is 0, then
// maximum expiration time is used (48 hours).
func (a *AtomicCache) getExprTime(expire time.Duration) time.Time {
	if expire == 0 {
		return time.Now().Add(48 * time.Hour)
	}

	return time.Now().Add(expire)
}

// collectGarbage provides garbage collect. It goes throught lookup table and
// checks expiration time. If shard end up empty, then garbage collect release
// him, but only if there is more than one shard in charge (we always have one
// active shard).
func (a *AtomicCache) collectGarbage() {
	a.Lock()
	for k, v := range a.lookup {
		if time.Now().After(v.Expiration) {
			a.shards[v.ShardIndex].Free(v.RecordIndex)
			if len(a.shardsActive) > 1 {
				a.releaseShard(v.ShardIndex)
			}
			delete(a.lookup, k)
		}
	}
	a.Unlock()
}
