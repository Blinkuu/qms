package memory

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
)

func TestNewTokenBucket_ReturnsNewTokenBucketWithCorrectArguments(t *testing.T) {
	// When
	got := NewTokenBucket(clock.NewMock(), 1, 1)

	// Then
	assert.NotNil(t, got)
}

func TestNewTokenBucket_PanicsWithInvalidRefillRate(t *testing.T) {
	// When
	panicFunc := func() { _ = NewTokenBucket(clock.NewMock(), 0, 1) }

	// Then
	assert.Panics(t, panicFunc)
}

func TestNewTokenBucket_PanicsWithInvalidCapacity(t *testing.T) {
	// When
	panicFunc := func() { _ = NewTokenBucket(clock.NewMock(), 1, 0) }

	// Then
	assert.Panics(t, panicFunc)
}

func TestTokenBucket_Allow_ReturnsNoErrorNotOKAndZeroWithMoreTokensRequestedThanCapacity(t *testing.T) {
	// Given
	rr := int64(2)
	c := int64(4)
	cl := clock.NewMock()
	b := NewTokenBucket(cl, rr, c)

	// When
	wait, ok, err := b.Allow(context.Background(), 5)

	// Then
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Zero(t, wait)
}

func TestTokenBucket_Allow_ReturnsNoErrorOKAndZeroWithLessTokensRequestedThanCapacity(t *testing.T) {
	// Given
	rr := int64(2)
	c := int64(4)
	cl := clock.NewMock()
	b := NewTokenBucket(cl, rr, c)

	// When
	wait, ok, err := b.Allow(context.Background(), 3)

	// Then
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Zero(t, wait)
}

func TestTokenBucket_Allow_ReturnsNoErrorNotOKWhenTokensAreExhausted(t *testing.T) {
	// Given
	startTime := time.Date(2022, time.Month(1), 11, 0, 0, 1, 0, time.UTC)
	c := clock.NewMock()
	c.Set(startTime)
	b := NewTokenBucket(c, 2, 4)

	// When
	wait, ok, err := b.Allow(context.Background(), 3)

	// Then
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Zero(t, wait)

	// When
	wait, ok, err = b.Allow(context.Background(), 3)

	// Then
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Zero(t, wait)
}

func TestTokenBucket_Allow_ReturnsNoErrorAndZeroWhenBucketHasTokens(t *testing.T) {
	// Given
	startTime := time.Date(2022, time.Month(1), 11, 0, 0, 1, 0, time.UTC)
	c := clock.NewMock()
	c.Set(startTime)
	b := NewTokenBucket(c, 2, 4)

	// When
	wait, ok, err := b.Allow(context.Background(), 3)

	// Then
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Zero(t, wait)

	// When
	wait, ok, err = b.Allow(context.Background(), 3)

	// Then
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Zero(t, wait)

	// When
	c.Set(startTime.Add(1 * time.Second))
	wait, ok, err = b.Allow(context.Background(), 3)

	// Then
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Zero(t, wait)
}
