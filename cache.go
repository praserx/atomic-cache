package atomiccache

// https://github.com/cespare/xxhash

// AtomicCache ...
type AtomicCache struct {
	RecordSize uint
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
func (a *AtomicCache) Set(data []byte) {

}

// Get ...
func (a *AtomicCache) Get() []byte {
	return nil
}
