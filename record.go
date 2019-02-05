package atomiccache

import (
	"sync"
)

// Record structure represents one record stored in cache memory.
type Record struct {
	size  uint32
	alloc uint32
	data  []byte
	mutex *sync.RWMutex
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
		mutex: &sync.RWMutex{},
	}
}

// Set store data to record memory. On output we have bytes, which are copied to
// record data property and size is set.
func (r *Record) Set(data []byte) {
	dataLength := uint32(len(data))

	r.mutex.Lock() // Lock for writing and reading
	if dataLength > r.size {
		r.alloc = r.size
	} else {
		r.alloc = dataLength
	}
	copy(r.data, data)
	r.mutex.Unlock() // Unlock for writing and reading
}

// Get returns bytes based on size of virtual allocation. It means that it
// returns only specific count of bytes, based on alloc property.
func (r *Record) Get() []byte {
	r.mutex.RLock() // Lock for reading
	data := r.data[:r.alloc]
	r.mutex.RUnlock() // Unlock for reading

	return data
}

// Free set alloc property to 0. Through this action, we empty memory of record
// without calling garbage collector.
func (r *Record) Free() {
	r.mutex.Lock() // Lock for writing and reading
	r.alloc = 0
	r.mutex.Unlock() // Unlock for writing and reading
}

// GetAllocated returns size of allocated bytes.
func (r *Record) GetAllocated() uint32 {
	r.mutex.RLock() // Lock for reading
	data := r.alloc
	r.mutex.RUnlock() // Unlock for reading
	return data
}

// GetDataLength returns real size of allocated bytes in memory.
func (r *Record) GetDataLength() uint32 {
	r.mutex.RLock() // Lock for reading
	data := uint32(len(r.data))
	r.mutex.RUnlock() // Unlock for reading
	return data
}
