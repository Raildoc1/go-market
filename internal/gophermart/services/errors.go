package services

import "errors"

var (
	ErrLoginTaken         = errors.New("login is already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
