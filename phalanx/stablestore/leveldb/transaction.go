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

func (tx *transaction) Delete(key []byte, wo *phalanx.WriteOptions) error {
	if wo == nil {
		wo = phalanx.DefaultWriteOptions()
	}

	return tx.internal.Delete(
		key,
		&opt.WriteOptions{
			Sync: wo.Sync,
		},
	)
}

func (tx *transaction) Discard() {
	tx.internal.Discard()
}

func (tx *transaction) Get(key []byte, ro *phalanx.ReadOptions) ([]byte, error) {
	if ro == nil {
		ro = phalanx.DefaultReadOptions()
	}

	return tx.internal.Get(key, &opt.ReadOptions{
		DontFillCache: !ro.FillCache,
	})
}

func (tx *transaction) Has(key []byte, ro *phalanx.ReadOptions) (bool, error) {
	if ro == nil {
		ro = phalanx.DefaultReadOptions()
	}

	return tx.internal.Has(key, &opt.ReadOptions{
		DontFillCache: !ro.FillCache,
	})
}

func (tx *transaction) NewIterator(slice *phalanx.Range, ro *phalanx.ReadOptions) phalanx.Iterator {
	if ro == nil {
		ro = phalanx.DefaultReadOptions()
	}

	return &iterator{
		internal: tx.internal.NewIterator(
			&util.Range{
				Start: slice.Start,
				Limit: slice.End,
			},
			&opt.ReadOptions{
				DontFillCache: !ro.FillCache,
			},
		),
	}
}

func (tx *transaction) Put(key, value []byte, wo *phalanx.WriteOptions) error {
	if wo == nil {
		wo = phalanx.DefaultWriteOptions()
	}

	return tx.internal.Put(key, value, &opt.WriteOptions{
		Sync: wo.Sync,
	})
}

func (tx *transaction) Write(b phalanx.Batch, wo *phalanx.WriteOptions) error {
	if wo == nil {
		wo = phalanx.DefaultWriteOptions()
	}

	if ba, ok := b.(*batch); ok {
		return tx.internal.Write(ba.internal, &opt.WriteOptions{
			Sync: wo.Sync,
		})
	}
	return errors.Errorf("cast error: %v", b)
}

func (tx *transaction) CreateBatch() phalanx.Batch {
	return &batch{
		internal: new(leveldb.Batch),
	}
}
