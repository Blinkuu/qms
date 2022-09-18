package ping

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Blinkuu/qms/pkg/log"
)

func TestNewService(t *testing.T) {
	// When
	s := NewService(log.NewNoopLogger())

	// Then
	require.NotNil(t, s)
}

func TestService_Ping(t *testing.T) {
	// Given
	s := &Service{}

	// When
	got := s.Ping(context.Background())

	// Then
	require.Equal(t, "pong", got)
}
