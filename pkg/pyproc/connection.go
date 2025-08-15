package pyproc

import (
	"fmt"
	"net"
	"time"
)

// ConnectToWorker connects to a worker via Unix domain socket
func ConnectToWorker(socketPath string, timeout time.Duration) (net.Conn, error) {
	// Set connection timeout
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		conn, err := net.Dial("unix", socketPath)
		if err == nil {
			return conn, nil
		}

		// Wait a bit before retrying
		time.Sleep(100 * time.Millisecond)
	}

	return nil, fmt.Errorf("failed to connect to worker at %s after %v", socketPath, timeout)
}
