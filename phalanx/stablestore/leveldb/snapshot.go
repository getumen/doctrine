package leveldbstablestore

import (
	"github.com/getumen/doctrine/phalanx"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type snapshot struct {
	internal *leveldb.Snapshot
}

func (snap *snapshot) Get(key []byte) (value []byte, err error) {
	return snap.internal.Get(key, nil)
}

func (snap *snapshot) Has(key []byte) (ret bool, err error) {
	return snap.internal.Has(key, nil)
}

func (snap *snapshot) NewIterator(
	slice *phalanx.Range) phalanx.Iterator {
	if slice == nil {
		slice = phalanx.FullScanRange()
	}
	return &iterator{
		internal: snap.internal.NewIterator(
			&util.Range{
				Start: slice.Start,
				Limit: slice.End,
			},
			&opt.ReadOptions{DontFillCache: true},
		),
	}
}

func (snap *snapshot) Release() {
	snap.internal.Release()
}
