package local

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v3"

	"github.com/Blinkuu/qms/internal/core/storage"
	"github.com/Blinkuu/qms/internal/core/storage/alloc/quota"
	"github.com/Blinkuu/qms/pkg/log"
	badgerlog "github.com/Blinkuu/qms/pkg/log/badger"
)

type item struct {
	Allocated int64
	Capacity  int64
	Version   int64
}

type Storage struct {
	cfg Config
	db  *badger.DB
}

func NewStorage(cfg Config, logger log.Logger) (*Storage, error) {
	opts := badger.DefaultOptions(cfg.Dir)
	opts.Logger = badgerlog.NewLogger(logger)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger: %w", err)
	}

	return &Storage{
		cfg: cfg,
		db:  db,
	}, nil
}

func (s *Storage) View(_ context.Context, namespace, resource string) (int64, int64, int64, error) {
	if s.db.IsClosed() {
		return 0, 0, 0, errors.New("badger db is closed")
	}

	id := strings.Join([]string{namespace, resource}, "_")

	txn := s.db.NewTransaction(true)
	defer txn.Discard()

	it, err := get[item](txn, id)
	if err != nil {
		switch {
		case errors.Is(err, badger.ErrKeyNotFound):
			return 0, 0, 0, storage.ErrNotFound
		default:
		}

		return 0, 0, 0, fmt.Errorf("failed to get: %w", err)
	}

	return it.Allocated, it.Capacity, it.Version, nil
}

func (s *Storage) Alloc(_ context.Context, namespace, resource string, tokens, version int64) (int64, int64, bool, error) {
	if s.db.IsClosed() {
		return 0, 0, false, errors.New("badger db is closed")
	}

	id := strings.Join([]string{namespace, resource}, "_")

	txn := s.db.NewTransaction(true)
	defer txn.Discard()

	it, err := get[item](txn, id)
	if err != nil {
		switch {
		case errors.Is(err, badger.ErrKeyNotFound):
			return 0, 0, false, storage.ErrNotFound
		default:
		}

		return 0, 0, false, fmt.Errorf("failed to get: %w", err)
	}

	if version != 0 && it.Version != version {
		return 0, 0, false, storage.ErrInvalidVersion
	}

	newAllocated := it.Allocated + tokens
	if newAllocated > it.Capacity {
		return it.Capacity - it.Allocated, it.Version, false, nil
	}

	it.Allocated = newAllocated
	it.Version += 1
	if err := set[item](txn, id, it); err != nil {
		return 0, 0, false, fmt.Errorf("failed to set item: %w", err)
	}

	if err := txn.Commit(); err != nil {
		return 0, 0, false, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return it.Capacity - it.Allocated, it.Version, true, nil
}

func (s *Storage) Free(_ context.Context, namespace, resource string, tokens, version int64) (int64, int64, bool, error) {
	if s.db.IsClosed() {
		return 0, 0, false, errors.New("badger db is closed")
	}

	id := strings.Join([]string{namespace, resource}, "_")

	txn := s.db.NewTransaction(true)
	defer txn.Discard()

	it, err := get[item](txn, id)
	if err != nil {
		switch {
		case errors.Is(err, badger.ErrKeyNotFound):
			return 0, 0, false, storage.ErrNotFound
		default:
		}

		return 0, 0, false, fmt.Errorf("failed to get: %w", err)
	}

	if version != 0 && it.Version != version {
		return 0, 0, false, storage.ErrInvalidVersion
	}

	newAllocated := it.Allocated - tokens
	if newAllocated < 0 {
		return it.Capacity - it.Allocated, it.Version, false, nil
	}

	it.Allocated = newAllocated
	it.Version += 1
	if err := set[item](txn, id, it); err != nil {
		return 0, 0, false, fmt.Errorf("failed to set item: %w", err)
	}

	if err := txn.Commit(); err != nil {
		return 0, 0, false, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return it.Capacity - it.Allocated, it.Version, true, nil
}

func (s *Storage) RegisterQuota(_ context.Context, namespace, resource string, cfg quota.Config) error {
	if s.db.IsClosed() {
		return errors.New("badger db is closed")
	}

	id := strings.Join([]string{namespace, resource}, "_")

	txn := s.db.NewTransaction(true)
	defer txn.Discard()

	_, err := get[item](txn, id)
	if !errors.Is(err, badger.ErrKeyNotFound) {
		return nil
	}

	if err := set[item](txn, id, item{Allocated: 0, Capacity: cfg.Capacity, Version: 1}); err != nil {
		return fmt.Errorf("failed to set item :%w", err)
	}

	if err := txn.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Storage) Shutdown(_ context.Context) error {
	return s.db.Close()
}

func get[T any](txn *badger.Txn, key string) (T, error) {
	item, err := txn.Get([]byte(key))
	if err != nil {
		var zero T
		return zero, fmt.Errorf("failed to get key: %w", err)
	}

	var result T
	err = item.Value(func(val []byte) error {
		err := binary.Read(bytes.NewReader(val), binary.BigEndian, &result)
		if err != nil {
			return fmt.Errorf("failed to read bytes: %w", err)
		}

		return nil
	})
	if err != nil {
		var zero T
		return zero, fmt.Errorf("failed to read item value: %w", err)
	}

	return result, nil
}

func set[T any](txn *badger.Txn, key string, value T) error {
	buf := bytes.NewBuffer(nil)
	if err := binary.Write(buf, binary.BigEndian, value); err != nil {
		return fmt.Errorf("failed to write to buffer: %w", err)
	}

	if err := txn.Set([]byte(key), buf.Bytes()); err != nil {
		return fmt.Errorf("failed to set value: %w", err)
	}

	return nil
}
