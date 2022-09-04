package ratelimit

import (
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
	Allow(tokens int64) (waitTime time.Duration, err error)
}

type StrategyFactory struct {
	clock clock.Clock
}

func NewStrategyFactory(clock clock.Clock) *StrategyFactory {
	return &StrategyFactory{
		clock: clock,
	}
}

func (f *StrategyFactory) Strategy(algorithm string, unit string, requestsPerUnit int64) (Strategy, error) {
	parsedUnit, err := timeunit.Parse(unit)
	if err != nil {
		return nil, fmt.Errorf("failed to parse time unit: %w", err)
	}

	switch algorithm {
	case TokenBucketAlgorithm:
		return memory.NewTokenBucket(requestsPerUnit*int64(parsedUnit), requestsPerUnit*int64(parsedUnit), f.clock), nil
	default:
	}

	return nil, fmt.Errorf("failed to create strategy")
}
