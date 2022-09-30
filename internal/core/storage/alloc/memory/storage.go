package memory

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

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

	s.buckets[id] = NewCappedBucket(cfg.Capacity)

	return nil
}
