package linker

const (
	MaxPayload = 1024 * 1024
	TIMEOUT    = 30
)

const (
	OPERATOR_HEARTBEAT = iota
	OPERATOR_MAX       = 1024
)

const (
	// Version current Linker version
	VERSION = "1.0.0"

	// MinimumGoVersion minimum required Go version for Linker
	MINIMUM_GO_VERSION = ">=1.9"
)
