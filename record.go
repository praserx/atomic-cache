package atomiccache

import (
	"sync"
)

// Record structure represents one record stored in cache memory.
type Record struct {
	sync.RWMutex
	size  int
	alloc int
	data  []byte
}

// NewRecord initialize one new record and return pointer to them. During
// initialization is allocated maximum size of record. So we have record which
// is smaller than maximum size, then we set `alloc` property. It specifies how
// many bytes are used. It prevents garbage collector to take action.
func NewRecord(size int) *Record {
	return &Record{
		size:  size,
		alloc: 0,
		data:  make([]byte, size),
	}
}

// Set store data to record memory. On output we have bytes, which are copied to
// record data property and size is set.
func (r *Record) Set(data []byte) {
	if len(data) > r.size {
		r.alloc = r.size
	} else {
		r.alloc = len(data)
	}
	copy(r.data, data)
}

// Get returns bytes based on size of virtual allocation. It means that it
// returns only specific count of bytes, based on alloc property. If array on
// output is empty, then record is not exists.
func (r *Record) Get() (data []byte) {
	data = r.data[:r.alloc]
	return
}

// Free set alloc property to 0. Through this action, we empty memory of record
// without calling garbage collector.
func (r *Record) Free() {
	r.alloc = 0
}

// GetAllocated returns size of allocated bytes.
func (r *Record) GetAllocated() (size int) {
	size = r.alloc
	return
}

// GetDataLength returns real size of allocated bytes in memory.
func (r *Record) GetDataLength() (size int) {
	size = len(r.data)
	return
}
