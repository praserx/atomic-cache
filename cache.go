package atomiccache

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// Internal cache errors
var (
	ErrNotFound   = errors.New("record not found")
	ErrDataLimit  = errors.New("cannot create new record: it violates data limit")
	ErrFullMemory = errors.New("cannot create new record: memory is full")
)

// Constants below are used for shard section identification.
const (
	// SMSH - Small Shards section
	SMSH = iota + 1
	// MDSH - Medium Shards section
	MDSH
	// LGSH - Large Shards section
	LGSH
)

// KeepTTL is used for setting expiration time to current expiration time.
// It means that record will be updated with the same expiration time.
const KeepTTL = time.Duration(-1)

// AtomicCache structure represents whole cache memory.
type AtomicCache struct {
	// RWMutex is used for access to shards array.
	sync.RWMutex
	// Deadlock mutex for debugging purpose.
	// deadlock.RWMutex

	// Lookup structure used for global index.
	lookup map[string]LookupRecord

	// Shards lookup tables which contain information about shard sections.
	smallShards, mediumShards, largeShards ShardsLookup

	// Size of byte array used for memory allocation at small shard section.
	RecordSizeSmall int
	// Size of byte array used for memory allocation at medium shard section.
	RecordSizeMedium int
	// Size of byte array used for memory allocation at large shard section.
	RecordSizeLarge int

	// Maximum records per shard.
	MaxRecords int

	// Maximum small shards which can be allocated in cache memory.
	MaxShardsSmall int
	// Maximum medium shards which can be allocated in cache memory.
	MaxShardsMedium int
	// Maximum large shards which can be allocated in cache memory.
	MaxShardsLarge int

	// Garbage collector starter (run garbage collection every X memory sets).
	GcStarter uint32
	// Garbage collector counter for starter.
	GcCounter uint32

	// Buffer contains all unattended cache set requests. It has a maximum size
	// which is equal to the MaxRecords value.
	buffer []BufferItem
}

// ShardsLookup represents data structure for for each shards section. In each
// section we have different size of records in that shards.
type ShardsLookup struct {
	// Array of pointers to shard objects.
	shards []*Shard
	// Array of shard indexes which are currently active.
	shardsActive []int
	// Array of shard indexes which are currently available for new allocation.
	shardsAvail []int
}

// LookupRecord represents an item in the lookup table. One record contains the index of
// the shard and record. So we can determine which shard to access and which record of
// the shard to get. Record also contains expiration time.
type LookupRecord struct {
	RecordIndex  int
	ShardIndex   int
	ShardSection int
	Expiration   time.Time
}

// BufferItem is used for the buffer, which contains all unattended cache set
// requests.
type BufferItem struct {
	Key    string
	Data   []byte
	Expire time.Duration
}

// New initializes the whole cache memory with one allocated shard.
func New(opts ...Option) *AtomicCache {
	var options = &Options{
		RecordSizeSmall:  512,
		RecordSizeMedium: 2048,
		RecordSizeLarge:  8128,
		MaxRecords:       2048,
		MaxShardsSmall:   256,
		MaxShardsMedium:  128,
		MaxShardsLarge:   64,
		GcStarter:        512000,
	}

	for _, opt := range opts {
		opt(options)
	}

	// Init cache structure
	cache := &AtomicCache{}

	// Init lookup table
	cache.lookup = make(map[string]LookupRecord)

	// Init small shards section
	initShardsSection(&cache.smallShards, options.MaxShardsSmall, options.MaxRecords, options.RecordSizeSmall)
	initShardsSection(&cache.mediumShards, options.MaxShardsMedium, options.MaxRecords, options.RecordSizeMedium)
	initShardsSection(&cache.largeShards, options.MaxShardsLarge, options.MaxRecords, options.RecordSizeLarge)

	// Define setup values
	cache.RecordSizeSmall = options.RecordSizeSmall
	cache.RecordSizeMedium = options.RecordSizeMedium
	cache.RecordSizeLarge = options.RecordSizeLarge
	cache.MaxRecords = options.MaxRecords
	cache.MaxShardsSmall = options.MaxShardsSmall
	cache.MaxShardsMedium = options.MaxShardsMedium
	cache.MaxShardsLarge = options.MaxShardsLarge
	cache.GcStarter = options.GcStarter

	return cache
}

