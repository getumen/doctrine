package phalanx

import (
	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
)

// LogStore is a raft log storage that supports creating snapshot
type LogStore interface {
	raft.Storage
	ApplySnapshot(snap raftpb.Snapshot) error
	SetHardState(st raftpb.HardState) error
	Append(entries []raftpb.Entry) error
	Compact(compactIndex uint64) error
	CreateSnapshot(i uint64, cs *raftpb.ConfState, data []byte) (raftpb.Snapshot, error)
}
