package atomiccache

import (
	"encoding/binary"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"

	big "github.com/allegro/bigcache"
	fre "github.com/coocood/freecache"
	has "github.com/hashicorp/golang-lru"
)

func TestCacheFuncGetShardsSectionBySize(t *testing.T) {
	for _, c := range []struct {
		in   int
		want int
	}{
		{256, 1}, {512, 1}, {2048, 2}, {8127, 3},
	} {
		cache := New()

		_, shardSectionID := cache.getShardsSectionBySize(c.in)
		if !reflect.DeepEqual(shardSectionID, uint8(c.want)) {
			t.Errorf("%v != %v", c.in, c.want)
		}
	}
}

func TestCacheSimple(t *testing.T) {
	for i, c := range []struct {
		in   []byte
		want []byte
	}{
		{[]byte{0}, []byte{0}},
		{[]byte{0, 1, 2, 3, 4, 5}, []byte{0, 1, 2, 3, 4, 5}},
	} {
		cache := New()
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

func TestCacheDiffSizeData(t *testing.T) {
	for _, c := range []struct {
		in int
	}{
		{256}, {512}, {513}, {2047}, {2048}, {2049}, {4096}, {8127}, {8128},
	} {
		var bigString string
		for x := 0; x < c.in; x++ {
			bigString += "x"
		}

		cache := New()
		if err := cache.Set([]byte{byte(0)}, []byte(bigString), 0); err != nil {
			t.Errorf("Set error: %s", err.Error())
		}

		value, err := cache.Get([]byte{byte(0)})
		if err != nil {
			t.Errorf("Get error: %s", err.Error())
		}

		if !reflect.DeepEqual(string(value), bigString) {
			t.Errorf("%v != %v", string(value), bigString)
		}
	}
}

func TestCacheIntermediate(t *testing.T) {
	for _, c := range []struct {
		in   []byte
		want []byte
	}{
		{[]byte("test value"), []byte("test value")},
	} {
		cache := New()
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
		recordSizeSmall  uint32
		recordSizeMedium uint32
		recordSizeLarge  uint32
		in               []byte
		want             []byte
	}{
		{1, 2, 3, []byte{0, 1, 2, 3, 4, 5}, []byte{0}},
	} {
		cache := New(OptionRecordSizeSmall(c.recordSizeSmall), OptionRecordSizeMedium(c.recordSizeMedium), OptionRecordSizeLarge(c.recordSizeLarge))
		if err := cache.Set([]byte{byte(i)}, c.in, 0); err == nil {
			t.Errorf("Expecting error 'errDataLimit'")
		}
	}
}

func TestCacheFreeAfterExpiration(t *testing.T) {
	cache := New(OptionGcStarter(1))

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

func benchmarkCacheNew(recordCount uint32, b *testing.B) {
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		New(OptionMaxRecords(recordCount))
	}
}

func benchmarkCacheSet(recordCount, dataSize uint32, b *testing.B) {
	var data []byte
	cache := New(OptionMaxRecords(recordCount))

	for i := uint32(0); i < dataSize; i++ {
		data = append(data, 1)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		cache.Set([]byte{byte(n)}, data, 0)
	}
}

func benchmarkCacheGet(recordCount, dataSize uint32, b *testing.B) {
	var data []byte
	cache := New(OptionMaxRecords(recordCount))

	for i := uint32(0); i < dataSize; i++ {
		data = append(data, 1)
	}

	for i := 0; i < 128000; i++ {
		cache.Set([]byte{byte(i)}, data, 0)
	}

	b.ReportAllocs()
	b.ResetTimer()

	rand.Seed(42)

	for n := 0; n < b.N; n++ {
		cache.Get([]byte{byte(rand.Intn(128000))})
	}
}

func benchmarkAdvanced(recordCount, dataSize uint32, b *testing.B) {
	var data []byte
	cache := New(OptionMaxRecords(recordCount))

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

func BenchmarkCacheNewSmall(b *testing.B) {
	benchmarkCacheNew(512, b)
}

func BenchmarkCacheNewMedium(b *testing.B) {
	benchmarkCacheNew(2048, b)
}

func BenchmarkCacheNewLarge(b *testing.B) {
	benchmarkCacheNew(16384, b)
}

func BenchmarkCacheSetSmall(b *testing.B) {
	benchmarkCacheSet(512, 1024, b)
}

func BenchmarkCacheSetMedium(b *testing.B) {
	benchmarkCacheSet(2048, 1024, b)
}

func BenchmarkCacheSetLarge(b *testing.B) {
	benchmarkCacheSet(16384, 1024, b)
}

func BenchmarkCacheGetSmall(b *testing.B) {
	benchmarkCacheGet(512, 1024, b)
}

func BenchmarkCacheGetMedium(b *testing.B) {
	benchmarkCacheGet(2048, 1024, b)
}

func BenchmarkCacheGetLarge(b *testing.B) {
	benchmarkCacheGet(16384, 1024, b)
}

const testRecordsCnt = 256000

func BenchmarkAtomicCacheSet(b *testing.B) {
	cache := newAtoCache()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		cache.Set([]byte{byte(n)}, []byte("Testing data input"), time.Duration(10*time.Minute))
	}
}

func BenchmarkBigCacheSet(b *testing.B) {
	cache := newBigCache()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		cache.Set(strconv.Itoa(n), []byte("Testing data input"))
	}
}

func BenchmarkFreeCacheSet(b *testing.B) {
	cache := newFreCache()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		cache.Set([]byte{byte(n)}, []byte("Testing data input"), 600)
	}
}