// initShardsSection provides shard section initialization. So the cache has
// one shard in each section at the beginning.
func initShardsSection(shardsSection *ShardsLookup, maxShards, maxRecords, recordSize int) {
	var shardIndex int

	shardsSection.shards = make([]*Shard, maxShards)
	for i := 0; i < maxShards; i++ {
		shardsSection.shardsAvail = append(shardsSection.shardsAvail, i)
	}

	shardIndex, shardsSection.shardsAvail = shardsSection.shardsAvail[0], shardsSection.shardsAvail[1:]
	shardsSection.shardsActive = append(shardsSection.shardsActive, shardIndex)
	shardsSection.shards[shardIndex] = NewShard(maxRecords, recordSize)
}

// Set stores data to cache memory. If the key/record is already in memory, then data
// are replaced. If not, it checks if there is an allocated shard with empty
// space for data. If there is no empty space, a new shard is allocated.
// Remarks:
// - If expiration time is set to 0 then maximum expiration time is used (48 hours).
// - If expiration time is KeepTTL, then current expiration time is preserved.
func (a *AtomicCache) Set(key string, data []byte, expire time.Duration) error {
	// Reject if data is too large for any shard
	if len(data) > int(a.RecordSizeLarge) {
		return ErrDataLimit
	}

	// Track if this is a new record and if garbage collection should be triggered
	new := false
	collectGarbage := false

	// Select the appropriate shard section based on data size
	shardSection, shardSectionID := a.getShardsSectionBySize(len(data))

	var (
		exists bool
		val    LookupRecord
	)

	// Only lock for shared state mutation: check if key exists in lookup
	a.RLock()
	val, exists = a.lookup[key]
	a.RUnlock()

	// Determine expiration time: if KeepTTL and record exists, preserve old
	// expiration; otherwise, calculate new.
	var expireTime time.Time
	if expire == KeepTTL && exists {
		expireTime = val.Expiration
	} else {
		expireTime = a.getExprTime(expire)
	}

	if !exists {
		// Key is new, will allocate new record
		new = true
	} else {
		if val.ShardSection != shardSectionID {
			// Key exists but data size changed: move to new section, free old record.
			// Explanation: If the record size changed and data should be stored in a different
			// shard section, we need to free the old record and allocate a new record in
			// the correct shard section.
			a.Lock()
			shardSection.shards[val.ShardIndex].Free(val.RecordIndex)
			val.RecordIndex = shardSection.shards[val.ShardIndex].Set(data)
			a.lookup[key] = LookupRecord{ShardIndex: val.ShardIndex, ShardSection: shardSectionID, RecordIndex: val.RecordIndex, Expiration: expireTime}
			a.Unlock()
		} else {
			// Key exists in same section: update existing record.
			// Explanation: If the record size is the same, we can simply update the existing record
			// in the same shard section without needing to free it first.
			// This is more efficient as it avoids unnecessary memory allocation and deallocation.
			// This is a performance optimization to avoid unnecessary memory allocation and deallocation.
			// It assumes that the record size has not changed and we can safely update it.
			a.Lock()
			shardSection.shards[val.ShardIndex].Seti(val.RecordIndex, data)
			a.Unlock()
		}
	}

	if new {
		// Allocate new record: try to find a shard with space, or allocate a new shard, or buffer if full
		a.Lock()
		if si, ok := a.getShard(shardSectionID); ok {
			// Found shard with available slot.
			// Explanation: If we found a shard with available space, we can simply set the data
			// in that shard and update the lookup table with the new record index.
			// This avoids unnecessary memory allocation and deallocation, improving performance.
			ri := shardSection.shards[si].Set(data)
			a.lookup[key] = LookupRecord{ShardIndex: si, ShardSection: shardSectionID, RecordIndex: ri, Expiration: expireTime}
			a.Unlock()
		} else if si, ok := a.getEmptyShard(shardSectionID); ok {
			// No shard with space, allocate new shard.
			// Explanation: If there is no shard with available space, we allocate a new shard
			// and set the data in that new shard. This is necessary when all existing shards
			// are full and we need to create a new shard to accommodate the new record.
			// This ensures that we can always store new records, even if it means creating a
			// new shard when all existing shards are full.
			shardSection.shards[si] = NewShard(a.MaxRecords, a.getRecordSizeByShardSectionID(shardSectionID))
			ri := shardSection.shards[si].Set(data)
			a.lookup[key] = LookupRecord{ShardIndex: si, ShardSection: shardSectionID, RecordIndex: ri, Expiration: expireTime}
			a.Unlock()
		} else {
			// All shards full, buffer the request or return error if buffer is full.
			if len(a.buffer) < int(a.MaxRecords) {
				// Buffer the request if there is space in buffer.
				// Explanation: If the buffer has space, we can store the request in the buffer
				// instead of allocating a new shard. This allows us to handle more requests without
				// immediately allocating new memory, which can be more efficient.
				// This is useful when the cache is under heavy load and we want to avoid
				// allocating new shards for every request.
				a.buffer = append(a.buffer, BufferItem{Key: key, Data: data, Expire: expire})
				a.Unlock()
			} else {
				a.Unlock()
				return ErrFullMemory
			}
			collectGarbage = true
		}
	}

	// Trigger garbage collection if needed
	if (atomic.AddUint32(&a.GcCounter, 1) == a.GcStarter) || collectGarbage {
		atomic.StoreUint32(&a.GcCounter, 0)
		go a.collectGarbage()
	}

	return nil
}

