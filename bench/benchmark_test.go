package bench

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/YuminosukeSato/pyproc/pkg/pyproc"
)

// BenchmarkWorker benchmarks single worker performance
func BenchmarkWorker(b *testing.B) {
	worker := createTestWorker(b, "/tmp/bench-worker.sock")
	defer worker.Stop()

	ctx := context.Background()
	req := map[string]interface{}{"value": 42}
	var resp map[string]interface{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := worker.Call(ctx, "echo", req, &resp); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPool benchmarks pool performance with various worker counts
func BenchmarkPool(b *testing.B) {
	workerCounts := []int{1, 2, 4, 8}

	for _, numWorkers := range workerCounts {
		b.Run(fmt.Sprintf("Workers-%d", numWorkers), func(b *testing.B) {
			pool := createTestPool(b, numWorkers, fmt.Sprintf("/tmp/bench-pool-%d", numWorkers))
			defer pool.Stop()

			ctx := context.Background()
			req := map[string]interface{}{"value": 42}
			var resp map[string]interface{}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := pool.Call(ctx, "echo", req, &resp); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkConcurrentRequests benchmarks concurrent request handling
func BenchmarkConcurrentRequests(b *testing.B) {
	concurrencyLevels := []int{10, 50, 100, 500}

	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrent-%d", concurrency), func(b *testing.B) {
			pool := createTestPool(b, 4, fmt.Sprintf("/tmp/bench-concurrent-%d", concurrency))
			defer pool.Stop()

			ctx := context.Background()
			req := map[string]interface{}{"value": 42}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				var resp map[string]interface{}
				for pb.Next() {
					if err := pool.Call(ctx, "echo", req, &resp); err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}

// BenchmarkPayloadSize benchmarks different payload sizes
func BenchmarkPayloadSize(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"1KB", 1024},
		{"10KB", 10 * 1024},
		{"100KB", 100 * 1024},
		{"1MB", 1024 * 1024},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			pool := createTestPool(b, 2, fmt.Sprintf("/tmp/bench-payload-%s", size.name))
			defer pool.Stop()

			ctx := context.Background()
			// Create payload of specified size
			data := make([]byte, size.size)
			for i := range data {
				data[i] = byte(i % 256)
			}
			req := map[string]interface{}{"data": data}
			var resp map[string]interface{}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := pool.Call(ctx, "echo", req, &resp); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkCodecs benchmarks different codec implementations
func BenchmarkCodecs(b *testing.B) {
	codecs := []string{"json", "msgpack"}

	for _, codecType := range codecs {
		b.Run(codecType, func(b *testing.B) {
			opts := pyproc.PoolOptions{
				NumWorkers:     2,
				PythonPath:     "python3",
				WorkerScript:   "../examples/basic/worker.py",
				SocketPrefix:   fmt.Sprintf("/tmp/bench-codec-%s", codecType),
				StartupTimeout: 10 * time.Second,
				CallTimeout:    5 * time.Second,
				CodecType:      codecType,
			}

			logger := pyproc.NewLogger(pyproc.LogLevelError)
			pool, err := pyproc.NewPool(opts, logger)
			if err != nil {
				b.Fatal(err)
			}

			ctx := context.Background()
			if err := pool.Start(ctx); err != nil {
				b.Fatal(err)
			}
			defer pool.Stop()

			req := map[string]interface{}{"value": 42}
			var resp map[string]interface{}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := pool.Call(ctx, "predict", req, &resp); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkTypedAPI benchmarks the generic typed API
func BenchmarkTypedAPI(b *testing.B) {
	type Request struct {
		Value int `json:"value"`
	}
	type Response struct {
		Result int `json:"result"`
	}

	pool := createTestPool(b, 2, "/tmp/bench-typed")
	defer pool.Stop()

	ctx := context.Background()
	req := Request{Value: 42}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pyproc.CallTyped[Request, Response](ctx, pool, "predict", req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLatencyPercentiles measures latency percentiles
func BenchmarkLatencyPercentiles(b *testing.B) {
	pool := createTestPool(b, 4, "/tmp/bench-latency")
	defer pool.Stop()

	ctx := context.Background()
	req := map[string]interface{}{"value": 42}
	var resp map[string]interface{}

	latencies := make([]time.Duration, 0, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		if err := pool.Call(ctx, "echo", req, &resp); err != nil {
			b.Fatal(err)
		}
		latencies = append(latencies, time.Since(start))
	}

	// Calculate percentiles
	p50 := calculatePercentile(latencies, 50)
	p95 := calculatePercentile(latencies, 95)
	p99 := calculatePercentile(latencies, 99)

	b.Logf("Latency - p50: %v, p95: %v, p99: %v", p50, p95, p99)

	// Verify performance requirements
	if p50 > 100*time.Microsecond {
		b.Errorf("p50 latency %v exceeds target of 100µs", p50)
	}
	if p99 > 500*time.Microsecond {
		b.Errorf("p99 latency %v exceeds target of 500µs", p99)
	}
}

// Helper functions

func createTestWorker(b *testing.B, socketPath string) *pyproc.Worker {
	b.Helper()

	opts := pyproc.WorkerOptions{
		PythonPath:     "python3",
		WorkerScript:   "../examples/basic/worker.py",
		SocketPath:     socketPath,
		StartupTimeout: 10 * time.Second,
		CallTimeout:    5 * time.Second,
	}

	logger := pyproc.NewLogger(pyproc.LogLevelError)
	worker, err := pyproc.NewWorker(opts, logger)
	if err != nil {
		b.Fatal(err)
	}

	if err := worker.Start(); err != nil {
		b.Fatal(err)
	}

	return worker
}

func createTestPool(b *testing.B, numWorkers int, socketPrefix string) *pyproc.Pool {
	b.Helper()

	opts := pyproc.PoolOptions{
		NumWorkers:     numWorkers,
		PythonPath:     "python3",
		WorkerScript:   "../examples/basic/worker.py",
		SocketPrefix:   socketPrefix,
		StartupTimeout: 10 * time.Second,
		CallTimeout:    5 * time.Second,
	}

	logger := pyproc.NewLogger(pyproc.LogLevelError)
	pool, err := pyproc.NewPool(opts, logger)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	if err := pool.Start(ctx); err != nil {
		b.Fatal(err)
	}

	return pool
}

func calculatePercentile(latencies []time.Duration, percentile float64) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	// Simple percentile calculation (not perfectly accurate but fast)
	index := int(float64(len(latencies)-1) * percentile / 100.0)
	if index < 0 {
		index = 0
	}
	if index >= len(latencies) {
		index = len(latencies) - 1
	}

	return latencies[index]
}
