package ping

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewService(t *testing.T) {
	// When
	s := NewService()

	// Then
	require.NotNil(t, s)
}

func TestService_Ping(t *testing.T) {
	// Given
	s := &Service{}

	// When
	got := s.Ping()

	// Then
	require.Equal(t, "pong", got)
}
