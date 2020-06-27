package rocksdb

import (
	"github.com/tecbot/gorocksdb"
)

type iterator struct {
	internal *gorocksdb.Iterator
}
