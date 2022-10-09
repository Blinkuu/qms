package memory

import (
	"context"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/juju/ratelimit"
)

const (
	TokenBucketAlgorithm = "token-bucket"
)

type TokenBucket struct {
	tb *ratelimit.Bucket
}

func NewTokenBucket(clock clock.Clock, refillRate int64, capacity int64) *TokenBucket {
	if refillRate <= 0 {
		panic("refill rate must be greater than 0")
	}

	if capacity <= 0 {
		panic("capacity must be greater than 0")
	}

	return &TokenBucket{
		tb: ratelimit.NewBucketWithRateAndClock(float64(refillRate), capacity, clock),
	}
}

// Allow true and wait time if a request is allowed. Returns false if request is not allowed.
func (b *TokenBucket) Allow(_ context.Context, tokens int64) (time.Duration, bool, error) {
	waitTime, ok := b.tb.TakeMaxDuration(tokens, 0)
	return waitTime, ok, nil
}
