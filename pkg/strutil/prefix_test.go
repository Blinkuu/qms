package strutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithPrefixOrDefault_ReturnsMergedWordsWithDotSeparatorWhenNonEmptyPrefixSpecified(t *testing.T) {
	// Given
	prefix, toPrefix := "prefix", "toPrefix"

	// When
	got := WithPrefixOrDefault(prefix, toPrefix)

	// Then
	assert.Equal(t, prefix+"."+toPrefix, got)
}

func TestWithPrefixOrDefault_ReturnsDefaultWhenEmptyPrefixSpecified(t *testing.T) {
	// Given
	prefix, toPrefix := "", "toPrefix"

	// When
	got := WithPrefixOrDefault(prefix, toPrefix)

	// Then
	assert.Equal(t, toPrefix, got)
}
