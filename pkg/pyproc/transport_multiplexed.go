package pyproc

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/YuminosukeSato/pyproc/internal/framing"
	"github.com/YuminosukeSato/pyproc/internal/protocol"
)

// MultiplexedTransport implements Transport with request multiplexing
type MultiplexedTransport struct {
	config TransportConfig
	logger *Logger
	conn   net.Conn
	framer *framing.Framer

	// Request tracking
	requestID atomic.Uint64
	pending   map[uint64]*pendingRequest
	mu        sync.RWMutex

	// Connection state
	closed    atomic.Bool
	closeOnce sync.Once
	closeCh   chan struct{}

	// Reader goroutine
	readerWg sync.WaitGroup
}

// pendingRequest tracks an in-flight request
type pendingRequest struct {
	id         uint64
	responseCh chan *protocol.Response
	errCh      chan error
	timer      *time.Timer
}

// NewMultiplexedTransport creates a new multiplexed transport
func NewMultiplexedTransport(config TransportConfig, logger *Logger) (*MultiplexedTransport, error) {
	if config.Address == "" {
		return nil, fmt.Errorf("address is required for multiplexed transport")
	}

	transport := &MultiplexedTransport{
		config:  config,
		logger:  logger,
		pending: make(map[uint64]*pendingRequest),
		closeCh: make(chan struct{}),
	}

	// Connect to the socket
	if err := transport.connect(); err != nil {
		return nil, err
	}

	// Start the reader goroutine
	transport.readerWg.Add(1)
	go transport.readLoop()

	return transport, nil
}

// connect establishes the connection
func (t *MultiplexedTransport) connect() error {
	timeout := 5 * time.Second
	if timeoutVal, ok := t.config.Options["timeout"].(time.Duration); ok {
		timeout = timeoutVal
	}

	// Connect with timeout
	conn, err := net.DialTimeout("unix", t.config.Address, timeout)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", t.config.Address, err)
	}

	t.conn = conn
	t.framer = framing.NewEnhancedFramer(conn)

	t.logger.Debug("multiplexed transport connected", "address", t.config.Address)
	return nil
}

// readLoop continuously reads responses from the connection
func (t *MultiplexedTransport) readLoop() {
	defer t.readerWg.Done()

	for {
		select {
		case <-t.closeCh:
			return
		default:
		}

		// Read a frame
		frame, err := t.framer.ReadFrame()
		if err != nil {
			if t.closed.Load() {
				return // Expected on shutdown
			}
			t.logger.Error("failed to read frame", "error", err)
			t.handleReadError(err)
			return
		}

		// Parse response
		var resp protocol.Response
		if err := resp.Unmarshal(frame.Payload); err != nil {
			t.logger.Error("failed to unmarshal response", "error", err)
			continue
		}

		// Set the request ID from frame header
		resp.ID = frame.Header.RequestID

		// Find pending request
		t.mu.RLock()
		pending, ok := t.pending[resp.ID]
		t.mu.RUnlock()

		if !ok {
			t.logger.Warn("received response for unknown request", "id", resp.ID)
			continue
		}

		// Deliver response
		select {
		case pending.responseCh <- &resp:
			// Response delivered
		case <-pending.timer.C:
			// Request already timed out
		}

		// Clean up pending request
		t.mu.Lock()
		delete(t.pending, resp.ID)
		t.mu.Unlock()
		pending.timer.Stop()
	}
}

// handleReadError handles errors from the read loop
func (t *MultiplexedTransport) handleReadError(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Notify all pending requests of the error
	for id, pending := range t.pending {
		select {
		case pending.errCh <- fmt.Errorf("connection error: %w", err):
		default:
		}
		pending.timer.Stop()
		delete(t.pending, id)
	}

	// Close the transport
	t.closed.Store(true)
	close(t.closeCh)
}

// Call sends a request and receives a response
func (t *MultiplexedTransport) Call(ctx context.Context, req *protocol.Request) (*protocol.Response, error) {
	if t.closed.Load() {
		return nil, fmt.Errorf("transport is closed")
	}

	// Generate request ID
	requestID := t.requestID.Add(1)
	req.ID = requestID

	// Create pending request
	pending := &pendingRequest{
		id:         requestID,
		responseCh: make(chan *protocol.Response, 1),
		errCh:      make(chan error, 1),
	}

	// Set timeout
	timeout := 30 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}
	pending.timer = time.NewTimer(timeout)

	// Register pending request
	t.mu.Lock()
	t.pending[requestID] = pending
	t.mu.Unlock()

	// Clean up on exit
	defer func() {
		pending.timer.Stop()
		t.mu.Lock()
		delete(t.pending, requestID)
		t.mu.Unlock()
	}()

	// Marshal request
	reqData, err := req.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create and write frame
	frame := framing.NewFrame(requestID, reqData)
	if err := t.framer.WriteFrame(frame); err != nil {
		return nil, fmt.Errorf("failed to write frame: %w", err)
	}

	// Wait for response
	select {
	case resp := <-pending.responseCh:
		return resp, nil
	case err := <-pending.errCh:
		return nil, err
	case <-pending.timer.C:
		return nil, fmt.Errorf("request timeout after %v", timeout)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Close closes the transport
func (t *MultiplexedTransport) Close() error {
	var closeErr error

	t.closeOnce.Do(func() {
		t.closed.Store(true)
		close(t.closeCh)

		// Close connection
		if t.conn != nil {
			closeErr = t.conn.Close()
		}

		// Wait for reader to finish
		t.readerWg.Wait()

		// Clean up pending requests
		t.mu.Lock()
		for id, pending := range t.pending {
			select {
			case pending.errCh <- fmt.Errorf("transport closed"):
			default:
			}
			pending.timer.Stop()
			delete(t.pending, id)
		}
		t.mu.Unlock()
	})

	return closeErr
}

// IsHealthy checks if the transport is healthy
func (t *MultiplexedTransport) IsHealthy() bool {
	return !t.closed.Load() && t.conn != nil
}
