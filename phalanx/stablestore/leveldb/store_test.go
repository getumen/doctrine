package leveldbstablestore

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
)

func TestStore_Checkpoint(t *testing.T) {

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
	internal, err := leveldb.OpenFile(tempDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	target := &store{
		internal: internal,
	}
	t.Cleanup(func() { target.Close() })

	batch := target.CreateBatch()
	for i := range inputs {
		batch.Put(inputs[i].key, inputs[i].value)
	}
	err = target.Write(batch)
	if err != nil {
		t.Fatal(err)
	}

	checkpoint, err := target.CreateCheckpoint()
	if err != nil {
		t.Fatal(err)
	}

	tempDir2, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tempDir2) })
	actualInternal, err := leveldb.OpenFile(tempDir2, nil)
	if err != nil {
		t.Fatal(err)
	}
	actual := &store{
		internal: actualInternal,
	}
	t.Cleanup(func() { actual.Close() })

	err = actual.RestoreToCheckpoint(checkpoint)
	if err != nil {
		t.Fatal(err)
	}

	snap, err := actual.GetSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	defer snap.Release()

	iter := snap.NewIterator(nil)
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
