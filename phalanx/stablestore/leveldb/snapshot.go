package leveldbstablestore

import (
	"github.com/getumen/doctrine/phalanx"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type snapshot struct {
	regionSnaps map[string]*leveldb.Snapshot
}

func (snap *snapshot) Get(region string, key []byte) (value []byte, err error) {
	if sn, exists := snap.regionSnaps[region]; exists {
		return sn.Get(key, nil)
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
