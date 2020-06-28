package rocksdb

import (
	"bytes"

	"github.com/tecbot/gorocksdb"
)

type iterator struct {
	internal   *gorocksdb.Iterator
	start, end []byte
}

func (it *iterator) Key() []byte {
	sl := it.internal.Key()
	defer sl.Free()
	return sl.Data()
}

func (it *iterator) Value() []byte {
	sl := it.internal.Value()
	defer sl.Free()
	return sl.Data()
}

func (it *iterator) Release() {
	it.internal.Close()
}

func (it *iterator) Error() error {
	return it.internal.Err()
}

func (it *iterator) First() bool {
	it.internal.SeekToFirst()
	return it.internal.Valid()
}

func (it *iterator) Last() bool {
	it.internal.SeekToLast()
	return it.internal.Valid()
}

func (it *iterator) Seek(key []byte) bool {
	current := it.Key()
	cmp := bytes.Compare(key, current)
	if cmp == 0 {
		return true
	} else if cmp < 0 {
		it.internal.SeekForPrev(key)
		return it.internal.Valid()
	} else {
		it.internal.Seek(key)
		return it.internal.Valid()
	}
}

func (it *iterator) Next() bool {
	if !it.internal.Valid() {
		return false
	}
	it.internal.Next()
	return it.internal.Valid()
}

func (it *iterator) Prev() bool {
	it.internal.Prev()
	return it.internal.Valid()
}
