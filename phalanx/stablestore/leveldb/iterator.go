package leveldb

import (
	itpkg "github.com/syndtr/goleveldb/leveldb/iterator"
)

type iterator struct {
	internal itpkg.Iterator
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

func (it *iterator) First() bool {
	return it.internal.First()
}

func (it *iterator) Last() bool {
	return it.internal.Last()
}

func (it *iterator) Seek(key []byte) bool {
	return it.internal.Seek(key)
}

func (it *iterator) Prev() bool {
	return it.internal.Prev()
}

func (it *iterator) Next() bool {
	return it.internal.Next()
}
