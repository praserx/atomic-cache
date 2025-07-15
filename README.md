# Atomic Cache

Atomic Cache is a high-performance, in-memory caching library for Go, designed for low-latency data retrieval in performance-critical systems. It uses a sharded architecture with configurable memory limits, supporting efficient storage and retrieval of byte slices with per-record expiration.

Atomic Cache is ideal for applications that require predictable memory usage, fast access times, and robust cache eviction strategies.

The library is production-ready, leverages only the Go standard library and select well-maintained dependencies, and is easy to integrate into any Go project.

## Installation

```go
go get github.com/praserx/atomic-cache
```

## Configuration

| Option                | Type    | Description                                                            |
| ----------------------| ------- | ---------------------------------------------------------------------- |
| RecordSizeSmall       | int     | Size of byte array used for memory allocation at small shard section.  |
| RecordSizeMedium      | int     | Size of byte array used for memory allocation at medium shard section. |
| RecordSizeLarge       | int     | Size of byte array used for memory allocation at large shard section.  |
| MaxRecords            | int     | Maximum records per shard.                                             |
| MaxShardsSmall        | int     | Maximum small shards which can be allocated in cache memory.           |
| MaxShardsMedium       | int     | Maximum medium shards which can be allocated in cache memory.          |
| MaxShardsLarge        | int     | Maximum large shards which can be allocated in cache memory.           |
| GcStarter             | uint32  | Garbage collector starter (run garbage collection every X sets).       |

### Option Functions

All options are set using functional options when calling `atomiccache.New`. Available option functions:

- `atomiccache.OptionRecordSizeSmall(int)`
- `atomiccache.OptionRecordSizeMedium(int)`
- `atomiccache.OptionRecordSizeLarge(int)`
- `atomiccache.OptionMaxRecords(int)`
- `atomiccache.OptionMaxShardsSmall(int)`
- `atomiccache.OptionMaxShardsMedium(int)`
- `atomiccache.OptionMaxShardsLarge(int)`
- `atomiccache.OptionGcStarter(uint32)`
     
## Example usage

```go
package main

import (
    "github.com/praserx/atomic-cache"
    "fmt"
    "os"
    "time"
)

func main() {
    // Initialize cache memory with custom options
    cache := atomiccache.New(
        atomiccache.OptionMaxRecords(512),
        atomiccache.OptionRecordSizeSmall(2048),
        atomiccache.OptionMaxShardsSmall(48),
    )

    // Store data in cache memory - key, data, record valid time
    if err := atomiccache.Set("key", []byte("data"), 500*time.Millisecond); err != nil {
        fmt.Fprintf(os.Stderr, "Set failed: %v\n", err)
        os.Exit(1)
    }

    // Get data from cache memory
    data, err := atomiccache.Get("key")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Cache miss: %v\n", err)
        os.Exit(1)
    }
    fmt.Printf("Got: %s\n", data)
}
```

## Benchmark

To run benchmarks, use:

```sh
go test -bench=. -benchmem
```

For this benchmark, memory was created with the following specs: `1024 bytes per record`, `4096 records per shard`, `256 shards (max)`.

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

## License

This project is licensed under the terms of the MIT License. See the [LICENSE](LICENSE) file for details.