package atomiccache

import (
	"sync"
)

// Record structure represents one record stored in cache memory.
type Record struct {
	sync.RWMutex
	size  uint32
	alloc uint32
	data  []byte
}

// NewRecord initialize one new record and return pointer to them. During
// initialization is allocated maximum size of record. So we have record which
// is smaller than maximum size, then we set `alloc` property. It specifies how
// many bytes are used. It prevents garbage collector to take action.
func NewRecord(size uint32) *Record {
	return &Record{
		size:  size,
		alloc: 0,
		data:  make([]byte, size),
	}
}

// Set store data to record memory. On output we have bytes, which are copied to
// record data property and size is set.
func (r *Record) Set(data []byte) {
	dataLength := uint32(len(data))

	r.Lock() // Lock for writing and reading
	if dataLength > r.size {
		r.alloc = r.size
	} else {
		r.alloc = dataLength
	}
	copy(r.data, data)
	r.Unlock() // Unlock for writing and reading
}

// Get returns bytes based on size of virtual allocation. It means that it
// returns only specific count of bytes, based on alloc property. If array on
// output is empty, then record is not exists.
func (r *Record) Get() []byte {
	r.RLock() // Lock for reading
	data := r.data[:r.alloc]
	r.RUnlock() // Unlock for reading

	return data
}

// Free set alloc property to 0. Through this action, we empty memory of record
// without calling garbage collector.
func (r *Record) Free() {
	r.Lock() // Lock for writing and reading
	r.alloc = 0
	r.Unlock() // Unlock for writing and reading
}

// GetAllocated returns size of allocated bytes.
func (r *Record) GetAllocated() uint32 {
	r.RLock() // Lock for reading
	data := r.alloc
	r.RUnlock() // Unlock for reading
	return data
}

// GetDataLength returns real size of allocated bytes in memory.
func (r *Record) GetDataLength() uint32 {
	r.RLock() // Lock for reading
	data := uint32(len(r.data))
	r.RUnlock() // Unlock for reading
	return data
}
