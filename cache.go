package atomiccache

import (
	"errors"
	"sync"
	"time"

	"github.com/cespare/xxhash"
)

// AtomicCache ...
type AtomicCache struct {
	lookup sync.Map
	tasks  sync.Map

	allocs      []uint64
	stackLocker sync.RWMutex

	shards       []*Shard
	shardsActive []uint32
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
	Shard uint32
	Index uint32
}

// New ...
func New(opts ...Option) *AtomicCache {
	var options = &Options{
		RecordSize: 2048,
		MaxRecords: 2048,
		MaxShards:  24,
	}

	for _, opt := range opts {
		opt(options)
	}

	cache := &AtomicCache{
		settings: Settings{
			RecordSize: options.RecordSize,
			MaxRecords: options.MaxRecords,
		},
		shards: make([]*Shard, options.MaxShards, options.MaxShards),
	}

	return cache
}

// Set ...
func (a *AtomicCache) Set(key, data []byte, expire time.Duration) {
	hash := xxhash.Sum64(key)

	value, status := a.lookup.Load(hash)
	if status == true {
		sre := value.(SearchRecord)

		a.shardsLocker.RLock()                    // Lock for reading
		a.shards[sre.Shard].Free(sre.Index)       // Empty old record
		sre.Index = a.shards[sre.Shard].Set(data) // Set data update
		a.shardsLocker.RUnlock()                  // Unlock for reading

		a.lookup.Store(hash, sre)              // Update lookup table
		go a.freeAfterExpiration(hash, expire) // Start expiration worker
	} else {
		// 1. Is there any free space? If not make some (remove oldest record)
		// 2. Use some shard or create new one
		// 3. Store data to memory
		// 4. Update lookup table
	}
}

// Get returns list of bytes if record is present in cache memory. If record is
// not found, then error is returned and list is nil.
func (a *AtomicCache) Get(key []byte) ([]byte, error) {
	var result []byte
	var nf = true

	value, status := a.lookup.Load(xxhash.Sum64(key))
	if status == true {
		sre := value.(SearchRecord)

		a.shardsLocker.RLock()
		if a.shards[sre.Shard] != nil {
			result = a.shards[sre.Shard].Get(sre.Index)
			nf = false
		}
		a.shardsLocker.RUnlock()

		if nf == false {
			return result, nil
		}
	}

	return nil, errors.New("Record not found")
}

// freeAfterExpiration frees memory after destinated time. It requires index and
// expiration time on input.
func (a *AtomicCache) freeAfterExpiration(hashKey uint64, expire time.Duration) {
	currentTime := time.Now()

	a.tasks.Store(currentTime, true)

	timer := time.NewTimer(expire)
	<-timer.C

	value, status := a.lookup.Load(hashKey)
	if status == true {
		sre := value.(SearchRecord)

		a.shardsLocker.RLock()
		a.shards[sre.Shard].Free(sre.Index)
		a.shardsLocker.RUnlock()

		a.shardsLocker.Lock()
		if a.releaseShard(sre.Shard) {
			a.lookup.Delete(hashKey)
		}
		a.shardsLocker.Unlock()
	}

	a.tasks.Delete(currentTime)
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
