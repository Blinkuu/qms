package memory

import (
	"context"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
)

const (
	FixedWindowAlgorithm = "fixed-window"
)

type FixedWindow struct {
	clock       clock.Clock
	windowStart time.Time
	interval    time.Duration
	allocated   int64
	capacity    int64
	mu          *sync.Mutex
}

func NewFixedWindow(clock clock.Clock, interval time.Duration, capacity int64) *FixedWindow {
	if interval <= 0 {
		panic("interval must be greater than 0")
	}

	if capacity <= 0 {
		panic("capacity must be greater than 0")
	}

	return &FixedWindow{
		clock:       clock,
		windowStart: time.Time{},
		interval:    interval,
		allocated:   0,
		capacity:    capacity,
		mu:          &sync.Mutex{},
	}
}

func (f *FixedWindow) Allow(_ context.Context, tokens int64) (waitTime time.Duration, ok bool, err error) {
	now := f.clock.Now()

	f.mu.Lock()
	defer f.mu.Unlock()

	windowEnd := f.windowStart.Add(f.interval)
	if now.After(windowEnd) {
		newWindowStart := now.Truncate(f.interval)
		f.windowStart = newWindowStart
		f.allocated = 0
	}

	if f.allocated+tokens > f.capacity {
		return f.windowStart.Add(f.interval).Sub(now), false, nil
	}

	f.allocated += tokens

	return 0, true, nil
}
