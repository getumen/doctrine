package rocksdb

import (
	"sync"

	"github.com/getumen/doctrine/phalanx"
	"github.com/tecbot/gorocksdb"
	"golang.org/x/xerrors"
)

type snapshot struct {
	snap    *gorocksdb.Snapshot
	db      *gorocksdb.DB
	cfMutex *sync.RWMutex
	cf      map[string]*gorocksdb.ColumnFamilyHandle
}

func (s *snapshot) Get(region string, key []byte) (value []byte, err error) {
	opt := gorocksdb.NewDefaultReadOptions()
	opt.SetSnapshot(s.snap)

	s.cfMutex.RLock()
	defer s.cfMutex.RUnlock()

	sl, err := s.db.GetCF(opt, s.cf[region], key)
	if err != nil {
		return nil, xerrors.Errorf("rocksdb stable store: %w", err)
	}
	defer sl.Free()

	if !sl.Exists() {
		return nil, phalanx.ErrKeyNotFound
	}
	return sl.Data(), nil
}

func (s *snapshot) Has(region string, key []byte) (ret bool, err error) {
	opt := gorocksdb.NewDefaultReadOptions()
	opt.SetSnapshot(s.snap)

	s.cfMutex.RLock()
	defer s.cfMutex.RUnlock()

	sl, err := s.db.GetCF(opt, s.cf[region], key)
	if err != nil {
		return false, xerrors.Errorf("rocksdb stable store: %w", err)
	}
	defer sl.Free()

	return sl.Exists(), nil
}

func (s *snapshot) NewIterator(region string, slice *phalanx.Range) (phalanx.Iterator, error) {
	s.cfMutex.RLock()
	defer s.cfMutex.RUnlock()
	return s.newIterator(region, slice)
}

// newIterator is an iterator of the region
// only ascendant order is supported
func (s *snapshot) newIterator(region string, slice *phalanx.Range) (phalanx.Iterator, error) {

	ro := gorocksdb.NewDefaultReadOptions()
	ro.SetFillCache(false)
	ro.SetSnapshot(s.snap)
	if slice != nil && slice.End != nil {
		ro.SetIterateUpperBound(slice.End)
	}
	it := s.db.NewIteratorCF(ro, s.cf[region])
	if slice != nil && slice.Start != nil {
		it.Seek(slice.Start)
	}
	return &iterator{
		internal: it,
	}, nil
}

func (s *snapshot) Release() {
	s.db.ReleaseSnapshot(s.snap)
}
