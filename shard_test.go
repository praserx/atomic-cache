package atomiccache

import (
	"reflect"
	"testing"
)

func TestShardSimple(t *testing.T) {
	for _, c := range []struct {
		recordCount uint32
		recordSize  uint32
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
		recordCount uint32
		recordSize  uint32
		in          []byte
		want        []byte
	}{
		{1024, 1024, []byte("test value"), []byte("test value")},
	} {
		var indexes []uint32
		shard := NewShard(c.recordCount, c.recordSize)
		
		for i := uint32(0); i < c.recordCount; i++ {
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

func benchmarkShardNew(recordCount, recordSize uint32, b *testing.B) {
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

func benchmarkShardSet(recordCount, recordSize, dataSize uint32, b *testing.B) {
	b.ReportAllocs()

	var data []byte
	shard := NewShard(recordCount, recordSize)

	for i := uint32(0); i < dataSize; i++ {
		data = append(data, 1)
	}

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

func benchmarkShardGet(recordCount, recordSize, dataSize uint32, b *testing.B) {
	b.ReportAllocs()

	var data []byte
	shard := NewShard(recordCount, recordSize)

	for i := uint32(0); i < dataSize; i++ {
		data = append(data, 1)
	}

	index := shard.Set(data)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		shard.Get(index)
	}

	shard.Free(index)
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
