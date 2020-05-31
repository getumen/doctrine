package leveldb

import (
	"github.com/getumen/doctrine/phalanx"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type snapshot struct {
	internal *leveldb.Snapshot
}

func (snap *snapshot) Get(key []byte, ro *phalanx.ReadOptions) (value []byte, err error) {
	return snap.internal.Get(key, &opt.ReadOptions{
		DontFillCache: !ro.FillCache,
	})
}

func (snap *snapshot) Has(key []byte, ro *phalanx.ReadOptions) (ret bool, err error) {
	return snap.internal.Has(key, &opt.ReadOptions{
		DontFillCache: !ro.FillCache,
	})
}

func (snap *snapshot) NewIterator(
	slice *phalanx.Range,
	ro *phalanx.ReadOptions) phalanx.Iterator {
	return &iterator{
		internal: snap.internal.NewIterator(
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

func (snap *snapshot) Release() {
	snap.internal.Release()
}
