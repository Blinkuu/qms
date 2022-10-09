package storage

import (
	"errors"
	"strings"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrInvalidVersion = errors.New("invalid version")
)

func IsErrNotFound(err string) bool {
	return strings.Contains(err, ErrNotFound.Error())
}

func IsErrInvalidVersion(err string) bool {
	return strings.Contains(err, ErrInvalidVersion.Error())
}
