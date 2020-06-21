package leveldbstablestore

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/getumen/doctrine/phalanx"
	"github.com/hashicorp/go-multierror"
	"github.com/linkedin/goavro"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"golang.org/x/xerrors"
)

const schema = `
{
	"namespace": "phalanx.avro",
 	"type": "record",
 	"name": "Checkpoint",
 	"fields": [
    	{"name": "key", "type": "bytes"},
    	{"name": "value",  "type": ["bytes", "null"]}
	]
}
`

const maxBatchSize = 256

const allowedRegionChars = `[0-9A-Za-z_\-]+`

var (
	regionNameRegExp = regexp.MustCompile(allowedRegionChars)
)

type store struct {
	sync.RWMutex
	storages map[string]*leveldb.DB
	dataPath string
}

type storeDriver struct {
}

// New creates stable store implemented by LevelDB
func (d *storeDriver) New(dataPath string) (phalanx.StableStore, error) {
	return &store{
		storages: map[string]*leveldb.DB{},
		dataPath: dataPath,
	}, nil
}

func init() {
	phalanx.RegisterStableStore("leveldb", &storeDriver{})
}

// CreateRegion creates a region
func (s *store) CreateRegion(name string) error {

	if matched := regionNameRegExp.Match([]byte(name)); !matched {
		return errors.Errorf(
			"leveldb stable store: invalid region name (%s) allowed chars are %s",
			name, allowedRegionChars,
		)
	}

	s.Lock()
	defer s.Unlock()
	if _, dup := s.storages[name]; dup {
		return phalanx.NewErrRegionAlreadyExists(name)
	}
	db, err := leveldb.OpenFile(filepath.Join(s.dataPath, name), nil)
	if err != nil {
		return xerrors.Errorf(
			"leveldb stable store: fail to create region(%s): %w",
			name, err)
	}
	s.storages[name] = db
	return nil
}

// DropRegion drop a region
func (s *store) DropRegion(name string) error {
	s.Lock()
	defer s.Unlock()
	if region, exist := s.storages[name]; exist {
		if err := region.Close(); err != nil {
			return xerrors.Errorf(
				"leveldb stable store: fail to close region(%s): %w",
				name, err)
		}

		os.RemoveAll(filepath.Join(s.dataPath, name))
		return nil
	}
	return phalanx.NewRegionNotFound(name)
}

func (s *store) HasRegion(name string) bool {
	s.RLock()
	defer s.RUnlock()
	_, exists := s.storages[name]
	return exists
}

// CreateBatch creates batch
func (s *store) CreateBatch() phalanx.Batch {
	s.RLock()
	defer s.RUnlock()
	batchs := make(map[string]*leveldb.Batch)
	for key := range s.storages {
		batchs[key] = new(leveldb.Batch)
	}
	return &batch{
		batchs: batchs,
	}
}

// Write apply the given batch to the StableStorage
func (s *store) Write(b phalanx.Batch) error {
	s.RLock()
	defer s.RUnlock()
	if bi, ok := b.(*batch); ok {
		// check all region exists
		for key := range bi.batchs {
			if _, exists := s.storages[key]; !exists {
				return phalanx.NewRegionNotFound(key)
			}
		}
		var errs *multierror.Error
		for key := range bi.batchs {
			err := s.storages[key].Write(bi.batchs[key], nil)
			if err != nil {
				errs = multierror.Append(
					errs,
					xerrors.Errorf(
						"leveldb stable store: fail to write batch to region(%s): %w",
						key, err))
			}

		}
		return errs.ErrorOrNil()
	}
	return errors.New("cast fail")
}

