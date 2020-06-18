package leveldbstablestore

import "github.com/syndtr/goleveldb/leveldb"

type batch struct {
	internal *leveldb.Batch
}

func (b *batch) Put(key, value []byte) {
	b.internal.Put(key, value)
}

func (b *batch) Delete(key []byte) {
	b.internal.Delete(key)
}

func (b *batch) Len() int {
	return b.internal.Len()
}

func (b *batch) Reset() {
	b.internal.Reset()
}
