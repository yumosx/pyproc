package pyproc

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"time"
)

// HMACAuth provides HMAC-based authentication for socket connections
type HMACAuth struct {
	secret []byte
}

// NewHMACAuth creates a new HMAC authenticator with the given secret
func NewHMACAuth(secret []byte) *HMACAuth {
	return &HMACAuth{
		secret: secret,
	}
}

// GenerateSecret generates a random secret key
func GenerateSecret() ([]byte, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}
	return secret, nil
}

// AuthenticateClient performs client-side authentication
func (h *HMACAuth) AuthenticateClient(conn net.Conn) error {
	// Set timeout for auth handshake
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return fmt.Errorf("failed to set deadline: %w", err)
	}
	defer conn.SetDeadline(time.Time{}) // Reset deadline

	// Read challenge from server
	challenge := make([]byte, 32)
	if _, err := io.ReadFull(conn, challenge); err != nil {
		return fmt.Errorf("failed to read challenge: %w", err)
	}

	// Compute HMAC response
	mac := hmac.New(sha256.New, h.secret)
	mac.Write(challenge)
	response := mac.Sum(nil)

	// Send response
	if _, err := conn.Write(response); err != nil {
		return fmt.Errorf("failed to send response: %w", err)
	}

	// Read authentication result
	result := make([]byte, 1)
	if _, err := io.ReadFull(conn, result); err != nil {
		return fmt.Errorf("failed to read auth result: %w", err)
	}

	if result[0] != 1 {
		return fmt.Errorf("authentication failed")
	}

	return nil
}

// AuthenticateServer performs server-side authentication
func (h *HMACAuth) AuthenticateServer(conn net.Conn) error {
	// Set timeout for auth handshake
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return fmt.Errorf("failed to set deadline: %w", err)
	}
	defer conn.SetDeadline(time.Time{}) // Reset deadline

	// Generate random challenge
	challenge := make([]byte, 32)
	if _, err := rand.Read(challenge); err != nil {
		return fmt.Errorf("failed to generate challenge: %w", err)
	}

	// Send challenge to client
	if _, err := conn.Write(challenge); err != nil {
		return fmt.Errorf("failed to send challenge: %w", err)
	}

	// Read response from client
	response := make([]byte, 32)
	if _, err := io.ReadFull(conn, response); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Verify HMAC
	mac := hmac.New(sha256.New, h.secret)
	mac.Write(challenge)
	expected := mac.Sum(nil)

	if !hmac.Equal(response, expected) {
		// Authentication failed
		conn.Write([]byte{0})
		return fmt.Errorf("HMAC verification failed")
	}

	// Authentication succeeded
	if _, err := conn.Write([]byte{1}); err != nil {
		return fmt.Errorf("failed to send auth success: %w", err)
	}

	return nil
}

// SecureListener wraps a listener with HMAC authentication
type HMACListener struct {
	net.Listener
	auth *HMACAuth
}

// NewHMACListener creates a new HMAC-authenticated listener
func NewHMACListener(listener net.Listener, secret []byte) *HMACListener {
	return &HMACListener{
		Listener: listener,
		auth:     NewHMACAuth(secret),
	}
}

// Accept accepts a connection and performs HMAC authentication
func (l *HMACListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	// Perform authentication
	if err := l.auth.AuthenticateServer(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return conn, nil
}

// SecureConn wraps a connection with HMAC authentication
type SecureConn struct {
	net.Conn
	authenticated bool
}

// DialSecure dials a connection with HMAC authentication
func DialSecure(network, address string, secret []byte) (*SecureConn, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}

	auth := NewHMACAuth(secret)
	if err := auth.AuthenticateClient(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return &SecureConn{
		Conn:          conn,
		authenticated: true,
	}, nil
}

// IsAuthenticated returns whether the connection is authenticated
func (c *SecureConn) IsAuthenticated() bool {
	return c.authenticated
}

// SecretFromString creates a secret from a string
func SecretFromString(s string) []byte {
	h := sha256.Sum256([]byte(s))
	return h[:]
}

// SecretFromHex decodes a hex-encoded secret
func SecretFromHex(hexStr string) ([]byte, error) {
	return hex.DecodeString(hexStr)
}

