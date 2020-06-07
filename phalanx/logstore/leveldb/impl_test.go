package leveldblogstore

import (
	"testing"

	"github.com/getumen/doctrine/phalanx"
)

func TestLogStoreImplementation(t *testing.T) {
	var target interface{} = new(store)
	if _, ok := target.(phalanx.LogStore); !ok {
		t.Fatalf("store implementation is incomplele")
	}
}
