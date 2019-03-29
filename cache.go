package atomiccache

import (
	"errors"
	"math/rand"
	"sync"
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
	// Lookup table with hashed values as key and LookupRecord as conent.
	lookup sync.Map

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
}

// LookupRecord ...
type LookupRecord struct {
	RecordIndex uint32
	ShardIndex  uint32
	Expiration  time.Duration
}

// New ...
func New(opts ...Option) *AtomicCache {
	var options = &Options{
		RecordSize: 4096,
		MaxRecords: 4096,
		MaxShards:  128,
	}

	for _, opt := range opts {
		opt(options)
	}

	// Init cache structure
	cache := &AtomicCache{}

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

	// Setup seed for random number generation
	rand.Seed(time.Now().UnixNano())

	return cache
}

// Set ...
func (a *AtomicCache) Set(key []byte, data []byte, expire time.Duration) error {
	if val, ok := a.lookup.Load(xxhash.Sum64(key)); ok == true {
		a.RLock()
		a.shards[val.(LookupRecord).ShardIndex].Free(val.(LookupRecord).RecordIndex)
		a.shards[val.(LookupRecord).ShardIndex].Set(data)
		a.RUnlock()
	} else {
		a.Lock()
		if si, ok := a.getShard(); ok == true {
			ri := a.shards[si].Set(data)
			a.lookup.Store(xxhash.Sum64(key), LookupRecord{ShardIndex: si, RecordIndex: ri, Expiration: expire})
		} else if si, ok := a.getEmptyShard(); ok == true {
			a.shards[si] = NewShard(a.MaxRecords, a.RecordSize)
			ri := a.shards[si].Set(data)
			a.lookup.Store(xxhash.Sum64(key), LookupRecord{ShardIndex: si, RecordIndex: ri, Expiration: expire})
		} else {
			si := uint32(rand.Intn(int(a.MaxShards)))
			ri := uint32(rand.Intn(int(a.MaxRecords)))
			a.shards[si].Free(ri)
			a.shards[si].Set(data)
			a.lookup.Store(xxhash.Sum64(key), LookupRecord{ShardIndex: si, RecordIndex: ri, Expiration: expire})
		}
		a.Unlock()
	}

	return nil
}

// Get returns list of bytes if record is present in cache memory. If record is
// not found, then error is returned and list is nil.
func (a *AtomicCache) Get(key []byte) ([]byte, error) {
	var result []byte
	var notfound = true

	if val, ok := a.lookup.Load(xxhash.Sum64(key)); ok == true {
		a.RLock()
		if a.shards[val.(LookupRecord).ShardIndex] != nil {
			result = a.shards[val.(LookupRecord).ShardIndex].Get(val.(LookupRecord).RecordIndex)
			notfound = false
		}
		a.RUnlock()

		if !notfound {
			return result, nil
		}
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
