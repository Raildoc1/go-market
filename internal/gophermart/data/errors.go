package data

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrInvalidLogin      = errors.New("invalid login")
)
