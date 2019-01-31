package atomiccache

// Options are used for AtomicCache construct function.
type Options struct {
	// Memory record size in bytes
	RecordSize uint
}

// Option specification for Printer package.
type Option func(*Options)

// OptionRecordSize option specification.
func OptionRecordSize(option uint) Option {
	return func(opts *Options) {
		opts.RecordSize = option
	}
}
