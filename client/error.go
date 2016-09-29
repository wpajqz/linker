package client

import "errors"

var (
	ErrClosed       = errors.New("connection is closed.")
	ErrPacketLength = errors.New("the packet is big than " + MaxPayload)
)
