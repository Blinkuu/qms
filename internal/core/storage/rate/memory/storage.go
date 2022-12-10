package memory

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/Blinkuu/qms/internal/core/storage"
	"github.com/Blinkuu/qms/internal/core/storage/rate/quota"
	"github.com/Blinkuu/qms/pkg/timeunit"
)

type allower interface {
	Allow(ctx context.Context, tokens int64) (waitTime time.Duration, ok bool, err error)
}

type Storage struct {
	clock      clock.Clock
	strategies map[string]allower
	bucketsMu  *sync.RWMutex
}

func NewStorage(clock clock.Clock) *Storage {
	return &Storage{
		clock:      clock,
		strategies: make(map[string]allower),
		bucketsMu:  &sync.RWMutex{},
	}
}

func (s *Storage) Allow(ctx context.Context, namespace, resource string, tokens int64) (time.Duration, bool, error) {
	id := strings.Join([]string{namespace, resource}, "_")

	s.bucketsMu.RLock()
	defer s.bucketsMu.RUnlock()

	bucket, found := s.strategies[id]
	if !found {
		return 0, false, storage.ErrNotFound
	}

	waitTime, ok, err := bucket.Allow(ctx, tokens)
	if err != nil {
		return 0, false, fmt.Errorf("failed to allow: %w", err)
	}

	return waitTime, ok, nil
}

func (s *Storage) RegisterQuota(_ context.Context, namespace, resource string, cfg quota.Config) error {
	s.bucketsMu.Lock()
	defer s.bucketsMu.Unlock()

	id := strings.Join([]string{namespace, resource}, "_")
	_, found := s.strategies[id]
	if found {
		return errors.New("only a single strategy for a namespace-resource pair can be registered")
	}

	unit, err := timeunit.Parse(cfg.Unit)
	if err != nil {
		return fmt.Errorf("failed to parse time unit: %w", err)
	}

	switch cfg.Algorithm {
	case FixedWindowAlgorithm:
		s.strategies[id] = NewFixedWindow(s.clock, unit, cfg.RequestPerUnit)

		return nil
	case TokenBucketAlgorithm:
		s.strategies[id] = NewTokenBucket(s.clock, float64(cfg.RequestPerUnit)/unit.Seconds(), cfg.RequestPerUnit)

		return nil
	default:
		return fmt.Errorf("%s algorithm is not supported", cfg.Algorithm)
	}
}

func (s *Storage) Shutdown(_ context.Context) error {
	return nil
}
