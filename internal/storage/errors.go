package storage

import "errors"

var (
	ErrURLNotFound       = errors.New("storage: url not found")
	ErrURLAlreadyExists  = errors.New("storage: url already exists")
	ErrUserAlreadyExists = errors.New("storage: user already exists")
)
