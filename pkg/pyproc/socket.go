package pyproc

import (
	"fmt"
	"os"
	"path/filepath"
)

// SocketManager manages Unix domain socket files
type SocketManager struct {
	dir         string
	prefix      string
	permissions os.FileMode
}

// NewSocketManager creates a new socket manager
func NewSocketManager(cfg SocketConfig) *SocketManager {
	return &SocketManager{
		dir:         cfg.Dir,
		prefix:      cfg.Prefix,
		permissions: os.FileMode(cfg.Permissions),
	}
}

// GenerateSocketPath generates a unique socket path for a worker
func (sm *SocketManager) GenerateSocketPath(workerID string) string {
	filename := fmt.Sprintf("%s-%s.sock", sm.prefix, workerID)
	return filepath.Join(sm.dir, filename)
}

// CleanupSocket removes a socket file if it exists
func (sm *SocketManager) CleanupSocket(socketPath string) error {
	// Check if the file exists
	if _, err := os.Stat(socketPath); err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, nothing to clean up
			return nil
		}
		return fmt.Errorf("failed to stat socket file: %w", err)
	}

	// Remove the socket file
	if err := os.Remove(socketPath); err != nil {
		return fmt.Errorf("failed to remove socket file: %w", err)
	}

	return nil
}

// CleanupAllSockets removes all socket files matching the prefix
func (sm *SocketManager) CleanupAllSockets() error {
	pattern := filepath.Join(sm.dir, fmt.Sprintf("%s-*.sock", sm.prefix))
	
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to glob socket files: %w", err)
	}

	var lastErr error
	for _, socketPath := range matches {
		if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
			lastErr = fmt.Errorf("failed to remove socket %s: %w", socketPath, err)
		}
	}

	return lastErr
}

// EnsureSocketDir ensures the socket directory exists with proper permissions
func (sm *SocketManager) EnsureSocketDir() error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(sm.dir, 0755); err != nil {
		return fmt.Errorf("failed to create socket directory: %w", err)
	}

	return nil
}

// SetSocketPermissions sets the proper permissions on a socket file
func (sm *SocketManager) SetSocketPermissions(socketPath string) error {
	if err := os.Chmod(socketPath, sm.permissions); err != nil {
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}
	return nil
}