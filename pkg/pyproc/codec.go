package pyproc

import (
	"fmt"
	"os"
)

// Codec defines the interface for encoding/decoding messages
type Codec interface {
	// Marshal serializes a value to bytes
	Marshal(v interface{}) ([]byte, error)

	// Unmarshal deserializes bytes to a value
	Unmarshal(data []byte, v interface{}) error

	// Name returns the name of the codec
	Name() string
}

// CodecType represents the type of codec to use
type CodecType string

const (
	// CodecJSON uses JSON encoding (default)
	CodecJSON CodecType = "json"
	// CodecMessagePack uses MessagePack encoding
	CodecMessagePack CodecType = "msgpack"
	// CodecProtobuf uses Protocol Buffers encoding
	CodecProtobuf CodecType = "protobuf"
)

// GetJSONCodecType returns the JSON codec implementation being used
// Can be overridden with PYPROC_JSON_CODEC environment variable
func GetJSONCodecType() string {
	if codecType := os.Getenv("PYPROC_JSON_CODEC"); codecType != "" {
		return codecType
	}
	// Return the compile-time selected codec
	return (&JSONCodec{}).Name()
}

// NewCodec creates a new codec based on the type
func NewCodec(codecType CodecType) (Codec, error) {
	switch codecType {
	case CodecJSON, "":
		return &JSONCodec{}, nil
	case CodecMessagePack:
		return &MessagePackCodec{}, nil
	case CodecProtobuf:
		// TODO: Implement in Phase 3
		return nil, fmt.Errorf("protobuf codec not yet implemented")
	default:
		return nil, fmt.Errorf("unknown codec type: %s", codecType)
	}
}
