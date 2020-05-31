package phalanx

// ReadOptions is options of read key.
type ReadOptions int

// StableStore is a local persistent storage.
type StableStore interface {
	// Get returns value corresponding to the key
	Get([]byte, ...ReadOptions) ([]byte, error)
	// CreateSnapshot creates a snapshot of this StableStore
	// In creating snapshot, StableStore must be able to get keys
	CreateSnapshot() ([]byte, error)
	// RestoreToSnapshot restores internal storage to snapshot
	RestoreToSnapshot(snapshot []byte) error
}
