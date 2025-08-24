//go:build !json_goccy && !json_segmentio

package pyproc

import (
	"encoding/json"
)

// JSONCodec implements Codec using standard library encoding/json
type JSONCodec struct{}

// Marshal serializes a value to JSON bytes using standard library
func (c *JSONCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal deserializes JSON bytes to a value using standard library
func (c *JSONCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// Name returns the name of the codec
func (c *JSONCodec) Name() string {
	return "json-stdlib"
}
