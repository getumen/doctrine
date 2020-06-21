package leveldbstablestore

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
)

func TestStore_Checkpoint(t *testing.T) {

	const region = "default"

	inputs := []struct {
		key, value []byte
	}{
		{
			key:   []byte("a"),
			value: []byte("0"),
		},
		{
			key:   []byte("a"),
			value: []byte("1"),
		},
		{
			key:   []byte("b"),
			value: []byte("0"),
		},
	}

	expected := []struct {
		key, value []byte
	}{
		{
			key:   []byte("a"),
			value: []byte("1"),
		},
		{
			key:   []byte("b"),
			value: []byte("0"),
		},
	}

	tempDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tempDir) })
	target := &store{
		storages: make(map[string]*leveldb.DB),
		dataPath: tempDir,
	}
	t.Cleanup(func() { target.Close() })

	err = target.CreateRegion(region)

	batch := target.CreateBatch()
	for i := range inputs {
		batch.Put(region, inputs[i].key, inputs[i].value)
	}
	err = target.Write(batch)
	if err != nil {
		t.Fatal(err)
	}

	checkpoint, err := target.CreateCheckpoint(region)
	if err != nil {
		t.Fatal(err)
	}

	tempDir2, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tempDir2) })

	actual := &store{
		storages: make(map[string]*leveldb.DB),
		dataPath: tempDir2,
	}
	t.Cleanup(func() { actual.Close() })

	err = actual.RestoreToCheckpoint(region, checkpoint)
	if err != nil {
		t.Fatal(err)
	}

	snap, err := actual.GetSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	defer snap.Release()

	iter, err := snap.NewIterator(region, nil)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	defer iter.Release()

	counter := 0

	for iter.Next() {
		if bytes.Compare(expected[counter].key, iter.Key()) != 0 {
			t.Fatalf("keys not match: expected %s, but got %s",
				string(expected[counter].key), string(iter.Key()))
		}
		if bytes.Compare(expected[counter].value, iter.Value()) != 0 {
			t.Fatalf("valuess not match: expected %s, but got %s",
				string(expected[counter].value), string(iter.Value()))
		}

		counter++
	}

}
