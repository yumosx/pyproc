package rpc_clients

import (
	"context"
	"fmt"
	"time"

	"github.com/YuminosukeSato/pyproc/pkg/pyproc"
)

// PyprocClient wraps the existing pyproc Pool to implement RPCClient interface
type PyprocClient struct {
	pool   *pyproc.Pool
	config pyproc.WorkerConfig
}

// NewPyprocClient creates a new pyproc client instance
func NewPyprocClient(pythonExec, workerScript string) *PyprocClient {
	return &PyprocClient{
		config: pyproc.WorkerConfig{
			PythonExec:   pythonExec,
			WorkerScript: workerScript,
			StartTimeout: 5 * time.Second,
		},
	}
}

// Connect establishes connection to the pyproc worker
func (c *PyprocClient) Connect(udsPath string) error {
	c.config.SocketPath = udsPath
	c.config.ID = fmt.Sprintf("pyproc-bench-%d", time.Now().Unix())

	// Create a pool with a single worker for fair comparison
	pool, err := pyproc.NewPool(pyproc.PoolOptions{
		Config: pyproc.PoolConfig{
			Workers:     1,
			MaxInFlight: 10,
		},
		WorkerConfig: c.config,
	}, nil)

	if err != nil {
		return fmt.Errorf("failed to create pyproc pool: %w", err)
	}

	ctx := context.Background()
	if err := pool.Start(ctx); err != nil {
		return fmt.Errorf("failed to start pyproc pool: %w", err)
	}

	// Give it time to initialize
	time.Sleep(100 * time.Millisecond)

	c.pool = pool
	return nil
}

// Call invokes a method on the pyproc worker
func (c *PyprocClient) Call(ctx context.Context, method string, args interface{}, reply interface{}) error {
	if c.pool == nil {
		return fmt.Errorf("pool not connected")
	}

	return c.pool.Call(ctx, method, args, reply)
}

// Close shuts down the pyproc worker
func (c *PyprocClient) Close() error {
	if c.pool == nil {
		return nil
	}

	ctx := context.Background()
	return c.pool.Shutdown(ctx)
}

// Name returns the protocol identifier
func (c *PyprocClient) Name() string {
	return "pyproc"
}
