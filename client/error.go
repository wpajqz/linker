package client

import (
	"errors"
	"strconv"
)

var (
	ErrClosed       = errors.New("connection is closed.")
	ErrPacketLength = errors.New("the packet is big than " + strconv.Itoa(MaxPayload))
)
