package protocol

import (
	"encoding/json"
	"fmt"
)

// Request represents a request from Go to Python
type Request struct {
	ID     uint64          `json:"id"`
	Method string          `json:"method"`
	Body   json.RawMessage `json:"body"`
}

// Response represents a response from Python to Go
type Response struct {
	ID       uint64          `json:"id"`
	OK       bool            `json:"ok"`
	Body     json.RawMessage `json:"body,omitempty"`
	ErrorMsg string          `json:"error,omitempty"`
}

// NewRequest creates a new request with the given method and body
func NewRequest(id uint64, method string, body interface{}) (*Request, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	
	return &Request{
		ID:     id,
		Method: method,
		Body:   bodyBytes,
	}, nil
}

// NewResponse creates a new successful response
func NewResponse(id uint64, body interface{}) (*Response, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response body: %w", err)
	}
	
	return &Response{
		ID:   id,
		OK:   true,
		Body: bodyBytes,
	}, nil
}

// NewErrorResponse creates a new error response
func NewErrorResponse(id uint64, err error) *Response {
	return &Response{
		ID:       id,
		OK:       false,
		ErrorMsg: err.Error(),
	}
}

// Marshal serializes the request to JSON
func (r *Request) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// Unmarshal deserializes the request from JSON
func (r *Request) Unmarshal(data []byte) error {
	return json.Unmarshal(data, r)
}

// UnmarshalBody unmarshals the request body into the given interface
func (r *Request) UnmarshalBody(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// Marshal serializes the response to JSON
func (r *Response) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// Unmarshal deserializes the response from JSON
func (r *Response) Unmarshal(data []byte) error {
	return json.Unmarshal(data, r)
}

// UnmarshalBody unmarshals the response body into the given interface
func (r *Response) UnmarshalBody(v interface{}) error {
	if r.Body == nil {
		return fmt.Errorf("response body is nil")
	}
	return json.Unmarshal(r.Body, v)
}

// Error returns the error message if the response is an error
func (r *Response) Error() error {
	if r.OK {
		return nil
	}
	if r.ErrorMsg == "" {
		return fmt.Errorf("unknown error")
	}
	return fmt.Errorf(r.ErrorMsg)
}