package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/Blinkuu/qms/pkg/math"
)

type TokenBucket struct {
	nanosBetweenTokens       int64
	tokensNextAvailableNanos int64
	accumulatedTokens        int64
	capacity                 int64
	clock                    clock.Clock
	mu                       *sync.Mutex
}

func NewTokenBucket(refillRate int64, capacity int64, clock clock.Clock) *TokenBucket {
	if refillRate <= 0 {
		panic("refill rate must be greater than 0")
	}

	if capacity <= 0 {
		panic("capacity must be greater than 0")
	}

	return &TokenBucket{
		nanosBetweenTokens:       1e9 / refillRate,
		tokensNextAvailableNanos: 0,
		accumulatedTokens:        capacity,
		capacity:                 capacity,
		clock:                    clock,
		mu:                       &sync.Mutex{},
	}
}

// Allow returns 0 if a request is allowed and a non-zero time in
// nanoseconds if a request is not allowed.
func (b *TokenBucket) Allow(_ context.Context, tokens int64) (time.Duration, error) {
	if tokens > b.capacity {
		return 0, errors.New("requested more tokens than available capacity")
	}

	currentTimeNanos := b.clock.Now().UnixNano()

	b.mu.Lock()
	defer b.mu.Unlock()

	waitTimeNanos := b.refillLocked(tokens, currentTimeNanos)

	return time.Duration(waitTimeNanos) * time.Nanosecond, nil
}

func (b *TokenBucket) refillLocked(requestedTokens int64, currentTimeNanos int64) int64 {
	tokensNextAvailableNanos := b.tokensNextAvailableNanos
	accumulatedTokens := b.accumulatedTokens

	var freshTokens int64

	if currentTimeNanos > tokensNextAvailableNanos {
		freshTokens = (currentTimeNanos - tokensNextAvailableNanos) / b.nanosBetweenTokens
		accumulatedTokens = math.Min(b.capacity, accumulatedTokens+freshTokens)
		tokensNextAvailableNanos = currentTimeNanos
	}

	waitTimeNanos := tokensNextAvailableNanos - currentTimeNanos
	accumulatedTokensUsed := math.Min(accumulatedTokens, requestedTokens)
	tokensToWaitFor := requestedTokens - accumulatedTokensUsed
	futureWaitNanos := tokensToWaitFor * b.nanosBetweenTokens

	tokensNextAvailableNanos += futureWaitNanos
	accumulatedTokens -= accumulatedTokensUsed

	b.tokensNextAvailableNanos = tokensNextAvailableNanos
	b.accumulatedTokens = accumulatedTokens

	return waitTimeNanos
}