// Get returns list of bytes if record is present in cache memory. If record is
// not found, then error is returned and list is nil.
func (a *AtomicCache) Get(key string) ([]byte, error) {
	a.RLock()
	val, ok := a.lookup[key]
	a.RUnlock()

	if ok {
		shardSection := a.getShardsSectionByID(val.ShardSection)
		if shardSection.shards[val.ShardIndex] != nil && time.Now().Before(val.Expiration) {
			return shardSection.shards[val.ShardIndex].Get(val.RecordIndex), nil
		}
	}

	return nil, ErrNotFound
}

// Exists checks if record is present in cache memory. It returns true if record
// is present, otherwise false.
func (a *AtomicCache) Exists(key string) bool {
	a.RLock()
	val, ok := a.lookup[key]
	a.RUnlock()
	if !ok {
		return false
	}
	// Check expiration
	if time.Now().After(val.Expiration) {
		return false
	}
	return true
}

// Delete removes record from cache memory. If record is not found, then error
// is returned. It also releases memory used by record in shard.
// If shard ends up empty, it is released.
func (a *AtomicCache) Delete(key string) error {
	a.Lock()
	defer a.Unlock()

	val, ok := a.lookup[key]
	if !ok {
		return ErrNotFound
	}

	shardSection := a.getShardsSectionByID(val.ShardSection)
	// Check if the shard at val.ShardIndex is nil. This is a defensive check to
	// handle cases where the shard might have been released or not initialized
	// due to concurrent modifications or unexpected states.
	if shardSection.shards[val.ShardIndex] != nil {
		shardSection.shards[val.ShardIndex].Free(val.RecordIndex)
		a.releaseShard(val.ShardSection, val.ShardIndex)
		delete(a.lookup, key)
		return nil
	}

	return ErrNotFound
}

// releaseShard release shard if there is no record in memory. It returns true
// if shard was released. The function requires the shard section ID and
// shard ID on input.
// This method is not thread safe and additional locks are required.
func (a *AtomicCache) releaseShard(shardSectionID int, shard int) bool {
	var shardSection *ShardsLookup

	if shardSection = a.getShardsSectionByID(shardSectionID); shardSection == nil {
		return false
	}

	if shardSection.shards[shard].IsEmpty() {
		shardSection.shards[shard] = nil

		shardSection.shardsAvail = append(shardSection.shardsAvail, shard)
		for k, v := range shardSection.shardsActive {
			if v == shard {
				shardSection.shardsActive = append(shardSection.shardsActive[:k], shardSection.shardsActive[k+1:]...)
				break
			}
		}

		return true
	}

	return false
}

