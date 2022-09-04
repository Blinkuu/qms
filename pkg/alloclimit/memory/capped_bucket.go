package memory

import (
	"sync"
)

type CappedBucket struct {
	allocated int64
	capacity  int64
	mu        *sync.Mutex
}

func NewCappedBucket(capacity int64) *CappedBucket {
	if capacity <= 0 {
		panic("capacity must be greater than 0")
	}

	return &CappedBucket{
		allocated: 0,
		capacity:  capacity,
		mu:        &sync.Mutex{},
	}
}

func (c *CappedBucket) Alloc(tokens int64) (int64, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.allocated+tokens > c.capacity {
		return c.remainingTokensLocked(), false, nil
	}

	c.allocated += tokens

	return c.remainingTokensLocked(), true, nil
}

func (c *CappedBucket) Free(tokens int64) (int64, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.allocated-tokens < 0 {
		return c.remainingTokensLocked(), false, nil
	}

	c.allocated -= tokens

	return c.remainingTokensLocked(), true, nil
}

func (c *CappedBucket) remainingTokensLocked() int64 {
	return c.capacity - c.allocated
}
