package memory

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/Blinkuu/qms/internal/core/storage/rate/quota"
	"github.com/Blinkuu/qms/pkg/timeunit"
)

type Storage struct {
	clock     clock.Clock
	buckets   map[string]*TokenBucket
	bucketsMu *sync.RWMutex
}

func NewStorage(clock clock.Clock) *Storage {
	return &Storage{
		clock:     clock,
		buckets:   make(map[string]*TokenBucket),
		bucketsMu: &sync.RWMutex{},
	}
}

func (s *Storage) Allow(ctx context.Context, namespace, resource string, tokens int64) (time.Duration, bool, error) {
	id := strings.Join([]string{namespace, resource}, "_")

	s.bucketsMu.RLock()
	defer s.bucketsMu.RUnlock()

	bucket, found := s.buckets[id]
	if !found {
		return 0, false, fmt.Errorf("st for %s not found", id)
	}

	waitTime, ok, err := bucket.Allow(ctx, tokens)
	if err != nil {
		return 0, false, fmt.Errorf("failed to allow: %w", err)
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

	parsedUnit, err := timeunit.Parse(cfg.Unit)
	if err != nil {
		return fmt.Errorf("failed to parse time unit: %w", err)
	}

	switch cfg.Algorithm {
	case TokenBucketAlgorithm:
		s.buckets[id] = NewTokenBucket(s.clock, cfg.RequestPerUnit/int64(parsedUnit), cfg.RequestPerUnit*int64(parsedUnit))

		return nil
	default:
		return fmt.Errorf("%s algorithm is not supported", cfg.Algorithm)
	}
}
