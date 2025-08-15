# Benchmarks

Performance benchmarks for pyproc Pool operations.

## Overview

This directory contains performance benchmarks to measure the efficiency and scalability of pyproc's worker pool implementation.

## Running Benchmarks

### Run all benchmarks
```bash
go test -bench=. -benchmem ./bench
```

### Run specific benchmark
```bash
go test -bench=BenchmarkPoolCall -benchmem ./bench
```

### Run with custom parameters
```bash
# Run benchmarks 10 times
go test -bench=. -benchmem -benchtime=10s ./bench

# Generate CPU profile
go test -bench=. -benchmem -cpuprofile=cpu.prof ./bench
```

## Current Benchmarks

### BenchmarkPoolCall
Measures the performance of single `Pool.Call` operations with various payload sizes.

**Metrics:**
- Operations per second
- Latency (ns/op)
- Memory allocation per operation
- Number of allocations

### BenchmarkPoolConcurrent
Tests the pool's ability to handle concurrent requests under load.

**Parameters tested:**
- Different worker counts (1, 2, 4, 8)
- Various concurrency levels
- Different payload sizes

## Interpreting Results

Example output:
```
BenchmarkPoolCall-8         1000    1045875 ns/op    2048 B/op    42 allocs/op
```

- `BenchmarkPoolCall-8`: Benchmark name with GOMAXPROCS value
- `1000`: Number of iterations
- `1045875 ns/op`: Nanoseconds per operation (~1ms)
- `2048 B/op`: Bytes allocated per operation
- `42 allocs/op`: Number of allocations per operation

## Performance Goals

Target performance metrics for pyproc:
- **Latency**: < 1ms for small payloads (< 1KB)
- **Throughput**: 1-5k RPS per worker
- **Memory**: < 10KB per request
- **Scalability**: Linear scaling up to 8 workers

## Profiling

To analyze performance bottlenecks:

```bash
# Generate profiles
go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof ./bench

# Analyze CPU profile
go tool pprof cpu.prof

# Analyze memory profile
go tool pprof mem.prof
```

## Contributing

When adding new benchmarks:
1. Follow the naming convention `BenchmarkXxx`
2. Include relevant metrics in comments
3. Test with different input sizes and concurrency levels
4. Update this README with benchmark descriptions