package leveldb

import (
	"bytes"

	"github.com/getumen/doctrine/phalanx"
	"github.com/hashicorp/go-multierror"
	"github.com/linkedin/goavro"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const schema = `
{
	"namespace": "phalanx.avro",
 	"type": "record",
 	"name": "Checkpoint",
 	"fields": [
    	{"name": "key", "type": "bytes"},
    	{"name": "value",  "type": ["bytes", "null"]},
	]
}
`

const maxBatchSize = 256

type store struct {
	internal *leveldb.DB
}

// CreateBatch creates batch
func (s *store) CreateBatch() phalanx.Batch {
	return &batch{
		internal: new(leveldb.Batch),
	}
}

// Write apply the given batch to the StableStorage
func (s *store) Write(b phalanx.Batch, wo *phalanx.WriteOptions) error {
	if bi, ok := b.(*batch); ok {
		return s.internal.Write(bi.internal, &opt.WriteOptions{
			Sync: wo.Sync,
		})
	}
	return errors.Errorf("cast fail")
}

// CreateCheckpoint creates a checkpoint of this StableStore
func (s *store) CreateCheckpoint() ([]byte, error) {
	codec, err := goavro.NewCodec(schema)
	if err != nil {
		return nil, err
	}
	buffer := new(bytes.Buffer)
	config := goavro.OCFConfig{
		W:               buffer,
		Codec:           codec,
		CompressionName: goavro.CompressionSnappyLabel,
	}
	writer, err := goavro.NewOCFWriter(config)
	if err != nil {
		return nil, err
	}

	block := []interface{}{}
	cp, err := s.GetSnapshot()
	if err != nil {
		return nil, err
	}
	defer cp.Release()
	it := cp.NewIterator(
		&phalanx.Range{Start: nil, End: nil},
		&phalanx.ReadOptions{FillCache: false},
	)

	var resultError *multierror.Error

	for it.Next() {

		record := map[string][]byte{
			"key":   it.Key(),
			"value": it.Value(),
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
	if err = writer.Append(block); err != nil {
		resultError = multierror.Append(resultError, err)
	}
	it.Release()
	resultError = multierror.Append(resultError, it.Error())

	return buffer.Bytes(), resultError.ErrorOrNil()
}

// RestoreToCheckpoint restores internal storage to checkpoint
func (s *store) RestoreToCheckpoint(checkpoint []byte) error {
	// delete all data
	snap, err := s.GetSnapshot()
	if err != nil {
		return err
	}
	defer snap.Release()
	it := snap.NewIterator(
		&phalanx.Range{Start: nil, End: nil},
		&phalanx.ReadOptions{FillCache: false},
	)
	var resultError *multierror.Error

	batch := s.CreateBatch()
	for it.Next() {
		batch.Delete(it.Key())

		if batch.Len() >= maxBatchSize {
			err = s.Write(batch, nil)
			resultError = multierror.Append(resultError, err)
			batch = s.CreateBatch()
		}
	}
	err = s.Write(batch, nil)
	resultError = multierror.Append(resultError, err)
	it.Release()
	resultError = multierror.Append(resultError, it.Error())

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
			m := record.(map[string][]byte)
			batch.Put(m["key"], m["value"])
		} else {
			resultError = multierror.Append(resultError, err)
		}

		if batch.Len() >= maxBatchSize {
			err = s.Write(batch, nil)
			resultError = multierror.Append(resultError, err)
			batch = s.CreateBatch()
		}
	}
	err = s.Write(batch, nil)
	resultError = multierror.Append(resultError, err)

	return resultError.ErrorOrNil()
}

// Close Close closes the StableStorage
func (s *store) Close() error {
	return s.internal.Close()
}

// GetSnapshot
func (s *store) GetSnapshot() (phalanx.Snapshot, error) {
	snap, err := s.internal.GetSnapshot()
	if err != nil {
		return nil, err
	}
	return &snapshot{
		internal: snap,
	}, nil
}

// OpenTransaction returns Transaction
func (s *store) OpenTransaction() (phalanx.Transaction, error) {
	tx, err := s.internal.OpenTransaction()
	if err != nil {
		return nil, err
	}
	return &transaction{
		internal: tx,
	}, nil
}
