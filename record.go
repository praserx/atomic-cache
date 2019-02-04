package atomiccache

import (
	"time"
)

// Record structure represents one record stored in cache memory. During
// initialization is allocated maximum size of record. It prevents garbage
// collector to take action.
type Record struct {
	size  uint32
	alloc uint32
	data  []byte
}

// NewRecord initialize one new record and return pointer to them.
func NewRecord(size uint32) *Record {
	return &Record{
		size:  size,
		alloc: 0,
		data:  make([]byte, size),
	}
}

func (r *Record) Set(data []byte) {
	dataLength := uint32(len(data))
	if dataLength > r.size {
		r.alloc = r.size
	} else {
		r.alloc = dataLength
	}
	copy(r.data, data)
}

func (r *Record) Get() []byte {
	return r.data[:r.alloc]
}

func (r *Record) Free(expiration time.Duration) {
	go func() {
		// plan free memory using time.Ticker
		// then allocated = 0
	}()
}

// GetAllocated returns size of allocated bytes.
func (r *Record) GetAllocated() uint32 {
	return r.alloc
}

// GetDataLength returns real size of allocated bytes in memory.
func (r *Record) GetDataLength() uint32 {
	return uint32(len(r.data))
}
