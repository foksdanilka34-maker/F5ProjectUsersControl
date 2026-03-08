package core

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrAccessDenied      = errors.New("access denied")
	ErrInvalidInput      = errors.New("invalid input")
	ErrProjectNameExists = errors.New("project with this name already exists")
)


