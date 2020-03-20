package atomiccache

// Options are used for AtomicCache construct function.
type Options struct {
	// Size of byte array used for memory allocation at small shard section.
	RecordSizeSmall uint32
	// Size of byte array used for memory allocation at medium shard section.
	RecordSizeMedium uint32
	// Size of byte array used for memory allocation at large shard section.
	RecordSizeLarge uint32
	// Maximum records per shard.
	MaxRecords uint32
	// Maximum small shards which can be allocated in cache memory.
	MaxShardsSmall uint32
	// Maximum medium shards which can be allocated in cache memory.
	MaxShardsMedium uint32
	// Maximum large shards which can be allocated in cache memory.
	MaxShardsLarge uint32
	// Garbage collector starter (run garbage collection every X sets).
	GcStarter uint32
}

// Option specification for Printer package.
type Option func(*Options)

// OptionRecordSizeSmall option specification.
func OptionRecordSizeSmall(option uint32) Option {
	return func(opts *Options) {
		opts.RecordSizeSmall = option
	}
}

// OptionRecordSizeMedium option specification.
func OptionRecordSizeMedium(option uint32) Option {
	return func(opts *Options) {
		opts.RecordSizeMedium = option
	}
}

// OptionRecordSizeLarge option specification.
func OptionRecordSizeLarge(option uint32) Option {
	return func(opts *Options) {
		opts.RecordSizeLarge = option
	}
}

// OptionMaxRecords option specification.
func OptionMaxRecords(option uint32) Option {
	return func(opts *Options) {
		opts.MaxRecords = option
	}
}

// OptionMaxShardsSmall option specification.
func OptionMaxShardsSmall(option uint32) Option {
	return func(opts *Options) {
		opts.MaxShardsSmall = option
	}
}

// OptionMaxShardsMedium option specification.
func OptionMaxShardsMedium(option uint32) Option {
	return func(opts *Options) {
		opts.MaxShardsMedium = option
	}
}

// OptionMaxShardsLarge option specification.
func OptionMaxShardsLarge(option uint32) Option {
	return func(opts *Options) {
		opts.MaxShardsLarge = option
	}
}

// OptionGcStarter option specification.
func OptionGcStarter(option uint32) Option {
	return func(opts *Options) {
		opts.GcStarter = option
	}
}
