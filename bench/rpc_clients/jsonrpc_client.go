package rpc_clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
)

// JSONRPCClient implements JSON-RPC 2.0 protocol over Unix Domain Socket
type JSONRPCClient struct {
	conn      net.Conn
	udsPath   string
	requestID uint64
	mu        sync.Mutex
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      uint64      `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
	ID      uint64          `json:"id"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewJSONRPCClient creates a new JSON-RPC client
func NewJSONRPCClient() *JSONRPCClient {
	return &JSONRPCClient{}
}

// Connect establishes connection to JSON-RPC server via UDS
func (c *JSONRPCClient) Connect(udsPath string) error {
	conn, err := net.Dial("unix", udsPath)
	if err != nil {
		return fmt.Errorf("failed to connect to JSON-RPC server: %w", err)
	}

	c.conn = conn
	c.udsPath = udsPath
	return nil
}

// Call invokes a JSON-RPC method
func (c *JSONRPCClient) Call(ctx context.Context, method string, args interface{}, reply interface{}) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	// Generate unique request ID
	id := atomic.AddUint64(&c.requestID, 1)

	// Create JSON-RPC request
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  args,
		ID:      id,
	}

	// Send request
	c.mu.Lock()
	encoder := json.NewEncoder(c.conn)
	if err := encoder.Encode(request); err != nil {
		c.mu.Unlock()
		return fmt.Errorf("failed to send request: %w", err)
	}

	// Receive response
	decoder := json.NewDecoder(c.conn)
	var response JSONRPCResponse
	if err := decoder.Decode(&response); err != nil {
		c.mu.Unlock()
		return fmt.Errorf("failed to receive response: %w", err)
	}
	c.mu.Unlock()

	// Check for error
	if response.Error != nil {
		return fmt.Errorf("JSON-RPC error %d: %s", response.Error.Code, response.Error.Message)
	}

	// Unmarshal result
	if reply != nil && response.Result != nil {
		if err := json.Unmarshal(response.Result, reply); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	return nil
}

// Close terminates the connection
func (c *JSONRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Name returns the protocol identifier
func (c *JSONRPCClient) Name() string {
	return "json-rpc"
}

