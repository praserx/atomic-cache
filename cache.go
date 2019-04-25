package atomiccache

import (
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cespare/xxhash"
)

// Internal cache errors
var (
	errNotFound  = errors.New("Record not found")
	errSetRecord = errors.New("Can't create new record, hit fail is too high")
)

// AtomicCache ...
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

// LookupRecord ...
type LookupRecord struct {
	RecordIndex uint32
	ShardIndex  uint32
	Expiration  time.Time
}

// New ...
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

	// Define setup values
	cache.RecordSize = options.RecordSize
	cache.MaxRecords = options.MaxRecords
	cache.MaxShards = options.MaxShards
	cache.GcStarter = options.GcStarter

	// Setup seed for random number generation
	rand.Seed(time.Now().UnixNano())

	return cache
}

// Set ...
func (a *AtomicCache) Set(key []byte, data []byte, expire time.Duration) error {
	a.Lock()
	if val, ok := a.lookup[xxhash.Sum64(key)]; ok {
		a.shards[val.ShardIndex].Free(val.RecordIndex)
		a.shards[val.ShardIndex].Set(data)
	} else {
		if si, ok := a.getShard(); ok == true {
			ri := a.shards[si].Set(data)
			a.lookup[xxhash.Sum64(key)] = LookupRecord{ShardIndex: si, RecordIndex: ri, Expiration: a.getExprTime(expire)}
		} else if si, ok := a.getEmptyShard(); ok == true {
			a.shards[si] = NewShard(a.MaxRecords, a.RecordSize)
			ri := a.shards[si].Set(data)
			a.lookup[xxhash.Sum64(key)] = LookupRecord{ShardIndex: si, RecordIndex: ri, Expiration: a.getExprTime(expire)}
		} else {
			for k, v := range a.lookup {
				delete(a.lookup, k)
				a.shards[v.ShardIndex].Free(v.RecordIndex)
				v.RecordIndex = a.shards[v.ShardIndex].Set(data)
				a.lookup[xxhash.Sum64(key)] = LookupRecord{ShardIndex: v.ShardIndex, RecordIndex: v.RecordIndex, Expiration: a.getExprTime(expire)}
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
		if a.shards[val.ShardIndex] != nil { //&& time.Now().Before(val.Expiration) {
			result = a.shards[val.ShardIndex].Get(val.RecordIndex)
			hit = true
		}
	}
	a.RUnlock()

	if hit {
		return result, nil
	}

	return nil, errNotFound
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

// collectGarbage ...
func (a *AtomicCache) collectGarbage() {
	a.Lock()
	for k, v := range a.lookup {
		if time.Now().After(v.Expiration) {
			a.shards[v.ShardIndex].Free(v.RecordIndex)
			a.releaseShard(v.ShardIndex)
			delete(a.lookup, k)
		}
	}
	a.Unlock()
}
