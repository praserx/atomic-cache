# Atomic Cache

Atomic Cache is a high-performance, in-memory caching library for Go, designed for low-latency data retrieval in performance-critical systems. It uses a sharded architecture with configurable memory limits, supporting efficient storage and retrieval of byte slices with per-record expiration.

Atomic Cache is ideal for applications that require predictable memory usage, fast access times, and robust cache eviction strategies.

The library is production-ready, leverages only the Go standard library and select well-maintained dependencies, and is easy to integrate into any Go project.

## Installation

```go
go get github.com/praserx/atomic-cache/v2
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
goos: linux
goarch: amd64
pkg: github.com/praserx/atomic-cache/v2
cpu: Intel(R) Core(TM) i7-10850H CPU @ 2.70GHz
BenchmarkCacheNewMedium-12       	     288	   4109240 ns/op	22750002 B/op	   12405 allocs/op
BenchmarkCacheSetMedium-12       	 4499152	       269.8 ns/op	      16 B/op	       0 allocs/op
BenchmarkCacheGetMedium-12       	19747963	        59.72 ns/op	       0 B/op	       0 allocs/op
```

*If you want do some special bencharking, go ahead.*

### AtomicCache vs. BigCache vs. FreeCache vs. HashicorpCache

**SET**
```
goos: linux
goarch: amd64
pkg: github.com/praserx/atomic-cache/v2
cpu: Intel(R) Core(TM) i7-10850H CPU @ 2.70GHz
BenchmarkAtomicCacheSet-12       	 5755452	       195.3 ns/op	      27 B/op	       1 allocs/op
BenchmarkBigCacheSet-12          	 4290684	       286.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkFreeCacheSet-12         	 5806412	       199.3 ns/op	      65 B/op	       1 allocs/op
BenchmarkHashicorpCacheSet-12    	 6333306	       170.0 ns/op	      65 B/op	       3 allocs/op
```

**GET**
```
goos: linux
goarch: amd64
pkg: github.com/praserx/atomic-cache/v2
cpu: Intel(R) Core(TM) i7-10850H CPU @ 2.70GHz
BenchmarkAtomicCacheGet-12       	13004460	        97.27 ns/op	       0 B/op	       0 allocs/op
BenchmarkBigCacheGet-12          	 4403041	       272.5 ns/op	      88 B/op	       2 allocs/op
BenchmarkFreeCacheGet-12         	 5586747	       231.9 ns/op	      88 B/op	       2 allocs/op
BenchmarkHashicorpCacheGet-12    	11445339	        99.70 ns/op	      16 B/op	       1 allocs/op
```

## License

This project is licensed under the terms of the MIT License. See the [LICENSE](LICENSE) file for details.