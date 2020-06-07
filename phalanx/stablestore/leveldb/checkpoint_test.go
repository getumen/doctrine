package leveldbstablestore

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"
)

func TestPhalanxDB_Checkpoint(t *testing.T) {
	tempDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tempDir) })
	internal, err := leveldb.OpenFile(tempDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	internal.Put([]byte("foo"), []byte("bar"), nil)

	s := &store{
		internal: internal,
	}

	defer s.Close()

	snap, err := s.GetSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	defer snap.Release()

	v, err := snap.Get([]byte("foo"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(v, []byte("bar")) != 0 {
		t.Fatalf("foo has unexpected value, got %s", v)
	}

	data, err := s.CreateCheckpoint()
	if err != nil {
		t.Fatal(err)
	}
	err = s.internal.Delete([]byte("foo"), nil)
	if err != nil {
		t.Fatal(err)
	}
	snap, err = s.GetSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	v, err = snap.Get([]byte("foo"), nil)
	if err != leveldb.ErrNotFound {
		t.Fatal(err)
	}

	if err := s.RestoreToCheckpoint(data); err != nil {
		t.Fatal(err)
	}
	snap, err = s.GetSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	v, err = snap.Get([]byte("foo"), nil)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(v, []byte("bar")) != 0 {
		t.Fatalf("foo has unexpected value, got %s", v)
	}
}
