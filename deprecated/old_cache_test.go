package atomiccache

// import (
// 	"reflect"
// 	// "sync"
// 	"testing"
// )

// func TestCacheSimple(t *testing.T) {
// 	for i, c := range []struct {
// 		slotCount  uint32
// 		slotSize   uint32
// 		shardCount uint32
// 		in         []byte
// 		want       []byte
// 	}{
// 		{4096, 4096, 2048, []byte{0}, []byte{0}},
// 		{4096, 4096, 2048, []byte{0, 1, 2, 3, 4, 5}, []byte{0, 1, 2, 3, 4, 5}},
// 		{4096, 1, 2048, []byte{0, 1, 2}, []byte{0}},
// 	} {
// 		cache := New(OptionMaxRecords(c.slotCount), OptionRecordSize(c.slotSize), OptionMaxShards(c.shardCount))
// 		if err := cache.Set([]byte{byte(i)}, c.in, 0); err != nil {
// 			t.Errorf("Set error: %s", err.Error())
// 		}

// 		value, err := cache.Get([]byte{byte(i)})
// 		if err != nil {
// 			t.Errorf("Get error: %s", err.Error())
// 		}

// 		if !reflect.DeepEqual(value, c.want) {
// 			t.Errorf("%v != %v", value, c.want)
// 		}
// 	}
// }

// func TestCacheFreeExpiration(t *testing.T) {
// 	// var data []byte
// 	// shard := NewShard(2048, 2048)

// 	// for i := uint32(0); i < 2048; i++ {
// 	// 	data = append(data, 1)
// 	// }

// 	// index := shard.Set(data, 500*time.Millisecond)

// 	// time.Sleep(100 * time.Millisecond)

// 	// if len(shard.Get(index)) == 0 {
// 	// 	t.Errorf("Cache is empty, but expecting some data")
// 	// }

// 	// time.Sleep(500 * time.Millisecond)

// 	// if len(shard.Get(index)) != 0 {
// 	// 	t.Errorf("Cache is not empty, but expecting nothing")
// 	// }
// }

// // benchmarkCacheNew is generic cache initialization benchmark.
// func benchmarkCacheNew(slotCount, slotSize, shardCount uint32, b *testing.B) {
// 	b.ReportAllocs()

// 	for n := 0; n < b.N; n++ {
// 		New(OptionMaxRecords(slotCount), OptionRecordSize(slotSize), OptionMaxShards(shardCount))
// 	}
// }

// func BenchmarkCacheNewSmall(b *testing.B) {
// 	benchmarkCacheNew(512, 2048, 48, b)
// }

// func BenchmarkCacheNewMedium(b *testing.B) {
// 	benchmarkCacheNew(2048, 2048, 128, b)
// }

// func BenchmarkCacheNewLarge(b *testing.B) {
// 	benchmarkCacheNew(8192, 4096, 2048, b)
// }

// func benchmarkCacheSet(slotCount, slotSize, shardCount, dataSize uint32, b *testing.B) {
// 	b.ReportAllocs()

// 	var data []byte
// 	cache := New(OptionMaxRecords(slotCount), OptionRecordSize(slotSize), OptionMaxShards(shardCount))

// 	for i := uint32(0); i < dataSize; i++ {
// 		data = append(data, 1)
// 	}

// 	b.ResetTimer()

// 	for n := 0; n < b.N; n++ {
// 		cache.Set([]byte{byte(n)}, data, 0)
// 	}
// }

// func BenchmarkCacheSetSmall(b *testing.B) {
// 	benchmarkCacheSet(512, 2048, 48, 1024, b)
// }

// func BenchmarkCacheSetMedium(b *testing.B) {
// 	benchmarkCacheSet(2048, 2048, 128, 1024, b)
// }

// func BenchmarkCacheSetLarge(b *testing.B) {
// 	benchmarkCacheSet(8192, 4096, 2048, 1024, b)
// }

// func benchmarkCacheGet(slotCount, slotSize, shardCount, dataSize uint32, b *testing.B) {
// 	b.ReportAllocs()

// 	var data []byte
// 	cache := New(OptionMaxRecords(slotCount), OptionRecordSize(slotSize), OptionMaxShards(shardCount))

// 	for i := uint32(0); i < dataSize; i++ {
// 		data = append(data, 1)
// 	}

// 	cache.Set([]byte{byte(1)}, data, 0)

// 	b.ResetTimer()

// 	for n := 0; n < b.N; n++ {
// 		cache.Get([]byte{byte(1)})
// 	}
// }

// func BenchmarkCacheGetSmall(b *testing.B) {
// 	benchmarkCacheGet(512, 2048, 48, 1024, b)
// }

// func BenchmarkCacheGetMedium(b *testing.B) {
// 	benchmarkCacheGet(2048, 2048, 128, 1024, b)
// }

// func BenchmarkCacheGetLarge(b *testing.B) {
// 	benchmarkCacheGet(8192, 4096, 2048, 1024, b)
// }

// func BenchmarkCacheSetPerformance01(b *testing.B) {
// 	benchmarkCacheSet(8192, 4096, 1024, 1024, b)
// }

// func BenchmarkCacheSetPerformance02(b *testing.B) {
// 	benchmarkCacheSet(512, 4096, 4096, 1024, b)
// }

// func BenchmarkCacheSetPerformance03(b *testing.B) {
// 	benchmarkCacheSet(512, 4096, 8192, 1024, b)
// }

// func BenchmarkCacheSetPerformance04(b *testing.B) {
// 	benchmarkCacheSet(256, 4096, 8192, 1024, b)
// }

// func BenchmarkCacheSetPerformance05(b *testing.B) {
// 	benchmarkCacheSet(128, 4096, 8192, 1024, b)
// }

// func BenchmarkCacheSetPerformance06(b *testing.B) {
// 	benchmarkCacheSet(64, 4096, 8192, 1024, b)
// }

// func BenchmarkCacheSetPerformance07(b *testing.B) {
// 	benchmarkCacheSet(64, 4096, 4096, 1024, b)
// }

// func BenchmarkCacheSetPerformance08(b *testing.B) {
// 	benchmarkCacheSet(64, 4096, 2048, 1024, b)
// }

// func BenchmarkCacheSetPerformance09(b *testing.B) {
// 	benchmarkCacheSet(64, 4096, 1024, 1024, b)
// }
