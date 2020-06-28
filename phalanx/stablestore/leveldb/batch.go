package leveldb

import "github.com/syndtr/goleveldb/leveldb"

// batch is batchs partitioned by region
// batch is not thread safe
type batch struct {
	batchs map[string]*leveldb.Batch
}

func (b *batch) Put(region string, key, value []byte) {
	b.batchs[region].Put(key, value)
}

func (b *batch) Delete(region string, key []byte) {
	b.batchs[region].Delete(key)
}

func (b *batch) Len() int {
	l := 0
	for _, v := range b.batchs {
		l += v.Len()
	}
	return l
}

func (b *batch) Reset() {
	for _, v := range b.batchs {
		v.Reset()
	}
}
