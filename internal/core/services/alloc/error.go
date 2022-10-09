package alloc

import (
	"errors"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrInvalidVersion = errors.New("invalid version")
)
