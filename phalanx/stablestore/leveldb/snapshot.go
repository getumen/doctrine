package leveldb

import (
	"github.com/getumen/doctrine/phalanx"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"golang.org/x/xerrors"
)

type snapshot struct {
	// read only snapshots
	regionSnaps map[string]*leveldb.Snapshot
}

func (snap *snapshot) Get(region string, key []byte) (value []byte, err error) {
	if sn, exists := snap.regionSnaps[region]; exists {
		v, err := sn.Get(key, nil)
		if err == leveldb.ErrNotFound {
			return nil, phalanx.ErrKeyNotFound
		} else if err != nil {
			return nil, xerrors.Errorf("leveldb stable store: %w", err)
		}
		return v, nil
	}
	return nil, phalanx.NewRegionNotFound(region)
}

func (snap *snapshot) MultiGet(region string, keys ...[]byte) ([][]byte, error) {
	if sn, exists := snap.regionSnaps[region]; exists {

		values := make([][]byte, len(keys))

		for i := range keys {
			v, err := sn.Get(keys[i], nil)
			if err == leveldb.ErrNotFound {
				values[i] = nil
			} else if err != nil {
				values[i] = v
			}
		}
		return values, nil
	}
	return nil, phalanx.NewRegionNotFound(region)

}

func (snap *snapshot) Has(region string, key []byte) (ret bool, err error) {
	if sn, exists := snap.regionSnaps[region]; exists {
		return sn.Has(key, nil)
	}
	return false, phalanx.NewRegionNotFound(region)
}

func (snap *snapshot) NewIterator(
	region string,
	slice *phalanx.Range,
) (phalanx.Iterator, error) {
	if slice == nil {
		slice = phalanx.FullScanRange()
	}
	if sn, exists := snap.regionSnaps[region]; exists {
		return &iterator{
			internal: sn.NewIterator(
				&util.Range{
					Start: slice.Start,
					Limit: slice.End,
				},
				&opt.ReadOptions{DontFillCache: true},
			),
		}, nil
	}
	return nil, phalanx.NewRegionNotFound(region)
}

func (snap *snapshot) Release() {
	for key := range snap.regionSnaps {
		snap.regionSnaps[key].Release()
	}
}
