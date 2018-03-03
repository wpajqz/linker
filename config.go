package linker

import "time"

type Config struct {
	Debug                   bool
	ReadBufferSize          int
	WriteBufferSize         int
	Timeout                 time.Duration
	MaxPayload              uint32
	ContentType             string
	PluginForPacketSender   []PacketPlugin
	PluginForPacketReceiver []PacketPlugin
}
