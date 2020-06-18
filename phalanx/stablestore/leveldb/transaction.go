package leveldbstablestore

import (
	"github.com/getumen/doctrine/phalanx"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type transaction struct {
	internal *leveldb.Transaction
}

func (tx *transaction) Commit() error {
	return tx.internal.Commit()
}

func (tx *transaction) Delete(key []byte) error {

	return tx.internal.Delete(key, nil)
}

func (tx *transaction) Discard() {
	tx.internal.Discard()
}

func (tx *transaction) Get(key []byte) ([]byte, error) {

	return tx.internal.Get(key, nil)
}

func (tx *transaction) Has(key []byte) (bool, error) {
	return tx.internal.Has(key, nil)
}

func (tx *transaction) NewIterator(slice *phalanx.Range) phalanx.Iterator {

	return &iterator{
		internal: tx.internal.NewIterator(
			&util.Range{
				Start: slice.Start,
				Limit: slice.End,
			},
			&opt.ReadOptions{
				DontFillCache: true,
			},
		),
	}
}

func (tx *transaction) Put(key, value []byte) error {

	return tx.internal.Put(key, value, nil)
}

func (tx *transaction) Write(b phalanx.Batch) error {

	if ba, ok := b.(*batch); ok {
		return tx.internal.Write(ba.internal, nil)
	}
	return errors.Errorf("cast error: %v", b)
}

func (tx *transaction) CreateBatch() phalanx.Batch {
	return &batch{
		internal: new(leveldb.Batch),
	}
}
