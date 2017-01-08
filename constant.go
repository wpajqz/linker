package linker

import "runtime"

const (
	MaxPayload = 2048
)

const (
	OPERATOR_HEARTBEAT = iota
	MAX_OPERATOR       = 1024
)

const (
	LINKER_VERSION = "1.0"
	LINKER_OS      = runtime.GOOS
	LINKER_ARCH    = runtime.GOARCH
)
