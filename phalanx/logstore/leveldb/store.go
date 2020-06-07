package leveldblogstore

import (
	"encoding/binary"
	"log"

	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/hashicorp/go-multierror"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	hardStateKey byte = 0x00
	snapshotKey  byte = 0x01
	logPrefix    byte = 0xff // logPrefix MUST be last prefix
)

type store struct {
	internal *leveldb.DB
}

// InitialState implements the Storage interface.
func (s *store) InitialState() (raftpb.HardState, raftpb.ConfState, error) {
	snap, err := s.internal.GetSnapshot()
	if err != nil {
		return raftpb.HardState{}, raftpb.ConfState{}, err
	}
	defer snap.Release()

	hardState, err := s.hardState(snap)
	if err != nil {
		return raftpb.HardState{}, raftpb.ConfState{}, err
	}

	snapshot, err := s.snapshot(snap)
	if err != nil {
		return raftpb.HardState{}, raftpb.ConfState{}, err
	}

	return hardState, snapshot.Metadata.ConfState, nil
}

// SetHardState saves the current HardState.
func (s *store) SetHardState(st raftpb.HardState) error {
	tx, err := s.internal.OpenTransaction()
	if err != nil {
		return err
	}
	hardStateBin, err := st.Marshal()
	if err != nil {
		tx.Discard()
		return err
	}
	err = tx.Put([]byte{hardStateKey}, hardStateBin, nil)
	if err != nil {
		tx.Discard()
		return err
	}
	err = tx.Commit()
	if err != nil {
		tx.Discard()
		return err
	}
	return nil
}

// Entries implements the Storage interface.
func (s *store) Entries(lo, hi, maxSize uint64) ([]raftpb.Entry, error) {

	snap, err := s.internal.GetSnapshot()
	if err != nil {
		return nil, err
	}
	defer snap.Release()

	firstIndex, err := s.firstIndex(snap)

	if lo <= firstIndex {
		return nil, raft.ErrCompacted
	}

	lastIndex, err := s.lastIndex(snap)

	if hi > lastIndex+1 {
		log.Panicf("entries' hi(%d) is out of bound lastindex(%d)", hi, lastIndex)
	}

	lowBytes := make([]byte, 8)
	highBytes := make([]byte, 8)

	binary.BigEndian.PutUint64(lowBytes, lo)
	binary.BigEndian.PutUint64(highBytes, hi)

	lowKey := s.logKey(lowBytes)
	highKey := s.logKey(highBytes)

	iter := snap.NewIterator(&util.Range{Start: lowKey, Limit: highKey}, nil)

	var resultError *multierror.Error
	var size uint64
	entries := make([]raftpb.Entry, 0)

	for iter.Next() {
		var ent raftpb.Entry
		err = ent.Unmarshal(iter.Value())
		if err != nil {
			resultError = multierror.Append(resultError, err)
			break
		}
		size += uint64(ent.Size())
		if size >= maxSize {
			break
		}

		entries = append(entries, ent)
	}

	// TODO: handle dummy entries
	// // only contains dummy entries.
	// if len(ms.ents) == 1 {
	// 	return nil, ErrUnavailable
	// }

	iter.Release()

	err = iter.Error()
	if err != nil {
		resultError = multierror.Append(resultError, err)
	}

	return entries, resultError.ErrorOrNil()
}

func (s *store) logKey(rawKey []byte) []byte {
	return append([]byte{logPrefix}, rawKey...)
}

// Term implements the Storage interface.
func (s *store) Term(termIndex uint64) (uint64, error) {

	snap, err := s.internal.GetSnapshot()
	if err != nil {
		return 0, err
	}
	defer snap.Release()

	first, err := s.firstIndex(snap)

	if err != nil {
		log.Panicln("leveldb logstore: first log does not exist in log store")
	}

	if termIndex < first {
		return 0, raft.ErrCompacted
	}

	termIndexByte := make([]byte, 8)
	binary.BigEndian.PutUint64(termIndexByte, termIndex)

	termIndexKey := s.logKey(termIndexByte)

	value, err := snap.Get(termIndexKey, nil)

	if err == leveldb.ErrNotFound {
		return 0, raft.ErrUnavailable
	}

	var ent raftpb.Entry
	err = ent.Unmarshal(value)
	if err != nil {
		return 0, err
	}

	return ent.Term, nil
}

// FirstIndex implements the Storage interface.
func (s *store) FirstIndex() (uint64, error) {
	snap, err := s.internal.GetSnapshot()
	if err != nil {
		return 0, err
	}
	defer snap.Release()

	return s.firstIndex(snap)
}

// LastIndex implements the Storage interface.
func (s *store) LastIndex() (uint64, error) {
	snap, err := s.internal.GetSnapshot()
	if err != nil {
		return 0, err
	}
	defer snap.Release()

	return s.lastIndex(snap)
}

