package framing

import (
	"bytes"
	"testing"
)

func TestFrame(t *testing.T) {
	t.Run("NewFrame", func(t *testing.T) {
		requestID := uint64(12345)
		payload := []byte("Hello, World!")

		frame := NewFrame(requestID, payload)

		if frame.Header.RequestID != requestID {
			t.Errorf("RequestID mismatch: got %d, want %d", frame.Header.RequestID, requestID)
		}

		if !bytes.Equal(frame.Payload, payload) {
			t.Errorf("Payload mismatch: got %s, want %s", frame.Payload, payload)
		}

		if frame.Header.Magic[0] != MagicByte1 || frame.Header.Magic[1] != MagicByte2 {
			t.Errorf("Magic bytes mismatch: got %02x%02x, want %02x%02x",
				frame.Header.Magic[0], frame.Header.Magic[1], MagicByte1, MagicByte2)
		}

		expectedLength := uint32(FrameHeaderSize + len(payload))
		if frame.Header.Length != expectedLength {
			t.Errorf("Length mismatch: got %d, want %d", frame.Header.Length, expectedLength)
		}
	})

	t.Run("Marshal and Unmarshal", func(t *testing.T) {
		requestID := uint64(67890)
		payload := []byte("Test message with CRC32C")

		originalFrame := NewFrame(requestID, payload)

		// Marshal
		data := originalFrame.Marshal()

		// Unmarshal
		decodedFrame, err := UnmarshalFrame(data)
		if err != nil {
			t.Fatalf("UnmarshalFrame failed: %v", err)
		}

		// Verify fields
		if decodedFrame.Header.RequestID != requestID {
			t.Errorf("RequestID mismatch after marshal/unmarshal: got %d, want %d",
				decodedFrame.Header.RequestID, requestID)
		}

		if !bytes.Equal(decodedFrame.Payload, payload) {
			t.Errorf("Payload mismatch after marshal/unmarshal: got %s, want %s",
				decodedFrame.Payload, payload)
		}

		if decodedFrame.Header.CRC32C != originalFrame.Header.CRC32C {
			t.Errorf("CRC32C mismatch: got %08x, want %08x",
				decodedFrame.Header.CRC32C, originalFrame.Header.CRC32C)
		}
	})

	t.Run("CRC32C Validation", func(t *testing.T) {
		frame := NewFrame(1, []byte("Test data"))

		// Should validate correctly
		if !frame.ValidateChecksum() {
			t.Error("ValidateChecksum failed for valid frame")
		}

		// Corrupt the payload
		frame.Payload = []byte("Corrupted data")

		// Should fail validation
		if frame.ValidateChecksum() {
			t.Error("ValidateChecksum passed for corrupted frame")
		}

		// Update checksum
		frame.UpdateChecksum()

		// Should validate again
		if !frame.ValidateChecksum() {
			t.Error("ValidateChecksum failed after UpdateChecksum")
		}
	})

	t.Run("Invalid Magic Bytes", func(t *testing.T) {
		data := make([]byte, FrameHeaderSize+10)
		data[0] = 0xFF // Invalid magic byte
		data[1] = 0xFF

		_, err := UnmarshalFrame(data)
		if err == nil {
			t.Error("UnmarshalFrame should fail with invalid magic bytes")
		}
	})

	t.Run("CRC32C Mismatch", func(t *testing.T) {
		frame := NewFrame(123, []byte("Test"))
		data := frame.Marshal()

		// Corrupt the CRC32C field
		data[14] ^= 0xFF

		_, err := UnmarshalFrame(data)
		if err == nil {
			t.Error("UnmarshalFrame should fail with CRC32C mismatch")
		}
	})

	t.Run("Empty Payload", func(t *testing.T) {
		frame := NewFrame(999, []byte{})

		data := frame.Marshal()
		decodedFrame, err := UnmarshalFrame(data)
		if err != nil {
			t.Fatalf("UnmarshalFrame failed for empty payload: %v", err)
		}

		if len(decodedFrame.Payload) != 0 {
			t.Errorf("Expected empty payload, got %d bytes", len(decodedFrame.Payload))
		}
	})
}

func TestEnhancedFramer(t *testing.T) {
	t.Run("Enhanced Mode Read/Write", func(t *testing.T) {
		var buf bytes.Buffer
		framer := NewEnhancedFramer(&buf)

		// Write frame
		frame := NewFrame(42, []byte("Enhanced frame test"))
		if err := framer.WriteFrame(frame); err != nil {
			t.Fatalf("WriteFrame failed: %v", err)
		}

		// Read frame
		readFrame, err := framer.ReadFrame()
		if err != nil {
			t.Fatalf("ReadFrame failed: %v", err)
		}

		if readFrame.Header.RequestID != 42 {
			t.Errorf("RequestID mismatch: got %d, want 42", readFrame.Header.RequestID)
		}

		if !bytes.Equal(readFrame.Payload, frame.Payload) {
			t.Errorf("Payload mismatch: got %s, want %s", readFrame.Payload, frame.Payload)
		}
	})

	t.Run("Backward Compatibility", func(t *testing.T) {
		var buf bytes.Buffer

		// Write with standard framer
		standardFramer := NewFramer(&buf)
		message := []byte("Standard message")
		if err := standardFramer.WriteMessage(message); err != nil {
			t.Fatalf("WriteMessage failed: %v", err)
		}

		// Read with standard framer
		readMessage, err := standardFramer.ReadMessage()
		if err != nil {
			t.Fatalf("ReadMessage failed: %v", err)
		}

		if !bytes.Equal(readMessage, message) {
			t.Errorf("Message mismatch: got %s, want %s", readMessage, message)
		}
	})
}

func BenchmarkFrame(b *testing.B) {
	payload := []byte("Benchmark payload for testing frame performance")

	b.Run("NewFrame", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewFrame(uint64(i), payload)
		}
	})

	b.Run("Marshal", func(b *testing.B) {
		frame := NewFrame(1, payload)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = frame.Marshal()
		}
	})

	b.Run("Unmarshal", func(b *testing.B) {
		frame := NewFrame(1, payload)
		data := frame.Marshal()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = UnmarshalFrame(data)
		}
	})

	b.Run("CRC32C", func(b *testing.B) {
		frame := NewFrame(1, payload)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			frame.UpdateChecksum()
		}
	})
}
