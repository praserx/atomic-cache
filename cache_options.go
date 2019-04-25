package atomiccache

// Options are used for AtomicCache construct function.
type Options struct {
	// Size of byte array used for memory allocation.
	RecordSize uint32
	// Maximum records per shard.
	MaxRecords uint32
	// Maximum shards which can be allocated in cache memory.
	MaxShards uint32
	// Garbage collector starter (run garbage collection every X sets).
	GcStarter uint32
}

// Option specification for Printer package.
type Option func(*Options)

// OptionRecordSize option specification.
func OptionRecordSize(option uint32) Option {
	return func(opts *Options) {
		opts.RecordSize = option
	}
}

// OptionMaxRecords option specification.
func OptionMaxRecords(option uint32) Option {
	return func(opts *Options) {
		opts.MaxRecords = option
	}
}

// OptionMaxShards option specification.
func OptionMaxShards(option uint32) Option {
	return func(opts *Options) {
		opts.MaxShards = option
	}
}

// OptionGcStarter option specification.
func OptionGcStarter(option uint32) Option {
	return func(opts *Options) {
		opts.GcStarter = option
	}
}
