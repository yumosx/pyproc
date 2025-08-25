//go:build json_segmentio

package pyproc

import (
	"github.com/segmentio/encoding/json"
)

// JSONCodec implements Codec using segmentio/encoding/json for high performance
type JSONCodec struct{}

// Marshal serializes a value to JSON bytes using segmentio/encoding/json
func (c *JSONCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal deserializes JSON bytes to a value using segmentio/encoding/json
func (c *JSONCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// Name returns the name of the codec
func (c *JSONCodec) Name() string {
	return "json-segmentio"
}
