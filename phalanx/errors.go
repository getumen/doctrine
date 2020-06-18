package phalanx

import "fmt"

// ErrStableStoreDriverNotFound is t/o
type ErrStableStoreDriverNotFound struct {
	driverName string
}

func (errStableStoreDriverNotFound *ErrStableStoreDriverNotFound) Error() string {
	return fmt.Sprintf("stable store driver '%s' not found",
		errStableStoreDriverNotFound.driverName)
}

// ErrLogStoreDriverNotFound is t/o
type ErrLogStoreDriverNotFound struct {
	driverName string
}

func (errLogStoreDriverNotFound *ErrLogStoreDriverNotFound) Error() string {
	return fmt.Sprintf("log store driver '%s' not found",
		errLogStoreDriverNotFound.driverName)
}
