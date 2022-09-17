package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/Blinkuu/qms/pkg/ratelimit/memory"
	"github.com/Blinkuu/qms/pkg/timeunit"
)

const (
	TokenBucketAlgorithm string = "token-bucket"
)

type Strategy interface {
	Allow(ctx context.Context, tokens int64) (waitTime time.Duration, ok bool, err error)
}

type StrategyFactory interface {
	Strategy(algorithm string, unit string, requestsPerUnit int64) (Strategy, error)
}

type MemoryStrategyFactory struct {
	clock clock.Clock
}

func NewMemoryStrategyFactory(clock clock.Clock) *MemoryStrategyFactory {
	return &MemoryStrategyFactory{
		clock: clock,
	}
}

func (f *MemoryStrategyFactory) Strategy(algorithm string, unit string, requestsPerUnit int64) (Strategy, error) {
	parsedUnit, err := timeunit.Parse(unit)
	if err != nil {
		return nil, fmt.Errorf("failed to parse time unit: %w", err)
	}

	switch algorithm {
	case TokenBucketAlgorithm:
		return memory.NewTokenBucket(f.clock, requestsPerUnit/int64(parsedUnit), requestsPerUnit*int64(parsedUnit)), nil
	default:
		return nil, fmt.Errorf("%s algorithm is not supported", algorithm)
	}
}
