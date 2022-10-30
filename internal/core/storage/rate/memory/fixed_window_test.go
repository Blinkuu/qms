package memory

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
)

func TestNewFixedWindow_ReturnsNewFixedWindowWithCorrectArguments(t *testing.T) {
	// When
	got := NewFixedWindow(clock.NewMock(), 1, 1)

	// Then
	assert.NotNil(t, got)
}

func TestNewFixedWindow_PanicsWithInvalidInterval(t *testing.T) {
	// When
	panicFunc := func() { _ = NewFixedWindow(clock.NewMock(), 0, 1) }

	// Then
	assert.Panics(t, panicFunc)
}

func TestNewFixedWindow_PanicsWithInvalidCapacity(t *testing.T) {
	// When
	panicFunc := func() { _ = NewFixedWindow(clock.NewMock(), 1, 0) }

	// Then
	assert.Panics(t, panicFunc)
}

func TestFixedWindow_Allow_ReturnsNoErrorNotOKAndZeroWithMoreTokensRequestedThanCapacity(t *testing.T) {
	// Given
	i := 2 * time.Second
	c := int64(4)
	cl := clock.NewMock()
	b := NewFixedWindow(cl, i, c)

	// When
	wait, ok, err := b.Allow(context.Background(), 5)

	// Then
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Equal(t, i, wait)
}

func TestFixedWindow_Allow_ReturnsNoErrorOKAndZeroWithLessTokensRequestedThanCapacity(t *testing.T) {
	// Given
	i := 2 * time.Second
	c := int64(4)
	cl := clock.NewMock()
	b := NewFixedWindow(cl, i, c)

	// When
	wait, ok, err := b.Allow(context.Background(), 3)

	// Then
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Zero(t, wait)
}

func TestFixedWindow_Allow_ReturnsNoErrorNotOKWhenTokensAreExhausted(t *testing.T) {
	// Given
	startTime := time.Date(2022, time.Month(1), 11, 0, 0, 1, 0, time.UTC)
	c := clock.NewMock()
	c.Set(startTime)
	b := NewFixedWindow(c, 2, 4)

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
	assert.EqualValues(t, 2, wait)
}

func TestFixedWindow_Allow_ReturnsNoErrorAndZeroWhenBucketHasTokens(t *testing.T) {
	// Given
	startTime := time.Date(2022, time.Month(1), 11, 0, 0, 1, 0, time.UTC)
	c := clock.NewMock()
	c.Set(startTime)
	b := NewFixedWindow(c, 2, 4)

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
	assert.EqualValues(t, 2, wait)

	// When
	c.Set(startTime.Add(1 * time.Second))
	wait, ok, err = b.Allow(context.Background(), 3)

	// Then
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Zero(t, wait)
}
