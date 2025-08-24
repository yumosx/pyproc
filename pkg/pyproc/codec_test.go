package pyproc

import (
	"reflect"
	"testing"
)

func TestGetJSONCodecType(t *testing.T) {
	t.Run("DefaultCodec", func(t *testing.T) {
		codecType := GetJSONCodecType()
		// Should be one of the valid JSON codec types
		validTypes := map[string]bool{
			"json-stdlib":    true,
			"json-goccy":     true,
			"json-segmentio": true,
		}
		if !validTypes[codecType] {
			t.Errorf("GetJSONCodecType() = %v, want one of json-stdlib, json-goccy, or json-segmentio", codecType)
		}
	})
}

func TestJSONCodec(t *testing.T) {
	codec := &JSONCodec{}

	// Test Marshal/Unmarshal with different types
	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name:  "string",
			input: "hello world",
		},
		{
			name:  "int",
			input: 42,
		},
		{
			name: "struct",
			input: struct {
				Name  string `json:"name"`
				Value int    `json:"value"`
			}{
				Name:  "test",
				Value: 123,
			},
		},
		{
			name: "map",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": float64(42), // JSON unmarshals numbers as float64
				"key3": true,
			},
		},
		{
			name:  "slice",
			input: []int{1, 2, 3, 4, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := codec.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			// Unmarshal
			outputType := reflect.TypeOf(tt.input)
			output := reflect.New(outputType).Interface()

			if err := codec.Unmarshal(data, output); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			// Compare
			actual := reflect.ValueOf(output).Elem().Interface()
			if !reflect.DeepEqual(tt.input, actual) {
				t.Errorf("Round-trip failed: got %v, want %v", actual, tt.input)
			}
		})
	}
}

func TestMessagePackCodec(t *testing.T) {
	codec := &MessagePackCodec{}

	// Test Marshal/Unmarshal with different types
	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name:  "string",
			input: "hello msgpack",
		},
		{
			name:  "int",
			input: 256,
		},
		{
			name: "map",
			input: map[string]interface{}{
				"msgpack": true,
				"fast":    "yes",
				"size":    int64(100), // MessagePack unmarshals integers as int64
			},
		},
		{
			name:  "slice",
			input: []string{"a", "b", "c"},
		},
		{
			name:  "bytes",
			input: []byte{0x01, 0x02, 0x03, 0x04},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := codec.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			// Unmarshal
			outputType := reflect.TypeOf(tt.input)
			output := reflect.New(outputType).Interface()

			if err := codec.Unmarshal(data, output); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			// Compare
			actual := reflect.ValueOf(output).Elem().Interface()
			if !reflect.DeepEqual(tt.input, actual) {
				t.Errorf("Round-trip failed: got %v, want %v", actual, tt.input)
			}
		})
	}
}

func TestNewCodec(t *testing.T) {
	// Get the actual JSON codec name at runtime
	jsonCodecName := (&JSONCodec{}).Name()

	tests := []struct {
		name      string
		codecType CodecType
		wantName  string
		wantErr   bool
	}{
		{
			name:      "JSON",
			codecType: CodecJSON,
			wantName:  jsonCodecName, // Will be json-stdlib, json-goccy, or json-segmentio
			wantErr:   false,
		},
		{
			name:      "MessagePack",
			codecType: CodecMessagePack,
			wantName:  "msgpack",
			wantErr:   false,
		},
		{
			name:      "Default (empty string)",
			codecType: "",
			wantName:  jsonCodecName,
			wantErr:   false,
		},
		{
			name:      "Protobuf (not implemented)",
			codecType: CodecProtobuf,
			wantName:  "",
			wantErr:   true,
		},
		{
			name:      "Unknown",
			codecType: "unknown",
			wantName:  "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codec, err := NewCodec(tt.codecType)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCodec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && codec.Name() != tt.wantName {
				t.Errorf("NewCodec() codec name = %v, want %v", codec.Name(), tt.wantName)
			}
		})
	}
}

func BenchmarkJSONCodec(b *testing.B) {
	codec := &JSONCodec{}
	data := map[string]interface{}{
		"method": "predict",
		"params": map[string]interface{}{
			"input": "test data",
			"count": 100,
		},
	}

	b.Run("Marshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := codec.Marshal(data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	marshaled, _ := codec.Marshal(data)
	b.Run("Unmarshal", func(b *testing.B) {
		var output map[string]interface{}
		for i := 0; i < b.N; i++ {
			err := codec.Unmarshal(marshaled, &output)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkMessagePackCodec(b *testing.B) {
	codec := &MessagePackCodec{}
	data := map[string]interface{}{
		"method": "predict",
		"params": map[string]interface{}{
			"input": "test data",
			"count": 100,
		},
	}

	b.Run("Marshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := codec.Marshal(data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	marshaled, _ := codec.Marshal(data)
	b.Run("Unmarshal", func(b *testing.B) {
		var output map[string]interface{}
		for i := 0; i < b.N; i++ {
			err := codec.Unmarshal(marshaled, &output)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
