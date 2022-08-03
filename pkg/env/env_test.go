package env

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestGetOrDefault_ReturnsCorrectValueIfEnvDefined(t *testing.T) {
	// Given
	key := "TEST"
	value := "VALUE"
	require.NoError(t, os.Setenv(key, value))

	// When
	got := GetOrDefault(key, "")

	// Then
	require.Equal(t, value, got)
}

func TestGetOrDefault_ReturnsDefaultValueIfEnvNotDefined(t *testing.T) {
	// Given
	key := "TEST"
	fallback := "FALLBACK"
	require.NoError(t, os.Unsetenv(key))

	// When
	got := GetOrDefault(key, fallback)

	// Then
	require.Equal(t, fallback, got)
}

func TestGetOrDie_ReturnsCorrectValueIfEnvDefined(t *testing.T) {
	// Given
	key := "TEST"
	value := "VALUE"
	require.NoError(t, os.Setenv(key, value))

	// When
	got := GetOrDie(key)

	// Then
	require.Equal(t, value, got)
}

func TestGetOrDie_PanicsIfEnvNotDefined(t *testing.T) {
	// Given
	key := "TEST"
	require.NoError(t, os.Unsetenv(key))

	// When
	panicFunc := func() { _ = GetOrDie(key) }

	// Then
	require.Panics(t, panicFunc)
}
