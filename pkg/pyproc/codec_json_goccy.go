//go:build json_goccy

package pyproc

import (
	"github.com/goccy/go-json"
)

// JSONCodec implements Codec using goccy/go-json for high performance
type JSONCodec struct{}

// Marshal serializes a value to JSON bytes using goccy/go-json
func (c *JSONCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal deserializes JSON bytes to a value using goccy/go-json
func (c *JSONCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// Name returns the name of the codec
func (c *JSONCodec) Name() string {
	return "json-goccy"
}
