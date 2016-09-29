package client

import "errors"

var (
	ErrClosed = errors.New("connection is closed.")
)
