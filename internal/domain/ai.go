package domain

import "errors"

var (
	ErrInvalidMimeType = errors.New("invalid mime type")
	ErrNoStream        = errors.New("no event stream")
)
