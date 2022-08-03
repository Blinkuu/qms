package helloworld

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_HelloWorld(t *testing.T) {
	got := HelloWorld()

	require.Equal(t, "Hello, World!", got)
}
