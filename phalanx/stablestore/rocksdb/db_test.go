package rocksdb

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/tecbot/gorocksdb"
)

func TestRocksDB_GetPut(t *testing.T) {
	tempDir, err := ioutil.TempDir("", t.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	opt := gorocksdb.NewDefaultOptions()
	opt.SetCreateIfMissing(true)

	db, err := gorocksdb.OpenDb(opt, tempDir)
	if err != nil {
		t.Fatalf("fail to create: %+v", err)
	}

	wo := gorocksdb.NewDefaultWriteOptions()

	err = db.Put(wo, []byte("test"), []byte("value"))
	if err != nil {
		t.Fatalf("fail to put: %+v", err)
	}

	ro := gorocksdb.NewDefaultReadOptions()

	sl, err := db.Get(ro, []byte("test"))
	defer sl.Free()
	val := sl.Data()

	if bytes.Compare(val, []byte("value")) != 0 {
		t.Fatalf("expected value, but got %s", string(val))
	}
}
