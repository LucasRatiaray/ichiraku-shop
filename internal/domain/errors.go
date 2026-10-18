package domain

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrEmailTaken   = errors.New("email already taken")
	ErrInvalidInput = errors.New("invalid input")
)
