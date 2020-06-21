package leveldbstablestore

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

func (b *batch) Len(region string) int {
	return b.batchs[region].Len()
}

func (b *batch) Reset(region string) {
	b.batchs[region].Reset()
}
