package bench

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/YuminosukeSato/pyproc/pkg/pyproc"
)

// BenchmarkSingleWorker benchmarks a single worker without pool
func BenchmarkSingleWorker(b *testing.B) {
	cfg := pyproc.WorkerConfig{
		ID:           "bench-worker",
		SocketPath:   "/tmp/bench-single.sock",
		PythonExec:   "python3",
		WorkerScript: "../examples/basic/worker.py",
		StartTimeout: 5 * time.Second,
	}

	worker := pyproc.NewWorker(cfg, nil)
	ctx := context.Background()

	if err := worker.Start(ctx); err != nil {
		b.Fatalf("failed to start worker: %v", err)
	}
	defer worker.Stop()

	// Wait for worker to be ready
	time.Sleep(500 * time.Millisecond)

	// Connect to worker
	conn, err := pyproc.ConnectToWorker(cfg.SocketPath, 5*time.Second)
	if err != nil {
		b.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	input := map[string]interface{}{"value": 42}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var output map[string]interface{}
		if err := conn.Call(ctx, "predict", input, &output); err != nil {
			b.Fatalf("call failed: %v", err)
		}
	}
}

// BenchmarkPool benchmarks the pool with multiple workers
func BenchmarkPool(b *testing.B) {
	for _, numWorkers := range []int{1, 2, 4, 8} {
		b.Run(fmt.Sprintf("workers=%d", numWorkers), func(b *testing.B) {
			opts := pyproc.PoolOptions{
				Config: pyproc.PoolConfig{
					Workers:     numWorkers,
					MaxInFlight: 10,
				},
				WorkerConfig: pyproc.WorkerConfig{
					SocketPath:   "/tmp/bench-pool.sock",
					PythonExec:   "python3",
					WorkerScript: "../examples/basic/worker.py",
					StartTimeout: 5 * time.Second,
				},
			}

			pool, err := pyproc.NewPool(opts, nil)
			if err != nil {
				b.Fatalf("failed to create pool: %v", err)
			}

			ctx := context.Background()
			if err := pool.Start(ctx); err != nil {
				b.Fatalf("failed to start pool: %v", err)
			}
			defer pool.Shutdown(ctx)

			// Wait for pool to be ready
			time.Sleep(500 * time.Millisecond)

			input := map[string]interface{}{"value": 42}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var output map[string]interface{}
				if err := pool.Call(ctx, "predict", input, &output); err != nil {
					b.Fatalf("call failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkPoolParallel benchmarks parallel requests to the pool
func BenchmarkPoolParallel(b *testing.B) {
	for _, numWorkers := range []int{2, 4, 8} {
		b.Run(fmt.Sprintf("workers=%d", numWorkers), func(b *testing.B) {
			opts := pyproc.PoolOptions{
				Config: pyproc.PoolConfig{
					Workers:     numWorkers,
					MaxInFlight: 20,
				},
				WorkerConfig: pyproc.WorkerConfig{
					SocketPath:   "/tmp/bench-pool-parallel.sock",
					PythonExec:   "python3",
					WorkerScript: "../examples/basic/worker.py",
					StartTimeout: 5 * time.Second,
				},
			}

			pool, err := pyproc.NewPool(opts, nil)
			if err != nil {
				b.Fatalf("failed to create pool: %v", err)
			}

			ctx := context.Background()
			if err := pool.Start(ctx); err != nil {
				b.Fatalf("failed to start pool: %v", err)
			}
			defer pool.Shutdown(ctx)

			// Wait for pool to be ready
			time.Sleep(500 * time.Millisecond)

			input := map[string]interface{}{"value": 42}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				var output map[string]interface{}
				for pb.Next() {
					if err := pool.Call(ctx, "predict", input, &output); err != nil {
						b.Fatalf("call failed: %v", err)
					}
				}
			})
		})
	}
}

// BenchmarkPoolThroughput measures throughput with various payloads
func BenchmarkPoolThroughput(b *testing.B) {
	testCases := []struct {
		name    string
		method  string
		payload interface{}
	}{
		{
			name:    "small_payload",
			method:  "predict",
			payload: map[string]interface{}{"value": 42},
		},
		{
			name:   "medium_payload",
			method: "process_batch",
			payload: map[string]interface{}{
				"values": []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			},
		},
		{
			name:   "large_payload",
			method: "compute_stats",
			payload: map[string]interface{}{
				"numbers": generateNumbers(100),
			},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			opts := pyproc.PoolOptions{
				Config: pyproc.PoolConfig{
					Workers:     4,
					MaxInFlight: 10,
				},
				WorkerConfig: pyproc.WorkerConfig{
					SocketPath:   fmt.Sprintf("/tmp/bench-throughput-%s.sock", tc.name),
					PythonExec:   "python3",
					WorkerScript: "../examples/basic/worker.py",
					StartTimeout: 5 * time.Second,
				},
			}

			pool, err := pyproc.NewPool(opts, nil)
			if err != nil {
				b.Fatalf("failed to create pool: %v", err)
			}

			ctx := context.Background()
			if err := pool.Start(ctx); err != nil {
				b.Fatalf("failed to start pool: %v", err)
			}
			defer pool.Shutdown(ctx)

			// Wait for pool to be ready
			time.Sleep(500 * time.Millisecond)

			b.ResetTimer()

			start := time.Now()
			var wg sync.WaitGroup
			errors := make(chan error, b.N)

			for i := 0; i < b.N; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					var output map[string]interface{}
					if err := pool.Call(ctx, tc.method, tc.payload, &output); err != nil {
						errors <- err
					}
				}()
			}

			wg.Wait()
			close(errors)

			elapsed := time.Since(start)

			// Check for errors
			for err := range errors {
				if err != nil {
					b.Fatalf("call failed: %v", err)
				}
			}

			// Report throughput
			throughput := float64(b.N) / elapsed.Seconds()
			b.ReportMetric(throughput, "req/s")
			b.ReportMetric(float64(elapsed.Nanoseconds())/float64(b.N)/1000, "μs/op")
		})
	}
}

// BenchmarkPoolLatency measures latency percentiles
func BenchmarkPoolLatency(b *testing.B) {
	opts := pyproc.PoolOptions{
		Config: pyproc.PoolConfig{
			Workers:     4,
			MaxInFlight: 10,
		},
		WorkerConfig: pyproc.WorkerConfig{
			SocketPath:   "/tmp/bench-latency.sock",
			PythonExec:   "python3",
			WorkerScript: "../examples/basic/worker.py",
			StartTimeout: 5 * time.Second,
		},
	}

	pool, err := pyproc.NewPool(opts, nil)
	if err != nil {
		b.Fatalf("failed to create pool: %v", err)
	}

	ctx := context.Background()
	if err := pool.Start(ctx); err != nil {
		b.Fatalf("failed to start pool: %v", err)
	}
	defer pool.Shutdown(ctx)

	// Wait for pool to be ready
	time.Sleep(500 * time.Millisecond)

	input := map[string]interface{}{"value": 42}
	latencies := make([]time.Duration, 0, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		var output map[string]interface{}
		if err := pool.Call(ctx, "predict", input, &output); err != nil {
			b.Fatalf("call failed: %v", err)
		}
		latencies = append(latencies, time.Since(start))
	}

	// Calculate percentiles
	p50 := percentile(latencies, 50)
	p95 := percentile(latencies, 95)
	p99 := percentile(latencies, 99)

	b.ReportMetric(float64(p50.Microseconds()), "p50_μs")
	b.ReportMetric(float64(p95.Microseconds()), "p95_μs")
	b.ReportMetric(float64(p99.Microseconds()), "p99_μs")
}

// Helper functions

func generateNumbers(n int) []int {
	numbers := make([]int, n)
	for i := 0; i < n; i++ {
		numbers[i] = i + 1
	}
	return numbers
}

func percentile(latencies []time.Duration, p float64) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	// Simple percentile calculation (not exact but good enough for benchmarks)
	index := int(float64(len(latencies)) * p / 100)
	if index >= len(latencies) {
		index = len(latencies) - 1
	}
	return latencies[index]
}
