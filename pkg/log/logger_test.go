package log

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMust_ReturnsValueOnNilError(t *testing.T) {
	// Given
	val := 1
	var err error

	// When
	got := Must(val, err)

	// Then
	assert.Equal(t, val, got)
}

func TestMust_PanicsOnNonNilError(t *testing.T) {
	// Given
	val := 1
	err := errors.New("error")

	// When
	panicFunc := func() {
		_ = Must(val, err)
	}

	// Then
	assert.Panics(t, panicFunc)
}
