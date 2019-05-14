package atomiccache

import (
	"reflect"
	"time"
	"encoding/binary"
	"testing"
)

func TestCacheSimple(t *testing.T) {
	for i, c := range []struct {
		recordCount uint32
		recordSize  uint32
		shardCount  uint32
		in          []byte
		want        []byte
	}{
		{4096, 4096, 128, []byte{0}, []byte{0}},
		{4096, 4096, 128, []byte{0, 1, 2, 3, 4, 5}, []byte{0, 1, 2, 3, 4, 5}},
	} {
		cache := New(OptionMaxRecords(c.recordCount), OptionRecordSize(c.recordSize), OptionMaxShards(c.shardCount))
		if err := cache.Set([]byte{byte(i)}, c.in, 0); err != nil {
			t.Errorf("Set error: %s", err.Error())
		}

		value, err := cache.Get([]byte{byte(i)})
		if err != nil {
			t.Errorf("Get error: %s", err.Error())
		}

		if !reflect.DeepEqual(value, c.want) {
			t.Errorf("%v != %v", value, c.want)
		}
	}
}

func TestCacheIntermediate(t *testing.T) {
	for _, c := range []struct {
		recordCount uint32
		recordSize  uint32
		shardCount  uint32
		in          []byte
		want        []byte
	}{
		{1024, 1024, 64, []byte("test value"), []byte("test value")},
	} {
		cache := New(OptionMaxRecords(c.recordCount), OptionRecordSize(c.recordSize), OptionMaxShards(c.shardCount))
		for i := uint32(0); i < 1000; i++ {
			bs := make([]byte, 4)
    		binary.LittleEndian.PutUint32(bs, i)
			if err := cache.Set(bs, c.in, 0); err != nil {
				t.Errorf("Set error: %s", err.Error())
			}
		}

		bs := make([]byte, 4)
    	binary.LittleEndian.PutUint32(bs, 0)
		value, err := cache.Get(bs)
		if err != nil {
			t.Errorf("Get error: %s", err.Error())
		}

		if !reflect.DeepEqual(value, c.want) {
			t.Errorf("%v != %v", value, c.want)
		}
	}
}

func TestCacheDataError(t *testing.T) {
	for i, c := range []struct {
		recordCount uint32
		recordSize  uint32
		shardCount  uint32
		in          []byte
		want        []byte
	}{
		{4096, 1, 2048, []byte{0, 1, 2}, []byte{0}},
	} {
		cache := New(OptionMaxRecords(c.recordCount), OptionRecordSize(c.recordSize), OptionMaxShards(c.shardCount))
		if err := cache.Set([]byte{byte(i)}, c.in, 0); err == nil {
			t.Errorf("Expecting error 'errDataLimit'")
		}
	}
}

func TestCacheFreeAfterExpiration(t *testing.T) {
	cache := New(OptionMaxRecords(512), OptionRecordSize(2048), OptionMaxShards(48), OptionGcStarter(1))

	cache.Set([]byte("key"), []byte("data"), 500*time.Millisecond)
	time.Sleep(100 * time.Millisecond)

	if _, err := cache.Get([]byte("key")); err != nil {
		t.Errorf("Cache is empty, but expecting some data")
	}
	time.Sleep(500 * time.Millisecond)

	cache.Set([]byte("key2"), []byte("data"), 500*time.Millisecond)
	time.Sleep(100 * time.Millisecond)

	if _, err := cache.Get([]byte("key")); err == nil {
		t.Errorf("Cache is not empty, but expecting nothing")
	}
}

func benchmarkCacheNew(recordCount, recordSize, shardCount uint32, b *testing.B) {
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		New(OptionMaxRecords(recordCount), OptionRecordSize(recordSize), OptionMaxShards(shardCount))
	}
}

func BenchmarkCacheNewSmall(b *testing.B) {
	benchmarkCacheNew(512, 2048, 48, b)
}

func BenchmarkCacheNewMedium(b *testing.B) {
	benchmarkCacheNew(2048, 2048, 128, b)
}

func BenchmarkCacheNewLarge(b *testing.B) {
	benchmarkCacheNew(16384, 4096, 256, b)
}

func benchmarkCacheSet(recordCount, recordSize, shardCount, dataSize uint32, b *testing.B) {
	var data []byte
	cache := New(OptionMaxRecords(recordCount), OptionRecordSize(recordSize), OptionMaxShards(shardCount))

	for i := uint32(0); i < dataSize; i++ {
		data = append(data, 1)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		cache.Set([]byte{byte(n)}, data, 0)
	}
}

func BenchmarkCacheSetSmall(b *testing.B) {
	benchmarkCacheSet(512, 2048, 48, 1024, b)
}

func BenchmarkCacheSetMedium(b *testing.B) {
	benchmarkCacheSet(2048, 2048, 128, 1024, b)
}

func BenchmarkCacheSetLarge(b *testing.B) {
	benchmarkCacheSet(16384, 4096, 256, 1024, b)
}

func benchmarkCacheGet(recordCount, recordSize, shardCount, dataSize uint32, b *testing.B) {
	var data []byte
	cache := New(OptionMaxRecords(recordCount), OptionRecordSize(recordSize), OptionMaxShards(shardCount))

	for i := uint32(0); i < dataSize; i++ {
		data = append(data, 1)
	}

	cache.Set([]byte{byte(1)}, data, 0)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		cache.Get([]byte{byte(1)})
	}
}

func BenchmarkCacheGetSmall(b *testing.B) {
	benchmarkCacheGet(512, 2048, 48, 1024, b)
}

func BenchmarkCacheGetMedium(b *testing.B) {
	benchmarkCacheGet(2048, 2048, 128, 1024, b)
}

func BenchmarkCacheGetLarge(b *testing.B) {
	benchmarkCacheGet(16384, 4096, 256, 1024, b)
}

func benchmarkAdvanced(recordCount, recordSize, shardCount, dataSize uint32, b *testing.B) {
	var data []byte
	cache := New(OptionMaxRecords(recordCount), OptionRecordSize(recordSize), OptionMaxShards(shardCount))

	for i := uint32(0); i < dataSize; i++ {
		data = append(data, 1)
	}

	cache.Set([]byte{byte(1)}, data, 0)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		cache.Get([]byte{byte(1)})
	}
}