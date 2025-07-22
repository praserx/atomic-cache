package atomiccache

import (
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
		if !reflect.DeepEqual(shardSectionID, c.want) {
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
		if err := cache.Set(string(byte(i)), c.in, 0); err != nil {
			t.Errorf("Set error: %s", err.Error())
		}

		value, err := cache.Get(string(byte(i)))
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
		if err := cache.Set("X", []byte(bigString), 0); err != nil {
			t.Errorf("Set error: %s", err.Error())
		}

		value, err := cache.Get("X")
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
		for i := 0; i < 1000; i++ {
			if err := cache.Set(strconv.Itoa(i), c.in, 0); err != nil {
				t.Errorf("Set error: %s", err.Error())
			}
		}

		value, err := cache.Get("0")
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
		recordSizeSmall  int
		recordSizeMedium int
		recordSizeLarge  int
		in               []byte
		want             []byte
	}{
		{1, 2, 3, []byte{0, 1, 2, 3, 4, 5}, []byte{0}},
	} {
		cache := New(OptionRecordSizeSmall(c.recordSizeSmall), OptionRecordSizeMedium(c.recordSizeMedium), OptionRecordSizeLarge(c.recordSizeLarge))
		if err := cache.Set(strconv.Itoa(i), c.in, 0); err == nil {
			t.Errorf("Expecting error 'errDataLimit'")
		}
	}
}

func TestCacheFreeAfterExpiration(t *testing.T) {
	cache := New(OptionGcStarter(1))

	cache.Set("key", []byte("data"), 500*time.Millisecond)
	time.Sleep(100 * time.Millisecond)

	if _, err := cache.Get("key"); err != nil {
		t.Errorf("Cache is empty, but expecting some data")
	}
	time.Sleep(500 * time.Millisecond)

	cache.Set("key2", []byte("data"), 500*time.Millisecond)
	time.Sleep(100 * time.Millisecond)

	if _, err := cache.Get("key"); err == nil {
		t.Errorf("Cache is not empty, but expecting nothing")
	}
}

func TestCacheKeepTTL(t *testing.T) {
	cache := New()
	key := "keep-ttl-key"
	data1 := []byte("first")
	data2 := []byte("second")
	expire := 2 * time.Second

	// Set initial value with expiration
	if err := cache.Set(key, data1, expire); err != nil {
		t.Fatalf("Set error: %s", err)
	}
	val, ok := cache.lookup[key]
	if !ok {
		t.Fatalf("Key not found after Set")
	}
	origExp := val.Expiration

	// Update value with KeepTTL
	if err := cache.Set(key, data2, KeepTTL); err != nil {
		t.Fatalf("Set error (KeepTTL): %s", err)
	}
	val2, ok := cache.lookup[key]
	if !ok {
		t.Fatalf("Key not found after Set with KeepTTL")
	}
	if !val2.Expiration.Equal(origExp) {
		t.Errorf("Expiration changed: got %v, want %v", val2.Expiration, origExp)
	}
	// Value should be updated
	got, err := cache.Get(key)
	if err != nil {
		t.Fatalf("Get error: %s", err)
	}
	if !reflect.DeepEqual(got, data2) {
		t.Errorf("Value not updated: got %v, want %v", got, data2)
	}
}

func TestCacheExists(t *testing.T) {
	cache := New()
	key := "exists-key"
	data := []byte("exists-data")

	// Should not exist before set
	if cache.Exists(key) {
		t.Errorf("Exists returned true for unset key")
	}

	// Set and check exists
	if err := cache.Set(key, data, 10*time.Second); err != nil {
		t.Fatalf("Set error: %s", err)
	}
	if !cache.Exists(key) {
		t.Errorf("Exists returned false for set key")
	}

	// Delete and check exists
	if err := cache.Delete(key); err != nil {
		t.Fatalf("Delete error: %s", err)
	}
	if cache.Exists(key) {
		t.Errorf("Exists returned true after Delete")
	}

	// Never-set key
	if cache.Exists("never-existed") {
		t.Errorf("Exists returned true for never-set key")
	}
}

