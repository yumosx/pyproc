// Package framing implements an enhanced framing protocol with request ID
// and CRC32C checksum for reliable multiplexed message transmission.
package framing

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
)

// Frame header constants
const (
	// Frame header size: 2 (magic) + 4 (length) + 8 (request ID) + 4 (CRC32C) = 18 bytes
	FrameHeaderSize = 18

	// Magic bytes to identify valid frames
	MagicByte1 = 0x50 // 'P'
	MagicByte2 = 0x59 // 'Y'
)

// FrameHeader represents the enhanced frame header
type FrameHeader struct {
	Magic     [2]byte // Magic bytes for frame validation
	Length    uint32  // Total frame length (including header)
	RequestID uint64  // Request ID for multiplexing
	CRC32C    uint32  // CRC32C checksum of the payload
}

// Frame represents a complete frame with header and payload
type Frame struct {
	Header  FrameHeader
	Payload []byte
}

var crc32cTable = crc32.MakeTable(crc32.Castagnoli)

// NewFrame creates a new frame with the given request ID and payload
func NewFrame(requestID uint64, payload []byte) *Frame {
	return &Frame{
		Header: FrameHeader{
			Magic:     [2]byte{MagicByte1, MagicByte2},
			Length:    uint32(FrameHeaderSize + len(payload)),
			RequestID: requestID,
			CRC32C:    crc32.Checksum(payload, crc32cTable),
		},
		Payload: payload,
	}
}

// Marshal serializes the frame to bytes
func (f *Frame) Marshal() []byte {
	buf := make([]byte, f.Header.Length)

	// Write magic bytes
	buf[0] = f.Header.Magic[0]
	buf[1] = f.Header.Magic[1]

	// Write length (4 bytes, big-endian)
	binary.BigEndian.PutUint32(buf[2:6], f.Header.Length)

	// Write request ID (8 bytes, big-endian)
	binary.BigEndian.PutUint64(buf[6:14], f.Header.RequestID)

	// Write CRC32C (4 bytes, big-endian)
	binary.BigEndian.PutUint32(buf[14:18], f.Header.CRC32C)

	// Copy payload (starting after the header)
	if len(f.Payload) > 0 {
		copy(buf[FrameHeaderSize:], f.Payload)
	}

	return buf
}

// UnmarshalFrame deserializes a frame from bytes
func UnmarshalFrame(data []byte) (*Frame, error) {
	if len(data) < FrameHeaderSize {
		return nil, fmt.Errorf("frame too short: %d bytes", len(data))
	}

	// Check magic bytes
	if data[0] != MagicByte1 || data[1] != MagicByte2 {
		return nil, fmt.Errorf("invalid magic bytes: %02x%02x", data[0], data[1])
	}

	// Parse header
	header := FrameHeader{
		Magic:     [2]byte{data[0], data[1]},
		Length:    binary.BigEndian.Uint32(data[2:6]),
		RequestID: binary.BigEndian.Uint64(data[6:14]),
		CRC32C:    binary.BigEndian.Uint32(data[14:18]),
	}

	// Validate length
	if int(header.Length) != len(data) {
		return nil, fmt.Errorf("frame length mismatch: header says %d, got %d", header.Length, len(data))
	}

	// Extract payload (starting after the header)
	payload := data[FrameHeaderSize:]

	// Verify CRC32C
	calculatedCRC := crc32.Checksum(payload, crc32cTable)
	if calculatedCRC != header.CRC32C {
		return nil, fmt.Errorf("CRC32C mismatch: expected %08x, got %08x", header.CRC32C, calculatedCRC)
	}

	return &Frame{
		Header:  header,
		Payload: payload,
	}, nil
}

// ValidateChecksum verifies the CRC32C checksum
func (f *Frame) ValidateChecksum() bool {
	return crc32.Checksum(f.Payload, crc32cTable) == f.Header.CRC32C
}

// UpdateChecksum recalculates and updates the CRC32C checksum
func (f *Frame) UpdateChecksum() {
	f.Header.CRC32C = crc32.Checksum(f.Payload, crc32cTable)
	f.Header.Length = uint32(FrameHeaderSize + len(f.Payload))
}
