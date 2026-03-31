package domain

import "errors"

var (
	ErrInvalidURL  = errors.New("invalid url")
	ErrInvalidCode = errors.New("invalid code")
	ErrNotFound    = errors.New("not found")
)
