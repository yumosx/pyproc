//go:build linux

package pyproc

import (
	"fmt"
	"syscall"
	"unsafe"
)

// getPeerCredentials retrieves the peer credentials using SO_PEERCRED (Linux-specific)
func getPeerCredentials(fd int) (*PeerCredentials, error) {
	ucred := &syscall.Ucred{}
	ucredLen := uint32(syscall.SizeofUcred)

	// Get peer credentials using SO_PEERCRED
	_, _, errno := syscall.Syscall6(
		syscall.SYS_GETSOCKOPT,
		uintptr(fd),
		uintptr(syscall.SOL_SOCKET),
		uintptr(syscall.SO_PEERCRED),
		uintptr(unsafe.Pointer(ucred)),
		uintptr(unsafe.Pointer(&ucredLen)),
		0,
	)

	if errno != 0 {
		return nil, fmt.Errorf("getsockopt SO_PEERCRED failed: %v", errno)
	}

	// Convert to platform-independent PeerCredentials
	return &PeerCredentials{
		UID: ucred.Uid,
		GID: ucred.Gid,
		PID: ucred.Pid,
	}, nil
}
