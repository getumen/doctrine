package rocksdb

import (
	"sync"

	"github.com/tecbot/gorocksdb"
)

// batch is batchs partitioned by region
// batch is not thread safe
type batch struct {
	cfMutex *sync.RWMutex
	cf      map[string]*gorocksdb.ColumnFamilyHandle
	batchs  *gorocksdb.WriteBatch
}

func (b *batch) Put(region string, key, value []byte) {
	b.cfMutex.RLock()
	defer b.cfMutex.RUnlock()
	b.put(region, key, value)
}

func (b *batch) put(region string, key, value []byte) {
	b.batchs.PutCF(b.cf[region], key, value)
}

func (b *batch) Delete(region string, key []byte) {
	b.cfMutex.RLock()
	defer b.cfMutex.RUnlock()
	b.delete(region, key)
}

func (b *batch) delete(region string, key []byte) {
	b.batchs.DeleteCF(b.cf[region], key)
}

func (b *batch) Len() int {
	return b.batchs.Count()
}

func (b *batch) Reset() {
	b.batchs.Clear()
}
