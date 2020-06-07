package leveldbstablestore

import (
	"testing"

	"github.com/getumen/doctrine/phalanx"
)

func TestStableStoreImplementation(t *testing.T) {
	var target interface{} = new(store)
	if _, ok := target.(phalanx.StableStore); !ok {
		t.Fatalf("store implementation is incomplele")
	}
}

func TestBatchImplementation(t *testing.T) {
	var target interface{} = new(batch)
	if _, ok := target.(phalanx.Batch); !ok {
		t.Fatalf("batch implementation is incomplele")
	}
}

func TestSnapshotImplementation(t *testing.T) {
	var target interface{} = new(snapshot)
	if _, ok := target.(phalanx.Snapshot); !ok {
		t.Fatalf("snapshot implementation is incomplele")
	}
}

func TestIteratorImplementation(t *testing.T) {
	var target interface{} = new(iterator)
	if _, ok := target.(phalanx.Iterator); !ok {
		t.Fatalf("iterator implementation is incomplele")
	}
}

func TestTransactionImplementation(t *testing.T) {
	var target interface{} = new(transaction)
	if _, ok := target.(phalanx.Transaction); !ok {
		t.Fatalf("transaction implementation is incomplele")
	}
}