// getShard return index of shard which have some available space for new
// record. If there is no shard with available space, then false is returned as
// a second value. The function requires the shard section ID on input.
// This method is not thread safe and additional locks are required.
func (a *AtomicCache) getShard(shardSectionID int) (int, bool) {
	var shardSection *ShardsLookup

	if shardSection = a.getShardsSectionByID(shardSectionID); shardSection == nil {
		return 0, false
	}

	for _, shardIndex := range shardSection.shardsActive {
		if shardSection.shards[shardIndex].GetSlotsAvail() != 0 {
			return shardIndex, true
		}
	}

	return 0, false
}

// getEmptyShard return index of shard that can be used for new shard
// allocation. If there is no left index, then false is returned as a second
// value. The function requires the shard section ID on input.
// This method is not thread safe and additional locks are required.
func (a *AtomicCache) getEmptyShard(shardSectionID int) (int, bool) {
	var shardSection *ShardsLookup

	if shardSection = a.getShardsSectionByID(shardSectionID); shardSection == nil {
		return 0, false
	}

	if len(shardSection.shardsAvail) == 0 {
		return 0, false
	}

	var shardIndex int
	shardIndex, shardSection.shardsAvail = shardSection.shardsAvail[0], shardSection.shardsAvail[1:]
	shardSection.shardsActive = append(shardSection.shardsActive, shardIndex)

	return shardIndex, true
}

// getShardsSectionBySize returns the shard section lookup structure and section
// identifier as a second value. The function requires the data size value as input.
// If data are bigger than the allowed value, then nil and 0 are returned.
// This method is not thread safe and additional locks are required.
func (a *AtomicCache) getShardsSectionBySize(dataSize int) (*ShardsLookup, int) {
	if dataSize <= int(a.RecordSizeSmall) {
		return &a.smallShards, SMSH
	} else if dataSize > int(a.RecordSizeSmall) && dataSize <= int(a.RecordSizeMedium) {
		return &a.mediumShards, MDSH
	} else if dataSize > int(a.RecordSizeMedium) && dataSize <= int(a.RecordSizeLarge) {
		return &a.largeShards, LGSH
	}

	return nil, 0
}

// getShardsSectionByID returns shards section lookup structure. The function
// requires the shard section ID on input. If section ID is not valid, nil
// is returned.
// This method is not thread safe and additional locks are required.
func (a *AtomicCache) getShardsSectionByID(sectionID int) *ShardsLookup {
	switch sectionID {
	case SMSH:
		return &a.smallShards
	case MDSH:
		return &a.mediumShards
	case LGSH:
		return &a.largeShards
	}

	return nil
}

// getRecordSizeByShardSectionID returns maximum record size for specified
// shard section ID. It returns 0 if there is not known section ID on input.
// This method is not thread safe and additional locks are required.
func (a *AtomicCache) getRecordSizeByShardSectionID(sectionID int) int {
	switch sectionID {
	case SMSH:
		return a.RecordSizeSmall
	case MDSH:
		return a.RecordSizeMedium
	case LGSH:
		return a.RecordSizeLarge
	}

	return 0
}

// getExprTime return expiration time based on duration. If duration is 0, then
// maximum expiration time is used (48 hours).
func (a *AtomicCache) getExprTime(expire time.Duration) time.Time {
	if expire == 0 {
		return time.Now().Add(48 * time.Hour)
	}

	return time.Now().Add(expire)
}

// collectGarbage provides garbage collection. It goes through the lookup table and
// checks expiration time. If a shard ends up empty, then garbage collection releases
// it, but only if there is more than one shard in use (there is always at least one active shard).
func (a *AtomicCache) collectGarbage() {
	a.Lock()
	for k, v := range a.lookup {
		shardSection := a.getShardsSectionByID(v.ShardSection) // get shard section
		if time.Now().After(v.Expiration) {
			shardSection.shards[v.ShardIndex].Free(v.RecordIndex)
			if len(shardSection.shardsActive) > 1 {
				a.releaseShard(v.ShardSection, v.ShardIndex)
			}
			delete(a.lookup, k)
		}
	}

	// Properly copy buffer to avoid concurrency issues
	localBuffer := append([]BufferItem(nil), a.buffer...)
	a.buffer = nil

	a.Unlock()

	var bi BufferItem
	for x := 0; x < len(localBuffer); x++ {
		bi, localBuffer = localBuffer[0], localBuffer[1:]
		if err := a.Set(bi.Key, bi.Data, bi.Expire); err != nil {
			break
		}
	}
}
