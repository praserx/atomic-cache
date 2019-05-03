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

For this benchmark was created memory with following specs: `2048 bytes per record`, `2048 records per shard`, `128 shards (max)`. The 2048 bytes was set.

```
BenchmarkCacheNewMedium-4    	    1000	   1012982 ns/op	 4416184 B/op	    4139 allocs/op
BenchmarkCacheSetMedium-4    	 5000000	       327 ns/op	      16 B/op	       0 allocs/op
BenchmarkCacheGetMedium-4    	30000000	        48.5 ns/op	       0 B/op	       0 allocs/op
```

*If you want do some special bencharking, go ahead.*
