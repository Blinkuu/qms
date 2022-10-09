package memory

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Blinkuu/qms/internal/core/storage"
	"github.com/Blinkuu/qms/internal/core/storage/alloc/quota"
)

type Storage struct {
	buckets   map[string]*CappedBucket
	bucketsMu *sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		buckets:   make(map[string]*CappedBucket),
		bucketsMu: &sync.RWMutex{},
	}
}

func (s *Storage) View(_ context.Context, namespace, resource string) (int64, int64, int64, error) {
	id := strings.Join([]string{namespace, resource}, "_")

	bucket, found := s.buckets[id]
	if !found {
		return 0, 0, 0, storage.ErrNotFound
	}

	allocated, capacity, version := bucket.View()

	return allocated, capacity, version, nil
}

func (s *Storage) Alloc(_ context.Context, namespace, resource string, tokens, version int64) (int64, int64, bool, error) {
	id := strings.Join([]string{namespace, resource}, "_")

	s.bucketsMu.RLock()
	defer s.bucketsMu.RUnlock()

	bucket, found := s.buckets[id]
	if !found {
		return 0, 0, false, storage.ErrNotFound
	}

	remainingTokens, currentVersion, ok, err := bucket.Alloc(tokens, version)
	if err != nil {
		return 0, 0, false, fmt.Errorf("failed to alloc: %w", err)
	}

	return remainingTokens, currentVersion, ok, nil
}

func (s *Storage) Free(_ context.Context, namespace, resource string, tokens, version int64) (int64, int64, bool, error) {
	id := strings.Join([]string{namespace, resource}, "_")

	s.bucketsMu.RLock()
	defer s.bucketsMu.RUnlock()

	bucket, found := s.buckets[id]
	if !found {
		return 0, 0, false, storage.ErrNotFound
	}

	remainingTokens, currentVersion, ok, err := bucket.Free(tokens, version)
	if err != nil {
		return 0, 0, false, fmt.Errorf("failed to free: %w", err)
	}

	return remainingTokens, currentVersion, ok, nil
}

func (s *Storage) RegisterQuota(_ context.Context, namespace, resource string, cfg quota.Config) error {
	s.bucketsMu.Lock()
	defer s.bucketsMu.Unlock()

	id := strings.Join([]string{namespace, resource}, "_")
	_, found := s.buckets[id]
	if found {
		return errors.New("only a single strategy for a namespace-resource pair can be registered")
	}

	s.buckets[id] = NewCappedBucket(cfg.Capacity, 1)

	return nil
}

func (s *Storage) Shutdown(_ context.Context) error {
	return nil
}
