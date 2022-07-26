package helloworld

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func Test_HelloWorld(t *testing.T) {
    got := HelloWorld()

    assert.Equal(t, "Hello, World!", got)
}
