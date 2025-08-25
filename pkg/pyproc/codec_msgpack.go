package pyproc

import (
	"github.com/vmihailenco/msgpack/v5"
)

// MessagePackCodec implements Codec using MessagePack encoding
type MessagePackCodec struct{}

// Marshal serializes a value to MessagePack bytes
func (c *MessagePackCodec) Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}

// Unmarshal deserializes MessagePack bytes to a value
func (c *MessagePackCodec) Unmarshal(data []byte, v interface{}) error {
	return msgpack.Unmarshal(data, v)
}

// Name returns the name of the codec
func (c *MessagePackCodec) Name() string {
	return "msgpack"
}