func BenchmarkHashicorpCacheSet(b *testing.B) {
	cache := newHasCache()

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		cache.Add(n, []byte("Testing data input"))
	}
}

func BenchmarkAtomicCacheGet(b *testing.B) {
	cache := newAtoCache()

	for n := 0; n < testRecordsCnt; n++ {
		cache.Set([]byte{byte(n)}, []byte("Testing data input"), time.Duration(10*time.Minute))
	}

	b.ReportAllocs()
	b.ResetTimer()

	rand.Seed(42)

	for n := 0; n < b.N; n++ {
		cache.Get([]byte{byte(rand.Intn(testRecordsCnt))})
	}
}

func BenchmarkBigCacheGet(b *testing.B) {
	cache := newBigCache()

	for n := 0; n < testRecordsCnt; n++ {
		cache.Set(strconv.Itoa(n), []byte("Testing data input"))
	}

	b.ReportAllocs()
	b.ResetTimer()

	rand.Seed(42)

	for n := 0; n < b.N; n++ {
		cache.Get(strconv.Itoa(rand.Intn(testRecordsCnt)))
	}
}

func BenchmarkFreeCacheGet(b *testing.B) {
	cache := newFreCache()

	for n := 0; n < testRecordsCnt; n++ {
		cache.Set([]byte{byte(n)}, []byte("Testing data input"), 600)
	}

	b.ReportAllocs()
	b.ResetTimer()

	rand.Seed(42)

	for n := 0; n < b.N; n++ {
		cache.Get([]byte{byte(rand.Intn(testRecordsCnt))})
	}
}

func BenchmarkHashicorpCacheGet(b *testing.B) {
	cache := newHasCache()

	for n := 0; n < testRecordsCnt; n++ {
		cache.Add(n, []byte("Testing data input"))
	}

	b.ReportAllocs()
	b.ResetTimer()

	rand.Seed(42)

	for n := 0; n < b.N; n++ {
		cache.Get(rand.Intn(testRecordsCnt))
	}
}

func newAtoCache() *AtomicCache {
	cache := New()
	return cache
}

func newBigCache() *big.BigCache {
	cache, _ := big.NewBigCache(big.Config{
		Shards:             128,
		LifeWindow:         10 * time.Minute,
		MaxEntriesInWindow: 1000 * 10 * 60,
		MaxEntrySize:       2048,
		HardMaxCacheSize:   0,
	})

	return cache
}

func newFreCache() *fre.Cache {
	cache := fre.NewCache(100 * 1024 * 1024)
	return cache
}

func newHasCache() *has.Cache {
	cache, _ := has.New(262144)
	return cache
}
