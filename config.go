package linker

import "time"

type Config struct {
	Debug       bool
	Timeout     time.Duration
	MaxPayload  uint32
	ContentType string
	Sender      []PacketPlugin
	Receiver    []PacketPlugin
}
