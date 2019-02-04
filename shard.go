package atomiccache

// Shard ...
type Shard struct {
	Allocs  uint32
	Records []*Record
}

// NewShard ...
func NewShard(allocs, size uint32) *Shard {

	shard := &Shard{}

	for i := uint32(0); i < allocs; i++ {
		shard.Records = append(shard.Records, NewRecord(size))
	}

	return nil
}

func (s *Shard) Set(i uint32, data []byte) {

}
