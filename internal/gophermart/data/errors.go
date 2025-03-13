package data

import "errors"

var (
	ErrUniqueConstraintViolation = errors.New("unique constraint violation")
	ErrInvalidPassword           = errors.New("invalid password")
	ErrInvalidLogin              = errors.New("invalid login")
)
