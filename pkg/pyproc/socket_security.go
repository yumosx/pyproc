package pyproc

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

// SocketSecurityConfig defines security settings for Unix domain sockets
type SocketSecurityConfig struct {
	// SocketDir is the directory where socket files will be created
	// Default: /run/pyproc if running as root, /tmp/pyproc otherwise
	SocketDir string

	// SocketPerms defines the permissions for socket files
	// Default: 0600 (read/write for owner only)
	SocketPerms os.FileMode

	// DirPerms defines the permissions for the socket directory
	// Default: 0750 (rwxr-x--- for owner and group)
	DirPerms os.FileMode

	// AllowedUIDs is a list of UIDs that are allowed to connect
	// If empty, any UID can connect (but still verified)
	AllowedUIDs []uint32

	// AllowedGIDs is a list of GIDs that are allowed to connect
	// If empty, any GID can connect (but still verified)
	AllowedGIDs []uint32

	// RequireSameUser if true, only allows connections from the same UID as the server
	RequireSameUser bool
}

// DefaultSocketSecurityConfig returns the default security configuration
func DefaultSocketSecurityConfig() SocketSecurityConfig {
	cfg := SocketSecurityConfig{
		SocketPerms:     0600,
		DirPerms:        0750,
		RequireSameUser: true,
	}

	// Use /run/pyproc if we have permissions, otherwise fallback to /tmp/pyproc
	if os.Geteuid() == 0 {
		cfg.SocketDir = "/run/pyproc"
	} else {
		cfg.SocketDir = filepath.Join(os.TempDir(), "pyproc")
	}

	return cfg
}

// SecureSocketPath creates a secure directory for socket files
func SecureSocketPath(config SocketSecurityConfig, socketName string) (string, error) {
	// Create the socket directory with proper permissions
	if err := os.MkdirAll(config.SocketDir, config.DirPerms); err != nil {
		return "", fmt.Errorf("failed to create socket directory %s: %w", config.SocketDir, err)
	}

	// Set directory permissions explicitly (in case it already existed)
	if err := os.Chmod(config.SocketDir, config.DirPerms); err != nil {
		return "", fmt.Errorf("failed to set permissions on socket directory: %w", err)
	}

	socketPath := filepath.Join(config.SocketDir, socketName)

	// Remove existing socket file if it exists
	if err := os.RemoveAll(socketPath); err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to remove existing socket file: %w", err)
	}

	return socketPath, nil
}

// SetSocketPermissions sets the appropriate permissions on a socket file
func SetSocketPermissions(socketPath string, perms os.FileMode) error {
	return os.Chmod(socketPath, perms)
}

// VerifyPeerCredentials verifies the credentials of a peer connection using SO_PEERCRED
func VerifyPeerCredentials(conn net.Conn, config SocketSecurityConfig) error {
	unixConn, ok := conn.(*net.UnixConn)
	if !ok {
		return errors.New("connection is not a Unix domain socket")
	}

	// Get the underlying file descriptor
	rawConn, err := unixConn.SyscallConn()
	if err != nil {
		return fmt.Errorf("failed to get raw connection: %w", err)
	}

	var peerCreds *PeerCredentials
	var credErr error

	// Get peer credentials using SO_PEERCRED
	err = rawConn.Control(func(fd uintptr) {
		peerCreds, credErr = getPeerCredentials(int(fd))
	})

	if err != nil {
		return fmt.Errorf("failed to control connection: %w", err)
	}
	if credErr != nil {
		return fmt.Errorf("failed to get peer credentials: %w", credErr)
	}
	if peerCreds == nil {
		return errors.New("peer credentials are nil")
	}

	// Verify credentials against configuration
	if config.RequireSameUser {
		currentUID := uint32(os.Geteuid())
		if peerCreds.UID != currentUID {
			return fmt.Errorf("peer UID %d does not match server UID %d", peerCreds.UID, currentUID)
		}
	}

	// Check allowed UIDs if specified
	if len(config.AllowedUIDs) > 0 {
		allowed := false
		for _, uid := range config.AllowedUIDs {
			if peerCreds.UID == uid {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("peer UID %d is not in allowed list", peerCreds.UID)
		}
	}

	// Check allowed GIDs if specified
	if len(config.AllowedGIDs) > 0 {
		allowed := false
		for _, gid := range config.AllowedGIDs {
			if peerCreds.GID == gid {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("peer GID %d is not in allowed list", peerCreds.GID)
		}
	}

	return nil
}

// getPeerCredentials is implemented in platform-specific files:
// - socket_security_linux.go for Linux (using SO_PEERCRED)
// - socket_security_darwin.go for macOS (using LOCAL_PEERCRED)

// SecureListener creates a Unix domain socket listener with security features
type SecureListener struct {
	net.Listener
	config SocketSecurityConfig
}

// NewSecureListener creates a new secure Unix domain socket listener
func NewSecureListener(socketPath string, config SocketSecurityConfig) (*SecureListener, error) {
	// Create secure socket path
	path, err := SecureSocketPath(config, filepath.Base(socketPath))
	if err != nil {
		return nil, err
	}

	// Create the listener
	listener, err := net.Listen("unix", path)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	// Set socket permissions
	if err := SetSocketPermissions(path, config.SocketPerms); err != nil {
		_ = listener.Close()
		return nil, fmt.Errorf("failed to set socket permissions: %w", err)
	}

	return &SecureListener{
		Listener: listener,
		config:   config,
	}, nil
}

// Accept accepts a connection and verifies peer credentials
func (l *SecureListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	// Verify peer credentials
	if err := VerifyPeerCredentials(conn, l.config); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("peer verification failed: %w", err)
	}

	return conn, nil
}
