package pyproc

// PeerCredentials represents the credentials of a peer process
// This is a platform-independent representation
type PeerCredentials struct {
	UID uint32 // User ID
	GID uint32 // Group ID
	PID int32  // Process ID (may be 0 on some platforms)
}
