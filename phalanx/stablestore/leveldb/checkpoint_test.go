package leveldb

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/getumen/doctrine/phalanx"
)

func TestPhalanxDB_Checkpoint(t *testing.T) {

	const region = "default"
	tempDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	driver := &storeDriver{}

	s, err := driver.New(tempDir)

	if err != nil {
		t.Fatalf("error: %+v", err)
	}

	defer s.Close()

	err = s.CreateRegion(region)
	if err != nil {
		t.Fatalf("error: %+v", err)
	}

	b := s.CreateBatch()
	b.Put(region, []byte("foo"), []byte("bar"))
	err = s.Write(b)

	if err != nil {
		t.Fatalf("error: %+v", err)
	}

	snap, err := s.GetSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	defer snap.Release()

	v, err := snap.Get(region, []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, []byte("bar")) {
		t.Fatalf("foo has unexpected value, got %s", v)
	}

	data, err := s.CreateCheckpoint(region)
	if err != nil {
		t.Fatal(err)
	}

	ba := s.CreateBatch()
	ba.Delete(region, []byte("foo"))
	err = s.Write(ba)
	if err != nil {
		t.Fatal(err)
	}
	snap, err = s.GetSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	_, err = snap.Get(region, []byte("foo"))
	if err != phalanx.ErrKeyNotFound {
		t.Fatal(err)
	}

	if err := s.RestoreToCheckpoint(region, data); err != nil {
		t.Fatal(err)
	}
	snap, err = s.GetSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	v, err = snap.Get(region, []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(v, []byte("bar")) {
		t.Fatalf("foo has unexpected value, got %s", v)
	}
}
