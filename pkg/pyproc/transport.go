package pyproc

import (
	"context"
	"fmt"

	"github.com/YuminosukeSato/pyproc/internal/protocol"
)

// Transport defines the interface for communication with Python workers
// This abstraction allows for different transport mechanisms (UDS, gRPC, etc.)
type Transport interface {
	// Call sends a request and receives a response
	Call(ctx context.Context, req *protocol.Request) (*protocol.Response, error)

	// Close closes the transport connection
	Close() error

	// IsHealthy checks if the transport is healthy
	IsHealthy() bool
}

// TransportConfig defines configuration for transport layer
type TransportConfig struct {
	Type    string // "uds", "grpc-tcp", "grpc-uds"
	Address string // Socket path or network address
	Options map[string]interface{}
}

// NewTransport creates a new transport based on configuration
func NewTransport(config TransportConfig, logger *Logger) (Transport, error) {
	switch config.Type {
	case "uds", "":
		// Default to UDS for backward compatibility
		return NewUDSTransport(config, logger)
	case "multiplexed":
		// Multiplexed transport with request ID support
		return NewMultiplexedTransport(config, logger)
	case "grpc-tcp", "grpc-uds":
		return NewGRPCTransport(config, logger)
	default:
		return nil, fmt.Errorf("unknown transport type: %s", config.Type)
	}
}
