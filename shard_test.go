package atomiccache

import (
	"reflect"
	"testing"
)

func TestShardSeti(t *testing.T) {
	shard := NewShard(10, 4)
	// Use Set to get a valid index, then SetI to update
	idx := shard.Set([]byte{0, 0, 0, 0})
	shard.Seti(idx, []byte{1, 2, 3, 4})
	got := shard.Get(idx)
	want := []byte{1, 2, 3, 4}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SetI: got %v, want %v", got, want)
	}

	// SetI with out-of-bounds index (should panic, so recover)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("SetI should panic for out-of-bounds index")
		}
	}()
	shard.Seti(20, []byte{9, 9, 9, 9})
}

func TestShardFreeAndIsEmpty(t *testing.T) {
	shard := NewShard(5, 4)
	idx := shard.Set([]byte{1, 2, 3, 4})
	if shard.IsEmpty() {
		t.Errorf("Shard should not be empty after Set")
	}
	shard.Free(idx)
	if !shard.IsEmpty() {
		t.Errorf("Shard should be empty after Free")
	}
}

func TestShardGetSlotsAvail(t *testing.T) {
	shard := NewShard(3, 2)
	if avail := shard.GetSlotsAvail(); avail != 3 {
		t.Errorf("Expected 3 slots available, got %d", avail)
	}
	idx := shard.Set([]byte{1, 2})
	if avail := shard.GetSlotsAvail(); avail != 2 {
		t.Errorf("Expected 2 slots available after Set, got %d", avail)
	}
	shard.Free(idx)
	if avail := shard.GetSlotsAvail(); avail != 3 {
		t.Errorf("Expected 3 slots available after Free, got %d", avail)
	}
}

func TestShardSimple(t *testing.T) {
	for _, c := range []struct {
		recordCount int
		recordSize  int
		in          []byte
		want        []byte
	}{
		{256, 2048, []byte{0}, []byte{0}},
		{256, 2048, []byte{0, 1, 2, 3, 4, 5}, []byte{0, 1, 2, 3, 4, 5}},
		{256, 1, []byte{0, 1, 2}, []byte{0}},
	} {
		shard := NewShard(c.recordCount, c.recordSize)
		index := shard.Set(c.in)
		if !reflect.DeepEqual(shard.Get(index), c.want) {
			t.Errorf("%v != %v", shard.Get(index), c.want)
		}
	}
}

func TestShardIntermediate(t *testing.T) {
	for _, c := range []struct {
		recordCount int
		recordSize  int
		in          []byte
		want        []byte
	}{
		{1024, 1024, []byte("test value"), []byte("test value")},
	} {
		var indexes []int
		shard := NewShard(c.recordCount, c.recordSize)

		for i := 0; i < c.recordCount; i++ {
			indexes = append(indexes, shard.Set(c.in))
		}

		// Check if value on index 0 is present
		if !reflect.DeepEqual(shard.Get(indexes[0]), c.want) {
			t.Errorf("%v != %v", shard.Get(indexes[0]), c.want)
		}

		// Check if value on index 128 is present
		if !reflect.DeepEqual(shard.Get(indexes[128]), c.want) {
			t.Errorf("%v != %v", shard.Get(indexes[0]), c.want)
		}
	}
}

func benchmarkShardNew(recordCount, recordSize int, b *testing.B) {
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		NewShard(recordCount, recordSize)
	}
}

func BenchmarkShardNewSmall(b *testing.B) {
	benchmarkShardNew(512, 2048, b)
}

func BenchmarkShardNewMedium(b *testing.B) {
	benchmarkShardNew(2048, 2048, b)
}

func BenchmarkShardNewLarge(b *testing.B) {
	benchmarkShardNew(16384, 4096, b)
}

func benchmarkShardSet(recordCount, recordSize, dataSize int, b *testing.B) {
	var data []byte
	shard := NewShard(recordCount, recordSize)

	for i := 0; i < dataSize; i++ {
		data = append(data, 1)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		index := shard.Set(data)
		shard.Free(index)
	}
}

func BenchmarkShardSetSmall(b *testing.B) {
	benchmarkShardSet(2048, 2048, 512, b)
}

func BenchmarkShardSetMedium(b *testing.B) {
	benchmarkShardSet(2048, 2048, 1024, b)
}

func BenchmarkShardSetLarge(b *testing.B) {
	benchmarkShardSet(16384, 4096, 2048, b)
}

func benchmarkShardGet(recordCount, recordSize, dataSize int, b *testing.B) {
	var data []byte
	shard := NewShard(recordCount, recordSize)

	for i := 0; i < dataSize; i++ {
		data = append(data, 1)
	}

	index := shard.Set(data)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		shard.Get(index)
	}
}

func BenchmarkShardGetSmall(b *testing.B) {
	benchmarkShardGet(2048, 2048, 512, b)
}

func BenchmarkShardGetMedium(b *testing.B) {
	benchmarkShardGet(2048, 2048, 1024, b)
}

func BenchmarkShardGetLarge(b *testing.B) {
	benchmarkShardGet(16384, 4096, 2048, b)
}
