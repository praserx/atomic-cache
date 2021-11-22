# Atomic Cache

Atomic cache is Golang fast in-memory cache (it wants to be fast - if you want to help, go ahead). Cache using limited number of shards with limited number of containing records. So the memory is limited, but the limit depends on you.

After cache initialization, only one shard is allocated. After that, if there is no left space in shard, new one is allocated. If shard is empty, memory is freed.

There is also support for record expiration. You can set expire time for every record in cache memory.

## Configuration

| Option           | Type   | Description                                                            |
| ---------------- | ------ | ---------------------------------------------------------------------- |
| RecordSizeSmall  | int    | Size of byte array used for memory allocation at small shard section.  |
| RecordSizeMedium | int    | Size of byte array used for memory allocation at medium shard section. |
| RecordSizeLarge  | int    | Size of byte array used for memory allocation at large shard section.  |
| MaxRecords       | int    | Maximum records per shard.                                             |
| MaxShardsSmall   | int    | Maximum small shards which can be allocated in cache memory.           |
| MaxShardsMedium  | int    | Maximum medium shards which can be allocated in cache memory.          |
| MaxShardsLarge   | int    | Maximum large shards which can be allocated in cache memory.           |
| GcStarter        | uint32 | Garbage collector starter (run garbage collection every X sets).       |
	 
## Example usage

```go
// Initialize cache memory (ac == atomiccache)
cache := ac.New(OptionMaxRecords(512), OptionRecordSize(2048), OptionMaxShards(48))

// Store data in cache memory - key, data, record valid time
cache.Set("key", []byte("data"), 500*time.Millisecond)

// Get data from  cache memory
if _, err := cache.Get("key"); err != nil {
    fmt.Fprintf(os.Stderr, "Cache is empty, but expecting some data: %v", err)
    os.Exit(1)
}
```

## Benchmark

For this benchmark was created memory with following specs: `1024 bytes per record`, `4096 records per shard`, `256 shards (max)`. The 1024 bytes was set.

```
BenchmarkCacheNewMedium-12       	     291	   3670372 ns/op	22776481 B/op	   12408 allocs/op
BenchmarkCacheSetMedium-12       	 1928548	       620.3 ns/op	      63 B/op	       1 allocs/op
BenchmarkCacheGetMedium-12       	16707145	        69.87 ns/op	       0 B/op	       0 allocs/op
```

*If you want do some special bencharking, go ahead.*

### AtomicCache vs. BigCache vs. FreeCache vs. HashicorpCache

**SET**
```
BenchmarkAtomicCacheSet-12       	 2921170	       413.0 ns/op	      55 B/op	       2 allocs/op
BenchmarkBigCacheSet-12          	 3448020	       345.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkFreeCacheSet-12         	 4777364	       217.2 ns/op	      65 B/op	       1 allocs/op
BenchmarkHashicorpCacheSet-12    	 6208528	       202.2 ns/op	      65 B/op	       3 allocs/op
```

**GET**
```
BenchmarkAtomicCacheGet-12       	 9697010	       121.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkBigCacheGet-12          	 4031352	       295.3 ns/op	      88 B/op	       2 allocs/op
BenchmarkFreeCacheGet-12         	 4813386	       276.8 ns/op	      88 B/op	       2 allocs/op
BenchmarkHashicorpCacheGet-12    	11071472	       107.4 ns/op	      16 B/op	       1 allocs/op
```