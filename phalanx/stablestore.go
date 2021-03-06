package phalanx

// StableStore is a local persistent storage.
type StableStore interface {
	// CreateBatch creates batch
	CreateBatch() Batch
	// Write apply the given batch to the StableStorage
	Write(batch Batch) error
	// Close Close closes the StableStorage
	Close() error
	// GetSnapshot returns snapshot
	GetSnapshot() (Snapshot, error)
	// CreateCheckpoint creates a checkpoint of the given region
	// In creating checkpoint, StableStore must be able to get keys
	// CreateCheckpoint returns checkpointInfo which enable stable store to restore to the checkpoint
	// For example, marshaled Amazon S3 bucket and object name.
	CreateCheckpoint(region string) ([]byte, error)
	// RestoreToCheckpoint restores the given region to checkpoint
	RestoreToCheckpoint(region string, checkpointInfo []byte) error
	// CreateRegion creates a region
	CreateRegion(name string) error
	// DropRegion drop a region
	DropRegion(name string) error
	// HasRegion returns if a region exists
	HasRegion(name string) bool
}

// Batch is a write batch
type Batch interface {
	Put(region string, key, value []byte)
	Delete(region string, key []byte)
	Len() int
	Reset()
}

// Snapshot is a snapshot of StableStorage
type Snapshot interface {
	Get(region string, key []byte) (value []byte, err error)
	MultiGet(region string, keys ...[]byte) (values [][]byte, err error)
	Has(region string, key []byte) (ret bool, err error)
	NewIterator(region string, slice *Range) (Iterator, error)
	Release()
}

// Iterator is an iterator of db
// not thread safe
type Iterator interface {
	Key() []byte
	Value() []byte
	Release()
	Error() error
	// First moves the iterator to the first key/value pair. If the iterator
	// only contains one key/value pair then First and Last would moves
	// to the same key/value pair.
	// It returns whether such pair exist.
	First() bool

	// Last moves the iterator to the last key/value pair. If the iterator
	// only contains one key/value pair then First and Last would moves
	// to the same key/value pair.
	// It returns whether such pair exist.
	Last() bool

	// Seek moves the iterator to the first key/value pair whose key is greater
	// than or equal to the given key.
	// It returns whether such pair exist.
	//
	// It is safe to modify the contents of the argument after Seek returns.
	Seek(key []byte) bool

	// Next moves the iterator to the next key/value pair.
	// It returns false if the iterator is exhausted.
	Next() bool

	// Prev moves the iterator to the previous key/value pair.
	// It returns false if the iterator is exhausted.
	Prev() bool
}

// Range is a key range
type Range struct {
	Start []byte
	End   []byte
}

// BytesPrefixRange returns Range of the givein prefix
func BytesPrefixRange(prefix []byte) *Range {
	var end []byte
	for i := len(prefix) - 1; i >= 0; i-- {
		c := prefix[i]
		if c < 0xff {
			end = make([]byte, i+1)
			copy(end, prefix)
			end[i] = c + 1
			break
		}
	}
	return &Range{Start: prefix, End: end}
}

// FullScanRange returns full scan range
func FullScanRange() *Range {
	return &Range{nil, nil}
}
