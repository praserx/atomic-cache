package atomiccache

// import (
// 	"errors"
// 	"fmt"
// 	"sync"
// 	"time"

// 	"github.com/cespare/xxhash"
// )

// // Internal cache errors
// var (
// 	errNotFound  = errors.New("Record not found")
// 	errSetRecord = errors.New("Can't create new record, hit fail is too high")
// )

// // AtomicCache ...
// type AtomicCache struct {
// 	chronoList       []ChronoListItem
// 	chronoListLocker sync.RWMutex

// 	// Key is uint64 hash
// 	lookup sync.Map
// 	// Key is chronoIndex (int64)
// 	tasks sync.Map

// 	shards       []*Shard
// 	shardsActive []uint32
// 	shardsAvail  []uint32
// 	shardsLocker sync.RWMutex

// 	settings Settings
// }

// // Settings ...
// type Settings struct {
// 	// Size of byte array used for memory allocation.
// 	RecordSize uint32
// 	// Maximum records per shard.
// 	MaxRecords uint32
// }

// // ChronoListItem ...
// type ChronoListItem struct {
// 	key   int64
// 	value uint64
// }

// // SearchRecord ...
// type SearchRecord struct {
// 	ShardIndex  uint32
// 	RecordIndex uint32
// 	ChronoIndex int64
// }

// // New ...
// func New(opts ...Option) *AtomicCache {
// 	var options = &Options{
// 		RecordSize: 4096,
// 		MaxRecords: 4096,
// 		MaxShards:  128,
// 	}

// 	for _, opt := range opts {
// 		opt(options)
// 	}

// 	// Init cache structure
// 	cache := &AtomicCache{}

// 	// Save cache settings
// 	cache.settings = Settings{
// 		RecordSize: options.RecordSize,
// 		MaxRecords: options.MaxRecords,
// 	}

// 	// Init shards list with nil values
// 	cache.shards = make([]*Shard, options.MaxShards, options.MaxShards)

// 	// Create shard available indexes
// 	for i := uint32(0); i < options.MaxShards; i++ {
// 		cache.shardsAvail = append(cache.shardsAvail, i)
// 	}

// 	return cache
// }

// // Set ...
// // 1. Is there any free space? If not make some (remove oldest record)
// // 2. Use some shard or create new one
// // 3. Store data to memory
// // 4. Update lookup table
// func (a *AtomicCache) Set(key, data []byte, expire time.Duration) error {
// 	var ri, si uint32
// 	var shardAvail, shardAvailSpace bool
// 	var chronoKey int64

// 	hash := xxhash.Sum64(key)

// 	if expire == 0 {
// 		chronoKey = 9223372036854775807 // max int64
// 	} else {
// 		chronoKey = a.newChronoIndex(expire)
// 	}

// 	value, status := a.lookup.Load(hash)
// 	if status == true {
// 		si = (value.(SearchRecord)).ShardIndex
// 		ri = (value.(SearchRecord)).RecordIndex

// 		a.shardsLocker.RLock()   // Lock for reading
// 		a.shards[si].Free(ri)    // Empty old record
// 		a.shardsLocker.RUnlock() // Unlock for reading

// 		a.chronoListLocker.Lock()
// 		a.chronoDeleteByValue(hash)
// 		a.chronoListLocker.Unlock()
// 	} else {
// 		a.shardsLocker.RLock()
// 		si, shardAvailSpace = a.getShard()
// 		a.shardsLocker.RUnlock()

// 		if shardAvailSpace != true {
// 			a.shardsLocker.RLock()
// 			si, shardAvail = a.getShardAvail()
// 			a.shardsLocker.RUnlock()

// 			if shardAvail {
// 				a.shardsLocker.Lock() // Lock for writing and reading
// 				a.shards[si] = NewShard(a.settings.MaxRecords, a.settings.RecordSize)
// 				a.shardsLocker.Unlock() // Unlock for writing and reading
// 			} else {
// 				var hitFail int

// 				for {
// 					if hitFail >= 69 { // need some magic number
// 						return errSetRecord
// 					}

// 					var candidateKey uint64
// 					a.chronoListLocker.Lock()
// 					fmt.Println(len(a.chronoList))
// 					candidateKey, a.chronoList = a.chronoList[0].value, a.chronoList[1:]
// 					a.chronoListLocker.Unlock()

// 					value, status := a.lookup.Load(candidateKey)
// 					if status == true {
// 						si = (value.(SearchRecord)).ShardIndex
// 						ri = (value.(SearchRecord)).RecordIndex
// 						break
// 					}

// 					hitFail++
// 				}

// 				a.shardsLocker.RLock()   // Lock for reading
// 				a.shards[si].Free(ri)    // Empty old record
// 				a.shardsLocker.RUnlock() // Unlock for reading

// 				a.chronoListLocker.Lock()
// 				a.chronoDeleteByValue(hash)
// 				a.chronoListLocker.Unlock()
// 			}
// 		}
// 	}

// 	a.shardsLocker.RLock()      // Lock for reading
// 	ri = a.shards[si].Set(data) // Set data update
// 	a.shardsLocker.RUnlock()    // Unlock for reading

// 	a.lookup.Store(hash, SearchRecord{ShardIndex: si, RecordIndex: ri}) // Update lookup table
// 	go a.freeAfterExpiration(hash, chronoKey, expire)                   // Start expiration worker

