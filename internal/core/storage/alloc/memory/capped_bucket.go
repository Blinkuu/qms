package memory

import (
	"context"
	"sync"

	"github.com/Blinkuu/qms/internal/core/storage"
)

type CappedBucket struct {
	allocated int64
	capacity  int64
	version   int64
	mu        *sync.Mutex
}

func NewCappedBucket(capacity, version int64) *CappedBucket {
	if capacity <= 0 {
		panic("capacity must be greater than 0")
	}

	return &CappedBucket{
		allocated: 0,
		capacity:  capacity,
		version:   version,
		mu:        &sync.Mutex{},
	}
}

func (c *CappedBucket) Alloc(_ context.Context, tokens, version int64) (int64, int64, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if version != 0 && c.version != version {
		return 0, 0, false, storage.ErrInvalidVersion
	}

	if c.allocated+tokens > c.capacity {
		return c.remainingTokensLocked(), c.version, false, nil
	}

	c.allocated += tokens
	c.version += 1

	return c.remainingTokensLocked(), c.version, true, nil
}

func (c *CappedBucket) Free(_ context.Context, tokens, version int64) (int64, int64, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if version != 0 && c.version != version {
		return 0, 0, false, storage.ErrInvalidVersion
	}

	if c.allocated-tokens < 0 {
		return c.remainingTokensLocked(), c.version, false, nil
	}

	c.allocated -= tokens
	c.version += 1

	return c.remainingTokensLocked(), c.version, true, nil
}

func (c *CappedBucket) remainingTokensLocked() int64 {
	return c.capacity - c.allocated
}
