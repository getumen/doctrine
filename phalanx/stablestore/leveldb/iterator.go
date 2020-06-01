package leveldb

import (
	itpkg "github.com/syndtr/goleveldb/leveldb/iterator"
)

type iterator struct {
	internal itpkg.Iterator
}

func (it *iterator) Next() bool {
	return it.internal.Next()
}

func (it *iterator) Key() []byte {
	return it.internal.Key()
}

func (it *iterator) Value() []byte {
	return it.internal.Value()
}

func (it *iterator) Release() {
	it.internal.Release()
}

func (it *iterator) Error() error {
	return it.internal.Error()
}
