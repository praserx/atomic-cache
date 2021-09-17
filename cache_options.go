package atomiccache

// Options are used for AtomicCache construct function.
type Options struct {
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
	// Garbage collector starter (run garbage collection every X sets).
	GcStarter uint32
}

// Option specification for Printer package.
type Option func(*Options)

// OptionRecordSizeSmall option specification.
func OptionRecordSizeSmall(option int) Option {
	return func(opts *Options) {
		opts.RecordSizeSmall = option
	}
}

// OptionRecordSizeMedium option specification.
func OptionRecordSizeMedium(option int) Option {
	return func(opts *Options) {
		opts.RecordSizeMedium = option
	}
}

// OptionRecordSizeLarge option specification.
func OptionRecordSizeLarge(option int) Option {
	return func(opts *Options) {
		opts.RecordSizeLarge = option
	}
}

// OptionMaxRecords option specification.
func OptionMaxRecords(option int) Option {
	return func(opts *Options) {
		opts.MaxRecords = option
	}
}

// OptionMaxShardsSmall option specification.
func OptionMaxShardsSmall(option int) Option {
	return func(opts *Options) {
		opts.MaxShardsSmall = option
	}
}

// OptionMaxShardsMedium option specification.
func OptionMaxShardsMedium(option int) Option {
	return func(opts *Options) {
		opts.MaxShardsMedium = option
	}
}

// OptionMaxShardsLarge option specification.
func OptionMaxShardsLarge(option int) Option {
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
