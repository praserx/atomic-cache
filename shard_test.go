package atomiccache

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestShardSimple(t *testing.T) {
	for _, c := range []struct {
		slotCount uint32
		slotSize  uint32
		in        []byte
		want      []byte
	}{
		{256, 2048, []byte{0}, []byte{0}},
		{256, 2048, []byte{0, 1, 2, 3, 4, 5}, []byte{0, 1, 2, 3, 4, 5}},
		{256, 1, []byte{0, 1, 2}, []byte{0}},
	} {
		shard := NewShard(c.slotCount, c.slotSize)
		index := shard.Set(c.in, 10*time.Minute)
		if !reflect.DeepEqual(shard.Get(index), c.want) {
			t.Errorf("%v != %v", shard.Get(index), c.want)
		}
	}
}

func TestShardFreeExpiration(t *testing.T) {
	var data []byte
	shard := NewShard(2048, 2048)

	for i := uint32(0); i < 2048; i++ {
		data = append(data, 1)
	}

	index := shard.Set(data, 500*time.Millisecond)

	time.Sleep(100 * time.Millisecond)

	if len(shard.Get(index)) == 0 {
		t.Errorf("Cache is empty, but expecting some data")
	}

	time.Sleep(500 * time.Millisecond)

	if len(shard.Get(index)) != 0 {
		t.Errorf("Cache is not empty, but expecting nothing")
	}
}

func benchmarkShardNew(slotCount, slotSize uint32, b *testing.B) {
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		NewShard(slotCount, slotSize)
	}
}

func BenchmarkShardNew512(b *testing.B) {
	benchmarkShardNew(512, 2048, b)
}

func BenchmarkShardNew1024(b *testing.B) {
	benchmarkShardNew(1024, 2048, b)
}

func BenchmarkShardNew2048(b *testing.B) {
	benchmarkShardNew(2048, 2048, b)
}

func benchmarkShardSet(slotCount, slotSize, dataSize uint32, b *testing.B) {
	b.ReportAllocs()

	var data []byte
	shard := NewShard(slotCount, slotSize)

	for i := uint32(0); i < dataSize; i++ {
		data = append(data, 1)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		index := shard.Set(data, 1*time.Minute)
		shard.Free(index)
	}
}

func BenchmarkShardSet512(b *testing.B) {
	benchmarkShardSet(2048, 2048, 512, b)
}

func BenchmarkShardSet1024(b *testing.B) {
	benchmarkShardSet(2048, 2048, 1024, b)
}

func BenchmarkShardSet2048(b *testing.B) {
	benchmarkShardSet(2048, 2048, 2048, b)
}

func benchmarkShardGet(slotCount, slotSize, dataSize uint32, b *testing.B) {
	b.ReportAllocs()

	var data []byte
	shard := NewShard(slotCount, slotSize)

	for i := uint32(0); i < dataSize; i++ {
		data = append(data, 1)
	}

	index := shard.Set(data, 1*time.Minute)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		shard.Get(index)
	}

	shard.Free(index)
}

func BenchmarkShardGet512(b *testing.B) {
	benchmarkShardGet(2048, 2048, 512, b)
}

func BenchmarkShardGet1024(b *testing.B) {
	benchmarkShardGet(2048, 2048, 1024, b)
}

func BenchmarkShardGet2048(b *testing.B) {
	benchmarkShardGet(2048, 2048, 2048, b)
}

func BenchmarkShardGet2048Concurrent(b *testing.B) {
	b.ReportAllocs()

	var data []byte
	var wg sync.WaitGroup

	shard := NewShard(2048, 2048)

	for i := uint32(0); i < 2048; i++ {
		data = append(data, 1)
	}

	index := shard.Set(data, 1*time.Minute)

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			shard.Get(index)
		}()
	}
	wg.Wait()
}
