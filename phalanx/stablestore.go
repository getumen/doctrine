package phalanx

// ReadOptions is options of read key.
type ReadOptions struct {
	FillCache bool
}

// DefaultReadOptions returns default read options
func DefaultReadOptions() *ReadOptions {
	return &ReadOptions{
		FillCache: true,
	}
}

// WriteOptions is options of write key.
type WriteOptions struct {
	Sync bool
}

// DefaultWriteOptions returns default write options
func DefaultWriteOptions() *WriteOptions {
	return &WriteOptions{
		Sync: false,
	}
}

// StableStore is a local persistent storage.
type StableStore interface {
	// CreateBatch creates batch
	CreateBatch() Batch
	// Write apply the given batch to the StableStorage
	Write(batch Batch, wo *WriteOptions) error
	// CreateCheckpoint creates a checkpoint of this StableStore
	// In creating checkpoint, StableStore must be able to get keys
	CreateCheckpoint() ([]byte, error)
	// RestoreToCheckpoint restores internal storage to checkpoint
	RestoreToCheckpoint(checkpoint []byte) error
	// Close Close closes the StableStorage
	Close() error
	// GetSnapshot
	GetSnapshot() (Snapshot, error)
	// OpenTransaction returns Transaction
	OpenTransaction() (Transaction, error)
}

// Batch is a write batch
type Batch interface {
	Put(key, value []byte)
	Delete(key []byte)
	Len() int
	Reset()
}

// Snapshot is a snapshot of StableStorage
type Snapshot interface {
	Get(key []byte, ro *ReadOptions) (value []byte, err error)
	Has(key []byte, ro *ReadOptions) (ret bool, err error)
	NewIterator(slice *Range, ro *ReadOptions) Iterator
	Release()
}

// Iterator is an iterator of
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

// Transaction is an transaction
type Transaction interface {
	Commit() error
	Delete(key []byte, wo *WriteOptions) error
	Discard()
	Get(key []byte, ro *ReadOptions) ([]byte, error)
	Has(key []byte, ro *ReadOptions) (bool, error)
	NewIterator(slice *Range, ro *ReadOptions) Iterator
	Put(key, value []byte, wo *WriteOptions) error
	Write(b Batch, wo *WriteOptions) error
	CreateBatch() Batch
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
