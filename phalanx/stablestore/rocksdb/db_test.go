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

func TestRocksDB_Compression(t *testing.T) {

	cases := []struct {
		name            string
		compressionType gorocksdb.CompressionType
	}{
		{
			name:            "no",
			compressionType: gorocksdb.NoCompression,
		},
		{
			name:            "snappy",
			compressionType: gorocksdb.SnappyCompression,
		},
		{
			name:            "zlib",
			compressionType: gorocksdb.ZLibCompression,
		},
		{
			name:            "zstd",
			compressionType: gorocksdb.ZSTDCompression,
		},
		{
			name:            "bz2",
			compressionType: gorocksdb.Bz2Compression,
		},
		{
			name:            "lz4",
			compressionType: gorocksdb.LZ4Compression,
		},
		{
			name:            "lz4hc",
			compressionType: gorocksdb.LZ4HCCompression,
		},
	}

	for _, c := range cases {

		tempDir, err := ioutil.TempDir("", t.Name())
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.RemoveAll(tempDir) })

		opt := gorocksdb.NewDefaultOptions()
		opt.SetCompression(c.compressionType)
		opt.SetCreateIfMissing(true)

		db, err := gorocksdb.OpenDb(opt, tempDir)
		if err != nil {
			t.Fatalf("[%s] fail to create: %+v", c.name, err)
		}

		wo := gorocksdb.NewDefaultWriteOptions()

		err = db.Put(wo, []byte("test"), []byte("value"))
		if err != nil {
			t.Fatalf("[%s] fail to put: %+v", c.name, err)
		}

		ro := gorocksdb.NewDefaultReadOptions()

		sl, err := db.Get(ro, []byte("test"))
		defer sl.Free()
		val := sl.Data()

		if bytes.Compare(val, []byte("value")) != 0 {
			t.Fatalf("[%s] expected value, but got %s",
				c.name, string(val))
		}

	}
}
