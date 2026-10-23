package domain

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrEmailTaken          = errors.New("email already taken")
	ErrInvalidInput        = errors.New("invalid input")
	ErrInvalidCode         = errors.New("invalid or expired code")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrAccountNotConfirmed = errors.New("account not confirmed")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrConflict            = errors.New("conflict")
	ErrSelfAction          = errors.New("action not allowed on your own account")
	ErrTooManyAttempts     = errors.New("too many attempts")
)
