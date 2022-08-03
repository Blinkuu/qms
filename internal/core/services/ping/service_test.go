package ping

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewService(t *testing.T) {
	// When
	s := NewService()

	// Then
	assert.NotNil(t, s)
}

func TestService_Ping(t *testing.T) {
	// Given
	s := &Service{}

	// When
	got := s.Ping()

	// Then
	assert.Equal(t, "Pong", got)
}
