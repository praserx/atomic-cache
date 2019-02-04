package atomiccache

// https://github.com/cespare/xxhash

import (
	"github.com/cespare/xxhash"
)

// AtomicCache ...
type AtomicCache struct {
	recordSize  uint32
	recordShard uint32 // nepotrebne

	recordStack     []uint64
	recordStackSize uint64
	lookup          map[uint64][]uint32

	allocs uint32
	shards []Shard
}

// New ...
func New(opts ...Option) *AtomicCache {
	var options = &Options{
		RecordSize: 2048,
	}

	for _, opt := range opts {
		opt(options)
	}

	return &AtomicCache{}
}

// Set ...
func (a *AtomicCache) Set(key, data []byte) {

	hash := xxhash.Sum64(key)

	if a.lookup[hash] == nil {

	} else {

	}
}

// Get ...
func (a *AtomicCache) Get() []byte {
	return nil
}
