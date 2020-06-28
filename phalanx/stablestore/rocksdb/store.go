package rocksdb

import (
	"bytes"
	"io"
	"regexp"
	"sync"

	"github.com/getumen/doctrine/phalanx"
	"github.com/hashicorp/go-multierror"
	"github.com/linkedin/goavro"
	"github.com/pkg/errors"
	"github.com/tecbot/gorocksdb"
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
	cfMutex  *sync.RWMutex
	cf       map[string]*gorocksdb.ColumnFamilyHandle
	storage  *gorocksdb.DB
	dataPath string
	opt      *gorocksdb.Options
}

type storeDriver struct {
}

// New creates stable store implemented by RocksDB
func (d *storeDriver) New(dataPath string) (phalanx.StableStore, error) {
	opt := gorocksdb.NewDefaultOptions()
	opt.SetCreateIfMissing(true)
	opt.SetCreateIfMissingColumnFamilies(true)
	storage, err := gorocksdb.OpenDb(opt, dataPath)
	if err != nil {
		return nil, xerrors.Errorf("fail to create rocksdb: %w", err)
	}
	return &store{
		storage:  storage,
		dataPath: dataPath,
		opt:      opt,
		cf:       make(map[string]*gorocksdb.ColumnFamilyHandle),
		cfMutex:  new(sync.RWMutex),
	}, nil
}

func init() {
	phalanx.RegisterStableStore("rocksdb", &storeDriver{})
}

// CreateRegion creates a region
func (s *store) CreateRegion(name string) error {
	s.cfMutex.Lock()
	defer s.cfMutex.Unlock()
	return s.createRegion(name)
}

func (s *store) createRegion(name string) error {

	if name == "default" {
		return errors.Errorf(
			"leveldb stable store: invalid region name (%s): default is the system region",
			name,
		)
	}

	if matched := regionNameRegExp.Match([]byte(name)); !matched {
		return errors.Errorf(
			"leveldb stable store: invalid region name (%s) allowed chars are %s",
			name, allowedRegionChars,
		)
	}

	if _, dup := s.cf[name]; dup {
		return phalanx.NewErrRegionAlreadyExists(name)
	}
	cf, err := s.storage.CreateColumnFamily(s.opt, name)
	if err != nil {
		return xerrors.Errorf(
			"leveldb stable store: fail to create region(%s): %w",
			name, err)
	}
	s.cf[name] = cf
	return nil
}

// DropRegion drop a region
func (s *store) DropRegion(name string) error {
	s.cfMutex.Lock()
	defer s.cfMutex.Unlock()
	return s.dropRegion(name)
}

func (s *store) dropRegion(name string) error {
	if name == "default" {
		return errors.Errorf(
			"leveldb stable store: invalid region name (%s): default is the system region",
			name,
		)
	}
	if cf, exist := s.cf[name]; exist {
		s.storage.DropColumnFamily(cf)
		delete(s.cf, name)
		return nil
	}
	return phalanx.NewRegionNotFound(name)
}

func (s *store) HasRegion(name string) bool {
	s.cfMutex.RLock()
	defer s.cfMutex.RUnlock()
	return s.hasRegion(name)
}

func (s *store) hasRegion(name string) bool {
	_, exists := s.cf[name]
	return exists
}

// CreateBatch creates batch
func (s *store) CreateBatch() phalanx.Batch {
	s.cfMutex.RLock()
	defer s.cfMutex.RUnlock()
	return s.createBatch()
}

func (s *store) createBatch() phalanx.Batch {
	b := gorocksdb.NewWriteBatch()
	return &batch{
		cf:      s.cf,
		cfMutex: s.cfMutex,
		batchs:  b,
	}
}

// Write apply the given batch to the StableStorage
func (s *store) Write(b phalanx.Batch) error {
	if ba, ok := b.(*batch); ok {
		defer ba.batchs.Destroy()
	}

	return s.write(b)
}

