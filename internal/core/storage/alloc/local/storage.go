package local

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/dgraph-io/badger/v3"

	"github.com/Blinkuu/qms/internal/core/storage/alloc/quota"
	"github.com/Blinkuu/qms/pkg/log"
	badgerlog "github.com/Blinkuu/qms/pkg/log/badger"
)

type Storage struct {
	cfg       Config
	logger    log.Logger
	db        *badger.DB
	buckets   map[string]*CappedBucket
	bucketsMu *sync.RWMutex
}

func NewStorage(cfg Config, logger log.Logger) (*Storage, error) {
	opts := badger.DefaultOptions(cfg.Dir)
	opts.Logger = badgerlog.NewLogger(logger)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger: %w", err)
	}

	return &Storage{
		cfg:       cfg,
		logger:    logger,
		db:        db,
		buckets:   make(map[string]*CappedBucket),
		bucketsMu: &sync.RWMutex{},
	}, nil
}

func (s *Storage) Alloc(ctx context.Context, namespace, resource string, tokens int64) (int64, bool, error) {
	id := strings.Join([]string{namespace, resource}, "_")

	s.bucketsMu.RLock()
	defer s.bucketsMu.RUnlock()

	bucket, found := s.buckets[id]
	if !found {
		return 0, false, fmt.Errorf("st for %s not found", id)
	}

	remainingTokens, ok, err := bucket.Alloc(ctx, tokens)
	if err != nil {
		return 0, false, fmt.Errorf("failed to alloc: %w", err)
	}

	return remainingTokens, ok, nil
}

func (s *Storage) Free(ctx context.Context, namespace, resource string, tokens int64) (int64, bool, error) {
	id := strings.Join([]string{namespace, resource}, "_")

	s.bucketsMu.RLock()
	defer s.bucketsMu.RUnlock()

	bucket, found := s.buckets[id]
	if !found {
		return 0, false, fmt.Errorf("st for %s not found", id)
	}

	remainingTokens, ok, err := bucket.Free(ctx, tokens)
	if err != nil {
		return 0, false, fmt.Errorf("failed to free: %w", err)
	}

	return remainingTokens, ok, nil
}

func (s *Storage) RegisterQuota(_ context.Context, namespace, resource string, cfg quota.Config) error {
	s.bucketsMu.Lock()
	defer s.bucketsMu.Unlock()

	id := strings.Join([]string{namespace, resource}, "_")
	_, found := s.buckets[id]
	if found {
		return errors.New("only a single strategy for a namespace-resource pair can be registered")
	}

	bucket, err := NewCappedBucket(s.db, []byte(id), cfg.Capacity)
	if err != nil {
		return fmt.Errorf("failed to create new capped bucket: %w", err)
	}

	s.buckets[id] = bucket

	return nil
}

func (s *Storage) Shutdown(_ context.Context) error {
	return s.db.Close()
}
