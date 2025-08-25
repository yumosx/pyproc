package pyproc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	pyprocv1 "github.com/YuminosukeSato/pyproc/api/v1"
	"github.com/YuminosukeSato/pyproc/internal/protocol"
)

// GRPCTransport implements Transport using gRPC
type GRPCTransport struct {
	config  TransportConfig
	logger  *Logger
	conn    *grpc.ClientConn
	client  pyprocv1.PyProcServiceClient
	mu      sync.RWMutex
	closed  bool
	healthy bool
}

// NewGRPCTransport creates a new gRPC transport
func NewGRPCTransport(config TransportConfig, logger *Logger) (*GRPCTransport, error) {
	// gRPC transport is not fully implemented yet
	return nil, fmt.Errorf("gRPC transport is not yet implemented")
	
	// Original implementation commented out for future use:
	/*
	if config.Address == "" {
		return nil, fmt.Errorf("address is required for gRPC transport")
	}

	transport := &GRPCTransport{
		config:  config,
		logger:  logger,
		healthy: false,
	}

	// Connect to gRPC server
	if err := transport.connect(); err != nil {
		return nil, err
	}

	return transport, nil
	*/
}

// connect establishes the gRPC connection
func (t *GRPCTransport) connect() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Close existing connection if any
	if t.conn != nil {
		_ = t.conn.Close()
	}

	// Configure gRPC options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	// Determine target based on transport type
	var target string
	switch t.config.Type {
	case "grpc-tcp":
		target = t.config.Address
	case "grpc-uds":
		target = "unix://" + t.config.Address
	default:
		return fmt.Errorf("unsupported gRPC transport type: %s", t.config.Type)
	}

	// grpc.NewClient doesn't use context for initial connection
	// The connection is established lazily on first RPC
	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", target, err)
	}

	t.conn = conn
	t.client = pyprocv1.NewPyProcServiceClient(conn)
	t.healthy = true

	t.logger.Debug("gRPC transport connected", "address", t.config.Address, "type", t.config.Type)
	return nil
}

// Call sends a request and receives a response via gRPC
func (t *GRPCTransport) Call(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	t.mu.RLock()
	client := t.client
	closed := t.closed
	t.mu.RUnlock()

	if closed {
		return nil, fmt.Errorf("transport is closed")
	}

	if client == nil {
		return nil, fmt.Errorf("gRPC client not initialized")
	}

	// Create gRPC request using the already marshaled Body
	grpcReq := &pyprocv1.CallRequest{
		Id:     req.ID,
		Method: req.Method,
		Input:  req.Body,
	}

	// Make gRPC call
	grpcResp, err := client.Call(ctx, grpcReq)
	if err != nil {
		t.mu.Lock()
		t.healthy = false
		t.mu.Unlock()
		return nil, fmt.Errorf("gRPC call failed: %w", err)
	}

	// Convert gRPC response to protocol.Response
	resp := &protocol.Response{
		ID:       grpcResp.Id,
		OK:       grpcResp.Ok,
		Body:     grpcResp.Body,
		ErrorMsg: grpcResp.ErrorMessage,
	}

	t.mu.Lock()
	t.healthy = true
	t.mu.Unlock()

	return resp, nil
}

// Close closes the gRPC connection
func (t *GRPCTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}

	t.closed = true
	t.healthy = false

	if t.conn != nil {
		err := t.conn.Close()
		t.conn = nil
		t.client = nil
		return err
	}

	return nil
}

// IsHealthy checks if the transport is healthy
func (t *GRPCTransport) IsHealthy() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.closed || t.conn == nil || t.client == nil {
		return false
	}

	// Optionally perform a health check RPC
	if healthCheckEnabled, ok := t.config.Options["health_check"].(bool); ok && healthCheckEnabled {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		_, err := t.client.HealthCheck(ctx, &pyprocv1.HealthCheckRequest{})
		if err != nil {
			return false
		}
	}

	return t.healthy
}
