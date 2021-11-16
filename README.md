# Atomic Cache

Atomic cache is Golang fast in-memory cache (it wants to be fast - if you want to help, go ahead). Cache using limited nubmer of shards with limited number of containing records. So the memory is limited, but the limit depends on you.

After cache initialization, only one shard is allocated. After that, if there is no left space in shard, new one is allocated. If shard is empty, memory is freed.

There is also support for record expiration. You can set expire time for every record in cache memory.

## Example usage

```go
// Initialize cache memory (ac == atomiccache)
cache := ac.New(OptionMaxRecords(512), OptionRecordSize(2048), OptionMaxShards(48))

// Store data in cache memory
cache.Set([]byte("key"), []byte("data"), 500*time.Millisecond)

// Get data from  cache memory
if _, err := cache.Get([]byte("key")); err != nil {
    fmt.Fprintf(os.Stderr, "Cache is empty, but expecting some data: %v", err)
    os.Exit(1)
}
```

## Benchmark

For this benchmark was created memory with following specs: `2048 bytes per record`, `2048 records per shard`, `128 shards (max)`. The 1024 bytes was set.

```
BenchmarkCacheNewMedium-12       	     314	   3804726 ns/op	22686418 B/op	   12402 allocs/op
BenchmarkCacheSetMedium-12       	 1845970	       664.2 ns/op	     129 B/op	       5 allocs/op
BenchmarkCacheGetMedium-12       	 5701435	       209.1 ns/op	      16 B/op	       1 allocs/op
```

*If you want do some special bencharking, go ahead.*

### AtomicCache vs. BigCache vs. FreeCache vs. HashicorpCache

**SET**
```
BenchmarkAtomicCacheSet-12       	 2402160	       519.6 ns/op	     153 B/op	       6 allocs/op
BenchmarkBigCacheSet-12          	 4021484	       376.7 ns/op	      66 B/op	       1 allocs/op
BenchmarkFreeCacheSet-12         	10443246	       100.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkHashicorpCacheSet-12    	 2490759	       498.4 ns/op	     150 B/op	       5 allocs/op
```

**GET**
```
BenchmarkAtomicCacheGet-12       	 6300744	       187.6 ns/op	      16 B/op	       1 allocs/op
BenchmarkBigCacheGet-12          	 3798052	       322.5 ns/op	      37 B/op	       2 allocs/op
BenchmarkFreeCacheGet-12         	 9000188	       127.0 ns/op	      24 B/op	       1 allocs/op
BenchmarkHashicorpCacheGet-12    	 3039728	       405.8 ns/op	       7 B/op	       0 allocs/op

```