func (s *store) firstIndex(snap *leveldb.Snapshot) (uint64, error) {
	iter := snap.NewIterator(util.BytesPrefix([]byte{logPrefix}), nil)

	defer iter.Release()

	for iter.Next() {
		first := binary.BigEndian.Uint64(iter.Key())
		return first, nil
	}

	return 0, leveldb.ErrNotFound
}

func (s *store) lastIndex(snap *leveldb.Snapshot) (uint64, error) {
	iter := snap.NewIterator(nil, nil)

	defer iter.Release()

	if exists := iter.Last(); !exists {
		return 0, leveldb.ErrNotFound
	}

	last := binary.BigEndian.Uint64(iter.Key())

	return last, nil
}

func (s *store) firstIndexTx(tx *leveldb.Transaction) (uint64, error) {
	iter := tx.NewIterator(util.BytesPrefix([]byte{logPrefix}), nil)

	defer iter.Release()

	if exists := iter.First(); !exists {
		return 0, leveldb.ErrNotFound
	}

	last := binary.BigEndian.Uint64(iter.Key())

	return last, nil

}

func (s *store) lastIndexTx(tx *leveldb.Transaction) (uint64, error) {
	iter := tx.NewIterator(nil, nil)

	defer iter.Release()

	if exists := iter.Last(); !exists {
		return 0, leveldb.ErrNotFound
	}

	last := binary.BigEndian.Uint64(iter.Key())

	return last, nil
}

func (s *store) snapshot(snap *leveldb.Snapshot) (raftpb.Snapshot, error) {
	b, err := snap.Get([]byte{snapshotKey}, nil)
	if err != nil {
		return raftpb.Snapshot{}, nil
	}
	var ss raftpb.Snapshot
	err = ss.Unmarshal(b)
	if err != nil {
		return raftpb.Snapshot{}, err
	}
	return ss, nil
}

func (s *store) snapshotTx(tx *leveldb.Transaction) (raftpb.Snapshot, error) {
	b, err := tx.Get([]byte{snapshotKey}, nil)
	if err != nil {
		return raftpb.Snapshot{}, nil
	}
	var ss raftpb.Snapshot
	err = ss.Unmarshal(b)
	if err != nil {
		return raftpb.Snapshot{}, err
	}
	return ss, nil
}

func (s *store) hardState(snap *leveldb.Snapshot) (raftpb.HardState, error) {
	b, err := snap.Get([]byte{hardStateKey}, nil)
	if err != nil {
		return raftpb.HardState{}, nil
	}
	var ss raftpb.HardState
	err = ss.Unmarshal(b)
	if err != nil {
		return raftpb.HardState{}, err
	}
	return ss, nil
}

func (s *store) hardStateTx(tx *leveldb.Transaction) (raftpb.HardState, error) {
	b, err := tx.Get([]byte{hardStateKey}, nil)
	if err != nil {
		return raftpb.HardState{}, nil
	}
	var ss raftpb.HardState
	err = ss.Unmarshal(b)
	if err != nil {
		return raftpb.HardState{}, err
	}
	return ss, nil
}

// Snapshot implements the Storage interface.
func (s *store) Snapshot() (raftpb.Snapshot, error) {
	snap, err := s.internal.GetSnapshot()
	if err != nil {
		return raftpb.Snapshot{}, err
	}
	defer snap.Release()
	snapshotBin, err := snap.Get([]byte{snapshotKey}, nil)
	if err != nil {
		return raftpb.Snapshot{}, err
	}
	var snapshot raftpb.Snapshot
	err = snapshot.Unmarshal(snapshotBin)
	if err != nil {
		return raftpb.Snapshot{}, err
	}
	return snapshot, nil
}

// ApplySnapshot overwrites the contents of this Storage object with
// those of the given snapshot.
func (s *store) ApplySnapshot(snap raftpb.Snapshot) error {
	tx, err := s.internal.OpenTransaction()
	if err != nil {
		return err
	}

	//handle check for old snapshot being applied
	snapshot, err := s.snapshotTx(tx)
	if err != nil {
		tx.Discard()
		return err
	}
	storeIndex := snapshot.Metadata.Index
	snapIndex := snap.Metadata.Index
	if storeIndex >= snapIndex {
		tx.Discard()
		return raft.ErrSnapOutOfDate
	}

	snapshotBin, err := snap.Marshal()
	if err != nil {
		tx.Discard()
		return err
	}
	err = tx.Put([]byte{snapshotKey}, snapshotBin, nil)
	if err != nil {
		tx.Discard()
		return err
	}

	indexBin := make([]byte, 8)
	binary.BigEndian.PutUint64(indexBin, snap.Metadata.Index)
	indexKey := s.logKey(indexBin)

	iter := tx.NewIterator(util.BytesPrefix([]byte{logPrefix}), nil)
	var resultError *multierror.Error
	for iter.Next() {
		err = tx.Delete(iter.Key(), nil)
		if err != nil {
			resultError = multierror.Append(resultError, err)
		}
	}

	iter.Release()
	err = iter.Error()
	if err != nil {
		resultError = multierror.Append(resultError, err)
	}
	if resultError.ErrorOrNil() != nil {
		return resultError.ErrorOrNil()
	}

	ent := raftpb.Entry{Term: snap.Metadata.Term, Index: snap.Metadata.Index}
	value, err := ent.Marshal()
	if err != nil {
		tx.Discard()
		return err
	}
	err = tx.Put(indexKey, value, nil)
	if err != nil {
		tx.Discard()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Discard()
		return err
	}
	return nil
}