// 	return nil
// }

// // Get returns list of bytes if record is present in cache memory. If record is
// // not found, then error is returned and list is nil.
// func (a *AtomicCache) Get(key []byte) ([]byte, error) {
// 	var result []byte
// 	var nf = true

// 	value, status := a.lookup.Load(xxhash.Sum64(key))
// 	if status == true {
// 		sre := value.(SearchRecord)

// 		a.shardsLocker.RLock()
// 		if a.shards[sre.ShardIndex] != nil {
// 			result = a.shards[sre.ShardIndex].Get(sre.RecordIndex)
// 			nf = false
// 		}
// 		a.shardsLocker.RUnlock()

// 		if nf == false {
// 			return result, nil
// 		}
// 	}

// 	return nil, errNotFound
// }

// // freeAfterExpiration frees memory after destinated time. It requires index and
// // expiration time on input.
// func (a *AtomicCache) freeAfterExpiration(hashKey uint64, chronoKey int64, expire time.Duration) {
// 	a.tasks.Store(chronoKey, true)

// 	a.chronoListLocker.Lock()
// 	a.chronoAppend(chronoKey, hashKey)
// 	a.chronoListLocker.Unlock()

// 	timer := time.NewTimer(expire)
// 	<-timer.C

// 	a.chronoListLocker.Lock()
// 	a.chronoDelete(chronoKey)
// 	a.chronoListLocker.Unlock()

// 	value, status := a.lookup.Load(hashKey)
// 	if status == true {
// 		sre := value.(SearchRecord)

// 		a.shardsLocker.RLock()
// 		a.shards[sre.ShardIndex].Free(sre.RecordIndex)
// 		a.shardsLocker.RUnlock()

// 		a.shardsLocker.Lock()
// 		if a.releaseShard(sre.ShardIndex) {
// 			a.lookup.Delete(hashKey)
// 		}
// 		a.shardsLocker.Unlock()
// 	}

// 	a.tasks.Delete(chronoKey)
// }

// // releaseShard release shard if there is no record in memory. It returns true
// // if shard was released. This method is not thread safe and additional locks
// // are required.
// func (a *AtomicCache) releaseShard(shard uint32) bool {
// 	if a.shards[shard].GetSlotsAvail() == a.settings.MaxRecords {
// 		a.shards[shard] = nil
// 		return true
// 	}
// 	return false
// }

// // getShard return index of shard which have some available space for new
// // record. If there is no shard with available space, then false is returned as
// // a second value. This method is not thread safe and additional locks are
// // required.
// func (a *AtomicCache) getShard() (uint32, bool) {
// 	for _, shardIndex := range a.shardsActive {
// 		if a.shards[shardIndex].GetSlotsAvail() != 0 {
// 			return shardIndex, true
// 		}
// 	}

// 	return 0, false
// }

// // getShardAvail return index of shard that can be used for new shard
// // allocation. If there is no left index, then false is returned as a second
// // value. This method is not thread safe and additional locks are required.
// func (a *AtomicCache) getShardAvail() (uint32, bool) {
// 	if len(a.shardsAvail) == 0 {
// 		return 0, false
// 	}

// 	var shardIndex uint32
// 	shardIndex, a.shardsAvail = a.shardsAvail[0], a.shardsAvail[1:]

// 	return shardIndex, true
// }

// // chronoAppend ...
// func (a *AtomicCache) chronoAppend(chronoKey int64, key uint64) {
// 	for chronoIndex, chronoItem := range a.chronoList {
// 		if chronoKey < chronoItem.key {
// 			a.chronoList = append(a.chronoList, ChronoListItem{})
// 			copy(a.chronoList[chronoIndex+1:], a.chronoList[chronoIndex:])
// 			a.chronoList[chronoIndex] = ChronoListItem{key: chronoKey, value: key}
// 			return
// 		}
// 	}

// 	a.chronoList = append(a.chronoList, ChronoListItem{key: chronoKey, value: key})
// 	return
// }

// // chronoDelete ...
// func (a *AtomicCache) chronoDelete(chronoKey int64) {
// 	for chronoIndex, chronoItem := range a.chronoList {
// 		if chronoKey == chronoItem.key {
// 			copy(a.chronoList[chronoIndex:], a.chronoList[chronoIndex+1:])
// 			a.chronoList[len(a.chronoList)-1] = ChronoListItem{}
// 			a.chronoList = a.chronoList[:len(a.chronoList)-1]
// 			return
// 		}
// 	}
// }

// // chronoDeleteByValue ...
// func (a *AtomicCache) chronoDeleteByValue(value uint64) {
// 	for chronoIndex, chronoItem := range a.chronoList {
// 		if value == chronoItem.value {
// 			copy(a.chronoList[chronoIndex:], a.chronoList[chronoIndex+1:])
// 			a.chronoList[len(a.chronoList)-1] = ChronoListItem{}
// 			a.chronoList = a.chronoList[:len(a.chronoList)-1]
// 		}
// 	}
// }

// // newChronoIndex ...
// func (a *AtomicCache) newChronoIndex(expire time.Duration) int64 {
// 	return time.Now().Add(expire).UnixNano()
// }
