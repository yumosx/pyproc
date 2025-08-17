// Package rpc_clients provides a common interface for different RPC protocols
// to enable fair performance comparison over Unix Domain Sockets.
package rpc_clients

import (
	"context"
	"time"
)

// RPCClient defines the common interface for all RPC protocol implementations.
// Each implementation must provide connection management and call functionality.
type RPCClient interface {
	// Connect establishes a connection to the RPC server via Unix Domain Socket
	Connect(udsPath string) error

	// Call invokes a remote method with given arguments and stores the reply
	Call(ctx context.Context, method string, args interface{}, reply interface{}) error

	// Close terminates the connection and cleans up resources
	Close() error

	// Name returns the protocol name for identification in benchmarks
	Name() string
}

// BenchmarkConfig holds configuration for benchmark testing
type BenchmarkConfig struct {
	UDSPath     string        // Unix Domain Socket path
	Timeout     time.Duration // Request timeout
	MaxRetries  int           // Maximum retry attempts
	WarmupCalls int           // Number of warmup calls before measurement
}

// TestPayload represents different payload sizes for benchmarking
type TestPayload struct {
	Size   string      // "small", "medium", "large"
	Method string      // RPC method to call
	Data   interface{} // Actual payload data
}

// Small payload (~64 bytes) - Simple operation
func SmallPayload() TestPayload {
	return TestPayload{
		Size:   "small",
		Method: "predict",
		Data: map[string]interface{}{
			"value": 42,
		},
	}
}

// Medium payload (~2KB) - Typical API request
func MediumPayload() TestPayload {
	values := make([]int, 100)
	for i := range values {
		values[i] = i + 1
	}

	return TestPayload{
		Size:   "medium",
		Method: "process_batch",
		Data: map[string]interface{}{
			"values": values,
			"metadata": map[string]interface{}{
				"user_id":   "test-user-123",
				"timestamp": time.Now().Unix(),
				"version":   "1.0.0",
			},
		},
	}
}

// Large payload (~1MB) - Data transfer scenario
func LargePayload() TestPayload {
	// Generate ~1MB of data
	numbers := make([]int, 100000)
	for i := range numbers {
		numbers[i] = i % 1000
	}

	return TestPayload{
		Size:   "large",
		Method: "compute_stats",
		Data: map[string]interface{}{
			"numbers": numbers,
			"options": map[string]interface{}{
				"compute_variance": true,
				"compute_std_dev":  true,
				"compute_median":   true,
			},
		},
	}
}

// BenchmarkResult stores the results of a benchmark run
type BenchmarkResult struct {
	Protocol    string        // Protocol name
	PayloadSize string        // Payload size category
	Latency     time.Duration // Average latency
	Throughput  float64       // Requests per second
	ErrorRate   float64       // Percentage of failed requests
	CPUUsage    float64       // Average CPU usage percentage
	MemoryUsage int64         // Memory usage in bytes
}

