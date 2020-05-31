package phalanx

// ReadOptions is options of read key.
type ReadOptions int

// WriteOptions is options of write key.
type WriteOptions int

// StableStore is a local persistent storage.
type StableStore interface {
	// Put sets the value for the given key.
	Put(key, value []byte, wo *WriteOptions) error
	// CreateBatch creates batch
	CreateBatch() Batch
	// Write apply the given batch to the StableStorage
	Write(batch *Batch, wo *WriteOptions) error
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
	Dump() []byte
	Load(data []byte) error
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
	Next() bool
	Key() []byte
	Value() []byte
	Release()
	Error() error
}

// Transaction is an transaction
type Transaction interface {
	Commit() error
	Delete(key []byte, wo *WriteOptions) error
	Discard()
	Get(key []byte, ro *ReadOptions) ([]byte, error)
	Has(key []byte, ro *ReadOptions) (bool, error)
	NewIterator(slice *Range, ro ReadOptions) Iterator
	Put(key, value []byte, wo *WriteOptions) error
	Write(b *Batch, wo *WriteOptions) error
	CreateBatch() Batch
}

// Range is a key range
type Range struct {
	Start []byte
	Limit []byte
}
