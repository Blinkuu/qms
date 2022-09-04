package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	// When
	s := NewPingService()

	// Then
	require.NotNil(t, s)
}

func TestService_Ping(t *testing.T) {
	// Given
	s := &PingService{}

	// When
	got := s.Ping(context.Background())

	// Then
	require.Equal(t, "pong", got)
}
