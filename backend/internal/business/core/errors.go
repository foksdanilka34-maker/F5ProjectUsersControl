package core

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrAccessDenied = errors.New("access denied")
	ErrInvalidInput = errors.New("invalid input")
)
