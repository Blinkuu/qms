package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCappedBucket_ReturnsNewTokenBucketWithCorrectArguments(t *testing.T) {
	// When
	got := NewCappedBucket(1)

	// Then
	assert.NotNil(t, got)
}

func TestNewCappedBucket_PanicsWithInvalidRefillRate(t *testing.T) {
	// When
	panicFunc := func() { _ = NewCappedBucket(0) }

	// Then
	assert.Panics(t, panicFunc)
}

func TestCappedBucket_Alloc_ReturnsNoErrorFalseAndRemainingTokensWithMoreTokensRequestedThanCapacity(t *testing.T) {
	// Given
	c := int64(4)
	b := NewCappedBucket(c)

	// When
	remaining, ok, err := b.Alloc(5)

	// Then
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.EqualValues(t, c, remaining)
}

func TestCappedBucket_Allow_ReturnsNoErrorTrueAndRemainingTokensWithLessTokensRequestedThanCapacity(t *testing.T) {
	// Given
	c := int64(4)
	b := NewCappedBucket(c)

	// When
	remaining, ok, err := b.Alloc(3)

	// Then
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.EqualValues(t, 1, remaining)
}

func TestCappedBucket_Allow_AllowsFirstAllocationAndDisallowsSecondDueToInsufficientCapacity(t *testing.T) {
	// Given
	c := int64(4)
	b := NewCappedBucket(c)

	// When
	remaining1, ok, err := b.Alloc(3)

	// Then
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.EqualValues(t, 1, remaining1)

	// When
	remaining2, ok, err := b.Alloc(2)

	// Then
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.EqualValues(t, 1, remaining2)
}

func TestCappedBucket_Free_ReturnsNoErrorFalseAndRemainingTokensWithMoreTokensFreedThanAllocated(t *testing.T) {
	// Given
	c := int64(4)
	b := NewCappedBucket(c)

	// When
	remaining, ok, err := b.Free(3)

	// Then
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.EqualValues(t, 4, remaining)
}

func TestCappedBucket_Free_ReturnsNoErrorTrueAndRemainingTokensWithLessTokensFreedThanAllocated(t *testing.T) {
	// Given
	c := int64(4)
	b := NewCappedBucket(c)
	remaining1, ok, err := b.Alloc(2)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.EqualValues(t, 2, remaining1)

	// When
	remaining2, ok, err := b.Free(1)

	// Then
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.EqualValues(t, 3, remaining2)
}
