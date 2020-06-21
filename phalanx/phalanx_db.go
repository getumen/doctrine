package phalanx

import (
	"log"

	"github.com/coreos/etcd/snap"
	"github.com/getumen/doctrine/phalanx/phalanxpb"
	"google.golang.org/protobuf/proto"
)

// DB is distributed embeddable db
type DB interface {
	Get(key []byte) ([]byte, error)
	Propose(command *phalanxpb.Command) error
}

type phananxDB struct {
	regionName    string
	proposeC      chan<- []byte // channel for proposing updates
	stableStore   StableStore
	commandHander CommandHandler
	snapshotter   *snap.Snapshotter
}

// NewDB creates new db
func NewDB(
	regionName string,
	snapshotter *snap.Snapshotter,
	proposeC chan []byte,
	commitC chan []byte,
	errorC chan error,
	stableStore StableStore,
	commandHander CommandHandler,
) DB {
	db := &phananxDB{
		regionName:    regionName,
		proposeC:      proposeC,
		stableStore:   stableStore,
		commandHander: commandHander,
		snapshotter:   snapshotter,
	}
	// replay log into key-value map
	db.readCommits(commitC, errorC)
	// read commits from raft into kvStore map until error
	go db.readCommits(commitC, errorC)

	return db
}

func (db *phananxDB) Get(key []byte) ([]byte, error) {
	snapshot, err := db.stableStore.GetSnapshot()
	if err != nil {
		return nil, err
	}
	defer snapshot.Release()
	return snapshot.Get(db.regionName, key)
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
		db.commandHander.Apply(db.regionName, &command, db.stableStore)
	}
	if err, ok := <-errorC; ok {
		return err
	}
	return nil
}

func (db *phananxDB) GetSnapshot() ([]byte, error) {
	return db.stableStore.CreateCheckpoint(db.regionName)
}

func (db *phananxDB) recoverFromSnapshot(snapshot []byte) error {
	return db.stableStore.RestoreToCheckpoint(db.regionName, snapshot)
}