func (s *store) writeRegionCheckpoint(
	region string,
	w io.Writer,
) error {
	codec, err := goavro.NewCodec(schema)
	if err != nil {
		return xerrors.Errorf("fail to create codec: %w", err)
	}

	config := goavro.OCFConfig{
		W:               w,
		Codec:           codec,
		CompressionName: goavro.CompressionSnappyLabel,
	}
	writer, err := goavro.NewOCFWriter(config)
	if err != nil {
		return xerrors.Errorf(
			"leveldb stable store: fail to create writer: %w",
			err)
	}

	block := []interface{}{}
	snap, err := s.GetSnapshot()
	if err != nil {
		return xerrors.Errorf(
			"leveldb stable store: fail to get snapshot: %w",
			err)
	}
	defer snap.Release()
	iter, err := snap.NewIterator(
		region,
		phalanx.FullScanRange(),
	)

	if err != nil {
		return xerrors.Errorf(
			"leveldb stable store: fail to get iterator of region(%s): %w",
			region, err,
		)
	}

	var resultError *multierror.Error

	for iter.Next() {
		keyRef := iter.Key()
		valueRef := iter.Value()

		key := make([]byte, len(keyRef))
		copy(key, keyRef)

		var record map[string]interface{}

		if valueRef == nil {
			record = map[string]interface{}{
				"key":   key,
				"value": goavro.Union("null", nil),
			}
		} else {
			value := make([]byte, len(valueRef))
			copy(value, valueRef)
			record = map[string]interface{}{
				"key":   key,
				"value": goavro.Union("bytes", value),
			}
		}

		block = append(block, record)

		if len(block) >= maxBatchSize {
			if err := writer.Append(block); err != nil {
				resultError = multierror.Append(resultError, err)
				break
			}
			block = []interface{}{}
		}
	}
	if len(block) > 0 {
		if err = writer.Append(block); err != nil {
			resultError = multierror.Append(resultError, err)
		}
	}
	iter.Release()
	resultError = multierror.Append(resultError, iter.Error())

	return resultError.ErrorOrNil()
}

// CreateCheckpoint creates a checkpoint of this StableStore
func (s *store) CreateCheckpoint(region string) ([]byte, error) {
	buffer := new(bytes.Buffer)
	err := s.writeRegionCheckpoint(region, buffer)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// RestoreToCheckpoint restores internal storage to checkpoint
func (s *store) RestoreToCheckpoint(
	region string,
	checkpoint []byte,
) error {

	if _, regionExists := s.storages[region]; !regionExists {
		err := s.CreateRegion(region)
		if err != nil {
			return xerrors.Errorf(
				"leveldb stable store: fail to create a(%s): %w",
				region, err,
			)
		}
	}

	// delete all data
	snap, err := s.GetSnapshot()
	if err != nil {
		return err
	}
	defer snap.Release()
	iter, err := snap.NewIterator(
		region,
		phalanx.FullScanRange(),
	)
	if err != nil {
		return xerrors.Errorf(
			"leveldb stable store: fail to get iterator of region(%s): %w",
			region, err,
		)
	}
	var resultError *multierror.Error

	batch := s.CreateBatch()
	for iter.Next() {
		key := iter.Key()
		batch.Delete(region, key)

		if batch.Len(region) >= maxBatchSize {
			err = s.Write(batch)
			resultError = multierror.Append(resultError, err)
			batch = s.CreateBatch()
		}
	}
	err = s.Write(batch)
	resultError = multierror.Append(resultError, err)
	iter.Release()
	resultError = multierror.Append(resultError, iter.Error())

	if resultError.ErrorOrNil() != nil {
		return resultError.ErrorOrNil()
	}

	reader, err := goavro.NewOCFReader(bytes.NewBuffer(checkpoint))
	batch = s.CreateBatch()
	resultError = new(multierror.Error)
	if err != nil {
		return err
	}
	for reader.Scan() {
		if record, err := reader.Read(); err == nil {
			m := record.(map[string]interface{})
			value := m["value"].(map[string]interface{})
			if el, ok := value["bytes"]; ok {
				batch.Put(region, m["key"].([]byte), el.([]byte))
			} else {
				batch.Put(region, m["key"].([]byte), nil)
			}
		} else {
			resultError = multierror.Append(resultError, err)
		}

		if batch.Len(region) >= maxBatchSize {
			err = s.Write(batch)
			resultError = multierror.Append(resultError, err)
			batch = s.CreateBatch()
		}
	}
	if batch.Len(region) > 0 {
		err = s.Write(batch)
		resultError = multierror.Append(resultError, err)
	}

	return resultError.ErrorOrNil()
}

// Close Close closes the StableStorage
func (s *store) Close() error {
	var errorResults *multierror.Error
	for key := range s.storages {
		err := s.storages[key].Close()
		errorResults = multierror.Append(errorResults, err)
	}
	return errorResults.ErrorOrNil()
}

// GetSnapshot
func (s *store) GetSnapshot() (phalanx.Snapshot, error) {
	s.RLock()
	defer s.RUnlock()

	snapshots := make(map[string]*leveldb.Snapshot)

	var err error
	for key := range s.storages {
		snapshots[key], err = s.storages[key].GetSnapshot()
		if err != nil {
			return nil, xerrors.Errorf(
				"leveldb stable store: fail to get snapshot of region(%s): %w",
				key, err,
			)
		}
	}
	return &snapshot{
		regionSnaps: snapshots,
	}, nil
}
