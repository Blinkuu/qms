package local

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/dgraph-io/badger/v3"
)

type CappedBucket struct {
	db       *badger.DB
	key      []byte
	capacity int64
}

func NewCappedBucket(db *badger.DB, key []byte, capacity int64) (*CappedBucket, error) {
	if capacity <= 0 {
		panic("capacity must be greater than 0")
	}

	if db.IsClosed() {
		return nil, errors.New("badger db is closed")
	}

	var value []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		tmpValue, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		// Copy the value as the value provided by Badger is only valid while the
		// transaction is open.
		value = make([]byte, len(tmpValue))
		copy(value, tmpValue)

		return nil
	})
	switch {
	case errors.Is(err, badger.ErrKeyNotFound):
		buf := bytes.NewBuffer(nil)
		err := binary.Write(buf, binary.BigEndian, int64(0))
		if err != nil {
			return nil, fmt.Errorf("failed to write to buffer: %w", err)
		}

		err = db.Update(func(txn *badger.Txn) error {
			return txn.Set(key, buf.Bytes())
		})
		if err != nil {
			return nil, fmt.Errorf("failed to set key: %w", err)
		}
	case err != nil:
		return nil, fmt.Errorf("failed to get key: %w", err)
	default:
	}

	return &CappedBucket{
		db:       db,
		key:      key,
		capacity: capacity,
	}, nil
}

func (c *CappedBucket) Alloc(_ context.Context, tokens int64) (int64, bool, error) {
	if c.db.IsClosed() {
		return 0, false, errors.New("badger db is closed")
	}

	txn := c.db.NewTransaction(true)
	defer txn.Discard()

	item, err := txn.Get(c.key)
	if err != nil {
		return 0, false, fmt.Errorf("failed to get key: %w", err)
	}

	var allocated int64
	err = item.Value(func(val []byte) error {
		err := binary.Read(bytes.NewReader(val), binary.BigEndian, &allocated)
		if err != nil {
			return fmt.Errorf("failed to read bytes: %w", err)
		}

		return nil
	})
	if err != nil {
		return 0, false, fmt.Errorf("failed to read item value: %w", err)
	}

	newAllocated := allocated + tokens
	if newAllocated > c.capacity {
		return c.capacity - allocated, false, nil
	}

	buf := bytes.NewBuffer(nil)
	if err := binary.Write(buf, binary.BigEndian, newAllocated); err != nil {
		return 0, false, fmt.Errorf("failed to write to buffer: %w", err)
	}

	if err := txn.Set(c.key, buf.Bytes()); err != nil {
		return 0, false, fmt.Errorf("failed to set value: %w", err)
	}

	if err := txn.Commit(); err != nil {
		return 0, false, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return c.capacity - newAllocated, true, nil
}

func (c *CappedBucket) Free(_ context.Context, tokens int64) (int64, bool, error) {
	if c.db.IsClosed() {
		return 0, false, errors.New("badger db is closed")
	}

	txn := c.db.NewTransaction(true)
	defer txn.Discard()

	item, err := txn.Get(c.key)
	if err != nil {
		return 0, false, fmt.Errorf("failed to get key: %w", err)
	}

	var allocated int64
	err = item.Value(func(val []byte) error {
		err := binary.Read(bytes.NewReader(val), binary.BigEndian, &allocated)
		if err != nil {
			return fmt.Errorf("failed to read bytes: %w", err)
		}

		return nil
	})
	if err != nil {
		return 0, false, fmt.Errorf("failed to read item value: %w", err)
	}

	newAllocated := allocated - tokens
	if newAllocated < 0 {
		return c.capacity - allocated, false, nil
	}

	buf := bytes.NewBuffer(nil)
	if err := binary.Write(buf, binary.BigEndian, newAllocated); err != nil {
		return 0, false, fmt.Errorf("failed to write to buffer: %w", err)
	}

	if err := txn.Set(c.key, buf.Bytes()); err != nil {
		return 0, false, fmt.Errorf("failed to set value: %w", err)
	}

	if err := txn.Commit(); err != nil {
		return 0, false, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return c.capacity - newAllocated, true, nil
}
