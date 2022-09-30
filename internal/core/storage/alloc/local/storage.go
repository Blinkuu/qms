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
	opts.Logger = badgerLogger{logger: logger}
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

func (s *Storage) Alloc(ctx context.Context, namespace, resource string, tokens int64) (remainingTokens int64, ok bool, err error) {
	id := strings.Join([]string{namespace, resource}, "_")

	s.bucketsMu.RLock()
	defer s.bucketsMu.RUnlock()

	bucket, found := s.buckets[id]
	if !found {
		return 0, false, fmt.Errorf("st for %s not found", id)
	}

	waitTime, ok, err := bucket.Alloc(ctx, tokens)
	if err != nil {
		return 0, false, fmt.Errorf("failed to alloc: %w", err)
	}

	return waitTime, ok, nil
}

func (s *Storage) Free(ctx context.Context, namespace, resource string, tokens int64) (remainingTokens int64, ok bool, err error) {
	id := strings.Join([]string{namespace, resource}, "_")

	s.bucketsMu.RLock()
	defer s.bucketsMu.RUnlock()

	bucket, found := s.buckets[id]
	if !found {
		return 0, false, fmt.Errorf("st for %s not found", id)
	}

	waitTime, ok, err := bucket.Free(ctx, tokens)
	if err != nil {
		return 0, false, fmt.Errorf("failed to free: %w", err)
	}

	return waitTime, ok, nil
}

func (s *Storage) RegisterQuota(namespace, resource string, cfg quota.Config) error {
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

type badgerLogger struct {
	logger log.Logger
}

func (b badgerLogger) Debugf(template string, args ...any) {
	b.logger.Debugf(template, args)
}

func (b badgerLogger) Infof(template string, args ...any) {
	b.logger.Infof(template, args)
}

func (b badgerLogger) Warningf(template string, args ...any) {
	b.logger.Warnf(template, args)
}

func (b badgerLogger) Errorf(template string, args ...any) {
	b.logger.Errorf(template, args)
}
