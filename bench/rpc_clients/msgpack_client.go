package rpc_clients

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"github.com/vmihailenco/msgpack/v5"
)

// MsgpackRPCClient implements MessagePack-RPC protocol over Unix Domain Socket
type MsgpackRPCClient struct {
	conn      net.Conn
	udsPath   string
	requestID uint32
	mu        sync.Mutex
}

// MsgpackRequest represents a MessagePack-RPC request
// Format: [type, msgid, method, params]
type MsgpackRequest struct {
	Type   uint8       // 0 for request
	MsgID  uint32      // Message ID
	Method string      // Method name
	Params interface{} // Parameters
}

// MsgpackResponse represents a MessagePack-RPC response
// Format: [type, msgid, error, result]
type MsgpackResponse struct {
	Type   uint8       // 1 for response
	MsgID  uint32      // Message ID
	Error  interface{} // Error if any
	Result interface{} // Result value
}

// NewMsgpackRPCClient creates a new MessagePack-RPC client
func NewMsgpackRPCClient() *MsgpackRPCClient {
	return &MsgpackRPCClient{}
}

// Connect establishes connection to MessagePack-RPC server via UDS
func (c *MsgpackRPCClient) Connect(udsPath string) error {
	conn, err := net.Dial("unix", udsPath)
	if err != nil {
		return fmt.Errorf("failed to connect to MessagePack-RPC server: %w", err)
	}

	c.conn = conn
	c.udsPath = udsPath
	return nil
}

// Call invokes a MessagePack-RPC method
func (c *MsgpackRPCClient) Call(ctx context.Context, method string, args interface{}, reply interface{}) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	// Generate unique request ID
	msgID := atomic.AddUint32(&c.requestID, 1)

	// Create MessagePack-RPC request array
	request := []interface{}{
		uint8(0), // Request type
		msgID,    // Message ID
		method,   // Method name
		args,     // Parameters
	}

	// Encode request
	var buf bytes.Buffer
	encoder := msgpack.NewEncoder(&buf)
	if err := encoder.Encode(request); err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}

	// Send request with length prefix (4 bytes)
	reqData := buf.Bytes()
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(reqData)))

	c.mu.Lock()
	if _, err := c.conn.Write(lenBuf); err != nil {
		c.mu.Unlock()
		return fmt.Errorf("failed to send length prefix: %w", err)
	}
	if _, err := c.conn.Write(reqData); err != nil {
		c.mu.Unlock()
		return fmt.Errorf("failed to send request: %w", err)
	}

	// Read response length
	if _, err := c.conn.Read(lenBuf); err != nil {
		c.mu.Unlock()
		return fmt.Errorf("failed to read response length: %w", err)
	}
	respLen := binary.BigEndian.Uint32(lenBuf)

	// Read response data
	respData := make([]byte, respLen)
	if _, err := c.conn.Read(respData); err != nil {
		c.mu.Unlock()
		return fmt.Errorf("failed to read response: %w", err)
	}
	c.mu.Unlock()

	// Decode response
	decoder := msgpack.NewDecoder(bytes.NewReader(respData))
	var response []interface{}
	if err := decoder.Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Validate response format
	if len(response) != 4 {
		return fmt.Errorf("invalid response format")
	}

	// Check message type (should be 1 for response)
	if respType, ok := response[0].(uint8); !ok || respType != 1 {
		return fmt.Errorf("invalid response type")
	}

	// Check message ID matches
	if respID, ok := response[1].(uint32); !ok || respID != msgID {
		return fmt.Errorf("message ID mismatch")
	}

	// Check for error
	if response[2] != nil {
		return fmt.Errorf("MessagePack-RPC error: %v", response[2])
	}

	// Extract result
	if reply != nil && response[3] != nil {
		// Convert response[3] to the expected reply type
		// This is simplified for benchmark purposes
		if m, ok := reply.(*map[string]interface{}); ok {
			if result, ok := response[3].(map[string]interface{}); ok {
				*m = result
			} else {
				*m = map[string]interface{}{
					"result": response[3],
				}
			}
		}
	}

	return nil
}

// Close terminates the connection
func (c *MsgpackRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Name returns the protocol identifier
func (c *MsgpackRPCClient) Name() string {
	return "msgpack-rpc"
}
