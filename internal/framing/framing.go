// Package framing implements the 4-byte length prefixed framing protocol
// for reliable message transmission over Unix Domain Sockets.
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
	rw           io.ReadWriter
	maxFrameSize int
	// Enhanced mode enables request ID and CRC32C
	enhancedMode bool
}

// NewFramer creates a new framer with default max frame size
func NewFramer(rw io.ReadWriter) *Framer {
	return &Framer{
		rw:           rw,
		maxFrameSize: DefaultMaxFrameSize,
		enhancedMode: false,
	}
}

// NewFramerWithMaxSize creates a new framer with specified max frame size
func NewFramerWithMaxSize(rw io.ReadWriter, maxSize int) *Framer {
	return &Framer{
		rw:           rw,
		maxFrameSize: maxSize,
		enhancedMode: false,
	}
}

// NewEnhancedFramer creates a framer with request ID and CRC32C support
func NewEnhancedFramer(rw io.ReadWriter) *Framer {
	return &Framer{
		rw:           rw,
		maxFrameSize: DefaultMaxFrameSize,
		enhancedMode: true,
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

// WriteFrame writes an enhanced frame with request ID and CRC32C
func (f *Framer) WriteFrame(frame *Frame) error {
	if !f.enhancedMode {
		// Fall back to simple message write
		return f.WriteMessage(frame.Payload)
	}

	if len(frame.Payload) > f.maxFrameSize {
		return fmt.Errorf("payload size %d exceeds max frame size %d", len(frame.Payload), f.maxFrameSize)
	}

	// Marshal the entire frame
	data := frame.Marshal()

	// Write the complete frame
	if _, err := f.rw.Write(data); err != nil {
		return fmt.Errorf("failed to write frame: %w", err)
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

// ReadFrame reads an enhanced frame with request ID and CRC32C
func (f *Framer) ReadFrame() (*Frame, error) {
	if !f.enhancedMode {
		// Fall back to simple message read
		data, err := f.ReadMessage()
		if err != nil {
			return nil, err
		}
		// Create a simple frame with no request ID
		return &Frame{
			Payload: data,
		}, nil
	}

	// Peek at magic bytes first
	magicBuf := make([]byte, 2)
	if _, err := io.ReadFull(f.rw, magicBuf); err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("failed to read magic bytes: %w", err)
	}

	// Check magic bytes
	if magicBuf[0] != MagicByte1 || magicBuf[1] != MagicByte2 {
		return nil, fmt.Errorf("invalid magic bytes: %02x%02x", magicBuf[0], magicBuf[1])
	}

	// Read the rest of the header
	headerBuf := make([]byte, FrameHeaderSize-2) // -2 for magic bytes already read
	if _, err := io.ReadFull(f.rw, headerBuf); err != nil {
		return nil, fmt.Errorf("failed to read frame header: %w", err)
	}

	// Parse header fields
	length := binary.BigEndian.Uint32(headerBuf[0:4])
	if int(length) > f.maxFrameSize+FrameHeaderSize {
		return nil, fmt.Errorf("frame size %d exceeds max frame size %d", length, f.maxFrameSize)
	}

	// Read payload
	payloadSize := int(length) - FrameHeaderSize
	payload := make([]byte, payloadSize)
	if payloadSize > 0 {
		if _, err := io.ReadFull(f.rw, payload); err != nil {
			return nil, fmt.Errorf("failed to read frame payload: %w", err)
		}
	}

	// Reconstruct complete frame data for unmarshaling
	completeData := make([]byte, length)
	copy(completeData[0:2], magicBuf)
	copy(completeData[2:FrameHeaderSize], headerBuf)
	if payloadSize > 0 {
		copy(completeData[FrameHeaderSize:], payload)
	}

	// Unmarshal and validate
	return UnmarshalFrame(completeData)
}
