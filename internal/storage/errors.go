package storage

import "errors"

var (
	ErrInvalidValue = errors.New("invalid value")
	ErrNotFound     = errors.New("not found")
)
