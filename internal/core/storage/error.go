package storage

import (
	"errors"
	"strings"
)

var (
	ErrInvalidVersion = errors.New("invalid version")
)

func IsInvalidVersionError(err string) bool {
	return strings.Contains(err, ErrInvalidVersion.Error())
}
