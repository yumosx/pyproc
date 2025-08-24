package pyproc

import (
	"context"
	"fmt"
)

// CallTyped is a type-safe wrapper for Pool.Call using Go generics
// TIn is the input type, TOut is the output type
func CallTyped[TIn any, TOut any](ctx context.Context, pool *Pool, method string, input TIn) (TOut, error) {
	var output TOut

	// Call the underlying pool method
	err := pool.Call(ctx, method, input, &output)
	if err != nil {
		return output, fmt.Errorf("call %s failed: %w", method, err)
	}

	return output, nil
}

// CallTypedWithTransport is a type-safe wrapper for PoolWithTransport.Call using Go generics
func CallTypedWithTransport[TIn any, TOut any](ctx context.Context, pool *PoolWithTransport, method string, input TIn) (TOut, error) {
	var output TOut

	// Call the underlying pool method
	err := pool.Call(ctx, method, input, &output)
	if err != nil {
		return output, fmt.Errorf("call %s failed: %w", method, err)
	}

	return output, nil
}

// TypedPool provides a type-safe wrapper around Pool with predefined types
type TypedPool[TIn any, TOut any] struct {
	pool *Pool
}

// NewTypedPool creates a new typed pool wrapper
func NewTypedPool[TIn any, TOut any](pool *Pool) *TypedPool[TIn, TOut] {
	return &TypedPool[TIn, TOut]{
		pool: pool,
	}
}

// Call executes a method with type safety
func (tp *TypedPool[TIn, TOut]) Call(ctx context.Context, method string, input TIn) (TOut, error) {
	return CallTyped[TIn, TOut](ctx, tp.pool, method, input)
}

// Start starts all workers in the pool
func (tp *TypedPool[TIn, TOut]) Start(ctx context.Context) error {
	return tp.pool.Start(ctx)
}

// Shutdown gracefully shuts down the pool
func (tp *TypedPool[TIn, TOut]) Shutdown(ctx context.Context) error {
	return tp.pool.Shutdown(ctx)
}

// Health returns the health status of the pool
func (tp *TypedPool[TIn, TOut]) Health() HealthStatus {
	return tp.pool.Health()
}

// TypedWorkerClient provides a type-safe client for specific worker methods
type TypedWorkerClient[TIn any, TOut any] struct {
	pool   *Pool
	method string
}

// NewTypedWorkerClient creates a client for a specific worker method
func NewTypedWorkerClient[TIn any, TOut any](pool *Pool, method string) *TypedWorkerClient[TIn, TOut] {
	return &TypedWorkerClient[TIn, TOut]{
		pool:   pool,
		method: method,
	}
}

// Call executes the predefined method with type safety
func (tc *TypedWorkerClient[TIn, TOut]) Call(ctx context.Context, input TIn) (TOut, error) {
	return CallTyped[TIn, TOut](ctx, tc.pool, tc.method, input)
}

// BatchCall executes multiple requests in parallel
func (tc *TypedWorkerClient[TIn, TOut]) BatchCall(ctx context.Context, inputs []TIn) ([]TOut, []error) {
	results := make([]TOut, len(inputs))
	errors := make([]error, len(inputs))

	// Use goroutines for parallel execution
	type result struct {
		index  int
		output TOut
		err    error
	}

	resultCh := make(chan result, len(inputs))

	for i, input := range inputs {
		go func(idx int, in TIn) {
			out, err := tc.Call(ctx, in)
			resultCh <- result{index: idx, output: out, err: err}
		}(i, input)
	}

	// Collect results
	for i := 0; i < len(inputs); i++ {
		res := <-resultCh
		results[res.index] = res.output
		errors[res.index] = res.err
	}

	return results, errors
}

// Example usage types for common patterns

// PredictRequest represents a prediction request
type PredictRequest struct {
	Value float64 `json:"value"`
}

// PredictResponse represents a prediction response
type PredictResponse struct {
	Result float64 `json:"result"`
}

// TransformRequest represents a text transformation request
type TransformRequest struct {
	Text string `json:"text"`
}

// TransformResponse represents a text transformation response
type TransformResponse struct {
	TransformedText string `json:"transformed_text"`
	WordCount       int    `json:"word_count"`
}

// BatchRequest represents a batch processing request
type BatchRequest struct {
	Items []map[string]interface{} `json:"items"`
}

// BatchResponse represents a batch processing response
type BatchResponse struct {
	Results []map[string]interface{} `json:"results"`
	Count   int                      `json:"count"`
}

// StatsRequest represents a statistics computation request
type StatsRequest struct {
	Numbers []float64 `json:"numbers"`
}

// StatsResponse represents a statistics computation response
type StatsResponse struct {
	Mean   float64 `json:"mean"`
	Median float64 `json:"median"`
	StdDev float64 `json:"std_dev"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
}
