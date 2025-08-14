package framing

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	// DefaultMaxFrameSize is the default maximum frame size (10MB)
	DefaultMaxFrameSize = 10 * 1024 * 1024
)

// Framer handles framing of messages over a stream
type Framer struct {
	rw          io.ReadWriter
	maxFrameSize int
}

// NewFramer creates a new framer with default max frame size
func NewFramer(rw io.ReadWriter) *Framer {
	return &Framer{
		rw:          rw,
		maxFrameSize: DefaultMaxFrameSize,
	}
}

// NewFramerWithMaxSize creates a new framer with specified max frame size
func NewFramerWithMaxSize(rw io.ReadWriter, maxSize int) *Framer {
	return &Framer{
		rw:          rw,
		maxFrameSize: maxSize,
	}
}

// WriteMessage writes a framed message
// Frame format: [4 bytes length (big-endian)] [message bytes]
func (f *Framer) WriteMessage(data []byte) error {
	if len(data) > f.maxFrameSize {
		return fmt.Errorf("message size %d exceeds max frame size %d", len(data), f.maxFrameSize)
	}
	
	// Write length header (4 bytes, big-endian)
	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, uint32(len(data)))
	
	if _, err := f.rw.Write(lengthBuf); err != nil {
		return fmt.Errorf("failed to write frame length: %w", err)
	}
	
	// Write message data
	if _, err := f.rw.Write(data); err != nil {
		return fmt.Errorf("failed to write frame data: %w", err)
	}
	
	return nil
}

// ReadMessage reads a framed message
func (f *Framer) ReadMessage() ([]byte, error) {
	// Read length header (4 bytes)
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(f.rw, lengthBuf); err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("failed to read frame length: %w", err)
	}
	
	// Parse length
	length := binary.BigEndian.Uint32(lengthBuf)
	if int(length) > f.maxFrameSize {
		return nil, fmt.Errorf("frame size %d exceeds max frame size %d", length, f.maxFrameSize)
	}
	
	// Read message data
	data := make([]byte, length)
	if _, err := io.ReadFull(f.rw, data); err != nil {
		return nil, fmt.Errorf("failed to read frame data: %w", err)
	}
	
	return data, nil
}