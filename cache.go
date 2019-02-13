package atomiccache

import (
	"errors"
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
	lookup LookupTable
	tasks  sync.Map

	shards       []*Shard
	shardsActive []uint32
	shardsAvail  []uint32
	shardsLocker sync.RWMutex

	settings Settings
}

// Settings ...
type Settings struct {
	// Size of byte array used for memory allocation.
	RecordSize uint32
	// Maximum records per shard.
	MaxRecords uint32
}

// SearchRecord ...
type SearchRecord struct {
	ShardIndex  uint32
	RecordIndex uint32
	ChronoIndex int64
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

	// Save cache settings
	cache.settings = Settings{
		RecordSize: options.RecordSize,
		MaxRecords: options.MaxRecords,
	}

	// Init shards list with nil values
	cache.shards = make([]*Shard, options.MaxShards, options.MaxShards)

	// Create shard available indexes
	for i := uint32(0); i < options.MaxShards; i++ {
		cache.shardsAvail = append(cache.shardsAvail, i)
	}

	return cache
}

// Set ...
func (a *AtomicCache) Set(key, data []byte, expire time.Duration) error {
	var shardAvail, shardAvailSpace bool

	hash := xxhash.Sum64(key)

	si, ri, status := a.lookup.Get(hash)
	if status == true {
		a.shardsLocker.RLock()   // Lock for reading
		a.shards[si].Free(ri)    // Empty old record
		a.shardsLocker.RUnlock() // Unlock for reading

		a.lookup.Delete(hash)
	} else {
		a.shardsLocker.RLock()
		si, shardAvailSpace = a.getShard()
		a.shardsLocker.RUnlock()

		if shardAvailSpace != true {
			a.shardsLocker.RLock()
			si, shardAvail = a.getShardAvail()
			a.shardsLocker.RUnlock()

			if shardAvail {
				a.shardsLocker.Lock() // Lock for writing and reading
				a.shards[si] = NewShard(a.settings.MaxRecords, a.settings.RecordSize)
				a.shardsLocker.Unlock() // Unlock for writing and reading
			} else {
				var hitFail int

				for {
					if hitFail >= 69 { // need some magic number
						return errSetRecord
					}

					si, ri, status = a.lookup.Pop()
					if status == true {
						break
					}

					hitFail++
				}

				a.shardsLocker.RLock() // Lock for reading
				if a.shards[si] != nil {
					a.shards[si].Free(ri) // Empty old record
				}
				a.shardsLocker.RUnlock() // Unlock for reading

				a.lookup.Delete(hash)
			}
		}
	}

	a.shardsLocker.RLock()      // Lock for reading
	ri = a.shards[si].Set(data) // Set data update
	a.shardsLocker.RUnlock()    // Unlock for reading

	a.lookup.Set(hash, si, ri, expire)     // Update lookup table
	go a.freeAfterExpiration(hash, expire) // Start expiration worker

	return nil
}

// Get returns list of bytes if record is present in cache memory. If record is
// not found, then error is returned and list is nil.
func (a *AtomicCache) Get(key []byte) ([]byte, error) {
	var result []byte
	var nf = true

	si, ri, status := a.lookup.Get(xxhash.Sum64(key))
	if status == true {
		a.shardsLocker.RLock()
		if a.shards[si] != nil {
			result = a.shards[si].Get(ri)
			nf = false
		}
		a.shardsLocker.RUnlock()

		if nf == false {
			return result, nil
		}
	}

	return nil, errNotFound
}

// freeAfterExpiration frees memory after destinated time. It requires index and
// expiration time on input.
func (a *AtomicCache) freeAfterExpiration(hashKey uint64, expire time.Duration) {
	chronoKey := time.Now().Add(expire).UnixNano()
	a.tasks.Store(chronoKey, hashKey)

	timer := time.NewTimer(expire)
	<-timer.C

	a.shardsLocker.Lock()
	si, ri, status := a.lookup.Get(hashKey)
	if status == true {
		a.shards[si].Free(ri)
		if a.releaseShard(si) {
			a.lookup.Delete(hashKey)
		}
	}
	a.shardsLocker.Unlock()

	a.tasks.Delete(chronoKey)
}

// releaseShard release shard if there is no record in memory. It returns true
// if shard was released. This method is not thread safe and additional locks
// are required.
func (a *AtomicCache) releaseShard(shard uint32) bool {
	if a.shards[shard].GetSlotsAvail() == a.settings.MaxRecords {
		a.shards[shard] = nil
		return true
	}
	return false
}

// getShard return index of shard which have some available space for new
// record. If there is no shard with available space, then false is returned as
// a second value. This method is not thread safe and additional locks are
// required.
func (a *AtomicCache) getShard() (uint32, bool) {
	for _, shardIndex := range a.shardsActive {
		if a.shards[shardIndex].GetSlotsAvail() != 0 {
			return shardIndex, true
		}
	}

	return 0, false
}

// getShardAvail return index of shard that can be used for new shard
// allocation. If there is no left index, then false is returned as a second
// value. This method is not thread safe and additional locks are required.
func (a *AtomicCache) getShardAvail() (uint32, bool) {
	if len(a.shardsAvail) == 0 {
		return 0, false
	}

	var shardIndex uint32
	shardIndex, a.shardsAvail = a.shardsAvail[0], a.shardsAvail[1:]

	return shardIndex, true
}