func (s *store) write(b phalanx.Batch) error {

	opt := gorocksdb.NewDefaultWriteOptions()

	if bi, ok := b.(*batch); ok {
		err := s.storage.Write(opt, bi.batchs)
		if err != nil {
			return xerrors.Errorf("fail to write: %w", err)
		}
		return nil
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
	snap, err := s.getSnapshot()
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

	defer iter.Release()

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

	s.cfMutex.Lock()
	defer s.cfMutex.Unlock()
	return s.restoreToCheckpoint(region, checkpoint)
}

func (s *store) restoreToCheckpoint(
	region string,
	checkpoint []byte,
) error {

	if _, regionExists := s.cf[region]; !regionExists {
		err := s.createRegion(region)
		if err != nil {
			return xerrors.Errorf(
				"leveldb stable store: fail to create a(%s): %w",
				region, err,
			)
		}
	}

	// delete all data
	snapIF, err := s.getSnapshot()
	if err != nil {
		return err
	}

	var snap *snapshot
	var ok bool

	if snap, ok = snapIF.(*snapshot); !ok {
		return errors.New("cast failed")
	}

	defer snap.Release()
	iter, err := snap.newIterator(
		region,
		phalanx.FullScanRange(),
	)
	if err != nil {
		return xerrors.Errorf(
			"leveldb stable store: fail to get iterator of region(%s): %w",
			region, err,
		)
	}

	defer iter.Release()

	var resultError *multierror.Error

	var ba *batch

	baIF := s.createBatch()

	if b, ok := baIF.(*batch); ok {
		ba = b
	} else {
		return errors.New("cast failed")
	}

	for iter.Next() {
		key := iter.Key()
		ba.delete(region, key)

		if ba.Len() >= maxBatchSize {
			err = s.write(ba)
			resultError = multierror.Append(resultError, err)
			baIF := s.createBatch()

			if b, ok := baIF.(*batch); !ok {
				ba = b
			} else {
				return errors.New("cast failed")
			}
		}
	}
	err = s.write(ba)
	resultError = multierror.Append(resultError, err)
	resultError = multierror.Append(resultError, iter.Error())

	if resultError.ErrorOrNil() != nil {
		return resultError.ErrorOrNil()
	}

	resultError = new(multierror.Error)

	reader, err := goavro.NewOCFReader(bytes.NewBuffer(checkpoint))
	if err != nil {
		return err
	}

	batchIF := s.createBatch()

	if b, ok := batchIF.(*batch); ok {
		ba = b
	} else {
		return errors.New("cast failed")
	}

	for reader.Scan() {
		if record, err := reader.Read(); err == nil {
			m := record.(map[string]interface{})
			value := m["value"].(map[string]interface{})
			if el, ok := value["bytes"]; ok {
				ba.put(region, m["key"].([]byte), el.([]byte))
			} else {
				ba.put(region, m["key"].([]byte), nil)
			}
		} else {
			resultError = multierror.Append(resultError, err)
		}

		if ba.Len() >= maxBatchSize {
			err = s.write(ba)
			resultError = multierror.Append(resultError, err)
			baIF := s.createBatch()

			if b, ok := baIF.(*batch); !ok {
				ba = b
			} else {
				return errors.New("cast failed")
			}
		}
	}
	if ba.Len() > 0 {
		err = s.write(ba)
		resultError = multierror.Append(resultError, err)
	}

	return resultError.ErrorOrNil()
}

// Close Close closes the StableStorage
func (s *store) Close() error {
	s.cfMutex.Lock()
	defer s.cfMutex.Unlock()

	for _, cf := range s.cf {
		cf.Destroy()
	}

	s.opt.Destroy()

	s.storage.Close()
	return nil
}

// GetSnapshot
func (s *store) GetSnapshot() (phalanx.Snapshot, error) {
	s.cfMutex.RLock()
	defer s.cfMutex.RUnlock()
	return s.getSnapshot()
}

func (s *store) getSnapshot() (phalanx.Snapshot, error) {

	snap := s.storage.NewSnapshot()

	return &snapshot{
		snap:    snap,
		db:      s.storage,
		cfMutex: s.cfMutex,
		cf:      s.cf,
	}, nil
}
