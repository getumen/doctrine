package phalanx

import (
	"errors"
	"fmt"
)

var (
	// ErrKeyNotFound represnts that the key is not found in the stable store
	ErrKeyNotFound = errors.New("not found")
)

// ErrStableStoreDriverNotFound is T/O
type ErrStableStoreDriverNotFound struct {
	DriverName string
}

func (e *ErrStableStoreDriverNotFound) Error() string {
	return fmt.Sprintf("stable store driver '%s' not found",
		e.DriverName)
}

// ErrLogStoreDriverNotFound is T/O
type ErrLogStoreDriverNotFound struct {
	DriverName string
}

func (e *ErrLogStoreDriverNotFound) Error() string {
	return fmt.Sprintf("log store driver '%s' not found",
		e.DriverName)
}

// ErrRegionAlreadyExists is T/O
type ErrRegionAlreadyExists struct {
	region string
}

// NewErrRegionAlreadyExists creates ErrRegionAlreadyExists
func NewErrRegionAlreadyExists(region string) *ErrRegionAlreadyExists {
	return &ErrRegionAlreadyExists{
		region: region,
	}
}

func (e *ErrRegionAlreadyExists) Error() string {
	return fmt.Sprintf("region '%s' already exists",
		e.region)
}

// ErrRegionNotFound is T/O
type ErrRegionNotFound struct {
	region string
}

// NewRegionNotFound creates ErrRegionNotFound
func NewRegionNotFound(region string) *ErrRegionNotFound {
	return &ErrRegionNotFound{
		region: region,
	}
}

func (e *ErrRegionNotFound) Error() string {
	return fmt.Sprintf("region '%s' not found",
		e.region)
}
