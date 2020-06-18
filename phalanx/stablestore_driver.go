package phalanx

import "sync"

// StableStoreDriver is driver of stable store
type StableStoreDriver interface {
	New(path string) (StableStore, error)
}

var (
	stableStoreDroverLock sync.RWMutex
	stableStoreDrivers    map[string]StableStoreDriver = make(map[string]StableStoreDriver)
)

// RegisterStableStore registers the given driver
func RegisterStableStore(
	driverName string,
	driver StableStoreDriver) {
	stableStoreDroverLock.Lock()
	defer stableStoreDroverLock.Unlock()
	if _, dup := stableStoreDrivers[driverName]; dup {
		panic("sql: Register called twice for driver " + driverName)
	}
	stableStoreDrivers[driverName] = driver
}

// NewStableStore creates new stable store
func NewStableStore(name string, path string) (StableStore, error) {
	stableStoreDroverLock.RLock()
	defer stableStoreDroverLock.RUnlock()
	if driver, ok := stableStoreDrivers[name]; ok {
		return driver.New(path)
	}
	return nil, &ErrStableStoreDriverNotFound{driverName: name}
}
