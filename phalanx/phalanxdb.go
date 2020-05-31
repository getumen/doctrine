package phalanx

import (
	"log"

	"github.com/coreos/etcd/snap"
	"github.com/getumen/doctrine/phalanx/phalanxpb"
	"google.golang.org/protobuf/proto"
)

type phananxDB struct {
	proposeC      chan<- []byte // channel for proposing updates
	stableStore   StableStore
	commandHander CommandHandler
	snapshotter   *snap.Snapshotter
}

func (db *phananxDB) Get(key []byte, ro *ReadOptions) ([]byte, error) {
	snapshot, err := db.stableStore.GetSnapshot()
	if err != nil {
		return nil, err
	}
	defer snapshot.Release()
	return snapshot.Get(key, ro)
}

func (db *phananxDB) Propose(command *phalanxpb.Command) error {
	message, err := proto.Marshal(command)
	if err != nil {
		return err
	}
	db.proposeC <- message
	return nil
}

func (db *phananxDB) readCommits(commitC chan []byte, errorC chan error) error {
	for data := range commitC {
		if data == nil {
			// done replaying log; new data incoming
			// OR signaled to load snapshot
			snapshot, err := db.snapshotter.Load()
			if err == snap.ErrNoSnapshot {
				return nil
			}
			if err != nil {
				log.Panic(err)
			}
			log.Printf("loading snapshot at term %d and index %d",
				snapshot.Metadata.Term, snapshot.Metadata.Index)
			if err := db.recoverFromSnapshot(snapshot.Data); err != nil {
				log.Panic(err)
			}
			continue
		}

		var command phalanxpb.Command
		err := proto.Unmarshal(data, &command)
		if err != nil {
			errorC <- err
			continue
		}
		db.commandHander.Apply(&command, db.stableStore)
	}
	if err, ok := <-errorC; ok {
		return err
	}
	return nil
}

func (db *phananxDB) getSnapshot() ([]byte, error) {
	return db.stableStore.CreateCheckpoint()
}

func (db *phananxDB) recoverFromSnapshot(snapshot []byte) error {
	return db.stableStore.RestoreToCheckpoint(snapshot)
}