func TestCacheDelete(t *testing.T) {
	cache := New()
	key := "del-key"
	data := []byte("to-delete")

	// Set and then delete
	if err := cache.Set(key, data, 0); err != nil {
		t.Fatalf("Set error: %s", err)
	}
	if err := cache.Delete(key); err != nil {
		t.Errorf("Delete error: %s", err)
	}
	// Should not be able to get deleted key
	if _, err := cache.Get(key); err == nil {
		t.Errorf("Expected error on Get after Delete, got nil")
	}

	// Deleting again should return ErrNotFound
	if err := cache.Delete(key); err != ErrNotFound {
		t.Errorf("Expected ErrNotFound on double Delete, got %v", err)
	}

	// Deleting a never-set key should return ErrNotFound
	if err := cache.Delete("never-existed"); err != ErrNotFound {
		t.Errorf("Expected ErrNotFound for never-set key, got %v", err)
	}
}

func benchmarkCacheNew(recordCount int, b *testing.B) {
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		New(OptionMaxRecords(recordCount))
	}
}

func benchmarkCacheSet(recordCount, dataSize int, b *testing.B) {
	var data []byte
	keys := generateKeys(32000, 64)

	for i := 0; i < dataSize; i++ {
		data = append(data, 1)
	}

	cache := New(OptionMaxRecords(recordCount))

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		cache.Set(keys[n%len(keys)], data, 0)
	}
}

func benchmarkCacheGet(recordCount, dataSize int, b *testing.B) {
	var data []byte

	for i := 0; i < dataSize; i++ {
		data = append(data, 1)
	}

	cache := New(OptionMaxRecords(recordCount))
	cache.Set("0", data, 0)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		cache.Get("0")
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

func BenchmarkAtomicCacheSet(b *testing.B) {
	cache := newAtoCache()
	keys := generateKeys(32000, 64)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		cache.Set(keys[n%len(keys)], []byte("Testing data input"), time.Duration(10*time.Minute))
	}
}

func BenchmarkBigCacheSet(b *testing.B) {
	cache := newBigCache()
	keys := generateKeys(32000, 64)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		cache.Set(keys[n%len(keys)], []byte("Testing data input"))
	}
}

func BenchmarkFreeCacheSet(b *testing.B) {
	cache := newFreCache()
	keys := generateKeys(32000, 64)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		cache.Set([]byte(keys[n%len(keys)]), []byte("Testing data input"), 600)
	}
}

func BenchmarkHashicorpCacheSet(b *testing.B) {
	cache := newHasCache()
	keys := generateKeys(32000, 64)

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		cache.Add(keys[n%len(keys)], []byte("Testing data input"))
	}
}

func BenchmarkAtomicCacheGet(b *testing.B) {
	cache := newAtoCache()
	keys := generateKeys(32000, 64)

	for i := 0; i < 32000; i++ {
		cache.Set(keys[i%len(keys)], []byte("Testing data input"), time.Duration(10*time.Minute))
	}

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		cache.Get(keys[n%len(keys)])
	}
}

func BenchmarkBigCacheGet(b *testing.B) {
	cache := newBigCache()
	keys := generateKeys(32000, 64)

	for i := 0; i < 32000; i++ {
		cache.Set(keys[i%len(keys)], []byte("Testing data input"))
	}

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		cache.Get(keys[n%len(keys)])
	}
}

func BenchmarkFreeCacheGet(b *testing.B) {
	cache := newFreCache()
	keys := generateKeys(32000, 64)

	for i := 0; i < 32000; i++ {
		cache.Set([]byte(keys[i%len(keys)]), []byte("Testing data input"), 600)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		cache.Get([]byte(keys[n%len(keys)]))
	}
}

func BenchmarkHashicorpCacheGet(b *testing.B) {
	cache := newHasCache()
	keys := generateKeys(32000, 64)

	for i := 0; i < 32000; i++ {
		cache.Add(keys[i%len(keys)], []byte("Testing data input"))
	}

	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		cache.Get(keys[n%len(keys)])
	}
}

func newAtoCache() *AtomicCache {
	var options []Option
	options = append(options, OptionRecordSizeSmall(256))
	options = append(options, OptionRecordSizeMedium(1024))
	options = append(options, OptionRecordSizeLarge(8128))
	options = append(options, OptionMaxRecords(4096))
	options = append(options, OptionMaxShardsSmall(512))
	options = append(options, OptionMaxShardsMedium(256))
	options = append(options, OptionMaxShardsLarge(64))
	cache := New(options...)
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

func generateKeys(cnt, length int) (keys []string) {
	var source = rand.NewSource(time.Now().UnixNano())
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	for count := 0; count < cnt; count++ {
		b := make([]byte, length)

		for i := range b {
			b[i] = charset[source.Int63()%int64(len(charset))]
		}

		keys = append(keys, string(b))
	}

	return
}
