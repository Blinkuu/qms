package memory

import (
	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewTokenBucket_ReturnsNewTokenBucketWithCorrectArguments(t *testing.T) {
	// When
	got := NewTokenBucket(1, 1, clock.NewMock())

	// Then
	assert.NotNil(t, got)
}

func TestNewTokenBucket_PanicsWithInvalidRefillRate(t *testing.T) {
	// When
	panicFunc := func() { _ = NewTokenBucket(0, 1, clock.NewMock()) }

	// Then
	assert.Panics(t, panicFunc)
}

func TestNewTokenBucket_PanicsWithInvalidCapacity(t *testing.T) {
	// When
	panicFunc := func() { _ = NewTokenBucket(1, 0, clock.NewMock()) }

	// Then
	assert.Panics(t, panicFunc)
}

func TestTokenBucket_Allow_ReturnsErrorAndZeroWithMoreTokensRequestedThanCapacity(t *testing.T) {
	// Given
	rr := int64(2)
	c := int64(4)
	cl := clock.NewMock()
	b := NewTokenBucket(rr, c, cl)

	// When
	wait, err := b.Allow(5)

	// Then
	assert.Error(t, err)
	assert.Zero(t, wait)
}

func TestTokenBucket_Allow_ReturnsNoErrorAndZeroWithLessTokensRequestedThanCapacity(t *testing.T) {
	// Given
	rr := int64(2)
	c := int64(4)
	cl := clock.NewMock()
	b := NewTokenBucket(rr, c, cl)

	// When
	wait, err := b.Allow(3)

	// Then
	assert.NoError(t, err)
	assert.Zero(t, wait)
}

func TestTokenBucket_Allow_ReturnsNoErrorAndCorrectWaitTimeWhenTokensAreExhausted(t *testing.T) {
	// Given
	startTime := time.Date(2022, time.Month(1), 11, 0, 0, 1, 0, time.UTC)
	c := clock.NewMock()
	c.Set(startTime)
	b := NewTokenBucket(2, 4, c)

	// When
	wait, err := b.Allow(3)

	// Then
	assert.NoError(t, err)
	assert.Zero(t, wait)

	// When
	wait, err = b.Allow(3)

	// Then
	assert.NoError(t, err)
	assert.Zero(t, wait)

	// When
	wait, err = b.Allow(3)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, time.Second, wait)
}

func TestTokenBucket_Allow_ReturnsNoErrorAndZeroWhenBucketHasTokens(t *testing.T) {
	// Given
	startTime := time.Date(2022, time.Month(1), 11, 0, 0, 1, 0, time.UTC)
	c := clock.NewMock()
	c.Set(startTime)
	b := NewTokenBucket(2, 4, c)

	// When
	wait, err := b.Allow(3)

	// Then
	assert.NoError(t, err)
	assert.Zero(t, wait)

	// When
	wait, err = b.Allow(3)

	// Then
	assert.NoError(t, err)
	assert.Zero(t, wait)

	// When
	c.Set(startTime.Add(1 * time.Second))
	wait, err = b.Allow(3)

	// Then
	assert.NoError(t, err)
	assert.Zero(t, wait)
}