// CreateSnapshot makes a snapshot which can be retrieved with Snapshot() and
// can be used to reconstruct the state at that point.
// If any configuration changes have been made since the last compaction,
// the result of the last ApplyConfChange must be passed in.
func (s *store) CreateSnapshot(i uint64, cs *raftpb.ConfState, data []byte) (raftpb.Snapshot, error) {
	tx, err := s.internal.OpenTransaction()
	if err != nil {
		return raftpb.Snapshot{}, err
	}

	snapshotBin, err := tx.Get([]byte{snapshotKey}, nil)
	if err != nil {
		tx.Discard()
		return raftpb.Snapshot{}, nil
	}
	var snapshot raftpb.Snapshot
	err = snapshot.Unmarshal(snapshotBin)
	if err != nil {
		tx.Discard()
		return raftpb.Snapshot{}, nil
	}

	if i <= snapshot.Metadata.Index {
		tx.Discard()
		return raftpb.Snapshot{}, raft.ErrSnapOutOfDate
	}

	lastIndex, err := s.lastIndexTx(tx)

	if i > lastIndex {
		tx.Discard()
		log.Panicf("snapshot %d is out of bound lastindex(%d)", i, lastIndex)
	}
	indexBin := make([]byte, 8)
	binary.BigEndian.PutUint64(indexBin, i)
	indexKey := s.logKey(indexBin)

	entBin, err := tx.Get(indexKey, nil)

	var ent raftpb.Entry

	err = ent.Unmarshal(entBin)

	if err != nil {
		tx.Discard()
		return raftpb.Snapshot{}, err
	}

	snapshot.Metadata.Index = i
	snapshot.Metadata.Term = ent.Term
	if cs != nil {
		snapshot.Metadata.ConfState = *cs
	}
	snapshot.Data = data
	snapshotBin, err = snapshot.Marshal()
	if err != nil {
		tx.Discard()
		return raftpb.Snapshot{}, nil
	}
	tx.Put([]byte{snapshotKey}, snapshotBin, nil)
	err = tx.Commit()
	if err != nil {
		tx.Discard()
		return raftpb.Snapshot{}, nil
	}

	return snapshot, nil
}

// Compact discards all log entries prior to compactIndex.
// It is the application's responsibility to not attempt to compact an index
// greater than raftLog.applied.
func (s *store) Compact(compactIndex uint64) error {

	tx, err := s.internal.OpenTransaction()
	if err != nil {
		return err
	}

	firstIndex, err := s.firstIndexTx(tx)
	if err != nil {
		tx.Discard()
		return err
	}

	if compactIndex <= firstIndex {
		tx.Discard()
		return raft.ErrCompacted
	}

	lastIndex, err := s.lastIndexTx(tx)
	if err != nil {
		tx.Discard()
		return err
	}

	if compactIndex > lastIndex {
		tx.Discard()
		log.Panicf("compact %d is out of bound lastindex(%d)",
			compactIndex, lastIndex)
	}
	compactIndexBin := make([]byte, 8)
	binary.BigEndian.PutUint64(compactIndexBin, compactIndex)
	compactIndexKey := s.logKey(compactIndexBin)
	iter := tx.NewIterator(&util.Range{Limit: compactIndexKey}, nil)
	for iter.Next() {
		tx.Delete(iter.Key(), nil)
	}
	err = tx.Commit()
	if err != nil {
		tx.Discard()
		return err
	}
	return nil
}

// Append the new entries to storage.
// TODO (xiangli): ensure the entries are continuous and
// entries[0].Index > ms.entries[0].Index
func (s *store) Append(entries []raftpb.Entry) error {
	if len(entries) == 0 {
		return nil
	}

	tx, err := s.internal.OpenTransaction()
	if err != nil {
		return err
	}

	first, err := s.firstIndexTx(tx)
	if err != nil {
		tx.Discard()
		return err
	}
	last := entries[0].Index + uint64(len(entries)) - 1

	// shortcut if there is no new entry.
	if last < first {
		tx.Discard()
		return nil
	}
	// truncate compacted entries
	if first > entries[0].Index {
		entries = entries[first-entries[0].Index:]
	}

	for i := range entries {
		indexBin := make([]byte, 8)
		binary.BigEndian.PutUint64(indexBin, entries[i].Index)
		indexKey := s.logKey(indexBin)
		value, err := entries[i].Marshal()
		if err != nil {
			tx.Discard()
			return err
		}
		tx.Put(indexKey, value, nil)
	}

	err = tx.Commit()
	if err != nil {
		tx.Discard()
		return err
	}
	return nil
}
