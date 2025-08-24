//go:build darwin

package pyproc

import (
	"fmt"
	"syscall"
	"unsafe"
)

// getPeerCredentials retrieves the peer credentials using LOCAL_PEERCRED (macOS-specific)
func getPeerCredentials(fd int) (*PeerCredentials, error) {
	// On macOS, we use LOCAL_PEERCRED instead of SO_PEERCRED
	// The structure is different: struct xucred instead of struct ucred

	// Note: macOS doesn't provide PID in peer credentials

	type xucred struct {
		version uint32
		uid     uint32
		ngroups int16
		groups  [16]uint32
	}

	const LOCAL_PEERCRED = 0x001 // from sys/un.h
	const SOL_LOCAL = 0          // from sys/socket.h

	cred := &xucred{}
	credLen := uint32(unsafe.Sizeof(*cred))

	_, _, errno := syscall.Syscall6(
		syscall.SYS_GETSOCKOPT,
		uintptr(fd),
		uintptr(SOL_LOCAL),
		uintptr(LOCAL_PEERCRED),
		uintptr(unsafe.Pointer(cred)),
		uintptr(unsafe.Pointer(&credLen)),
		0,
	)

	if errno != 0 {
		return nil, fmt.Errorf("getsockopt LOCAL_PEERCRED failed: %v", errno)
	}

	// Convert to platform-independent PeerCredentials
	return &PeerCredentials{
		UID: cred.uid,
		GID: cred.groups[0], // Use first group as primary GID
		PID: 0,              // PID not available on macOS
	}, nil
}
