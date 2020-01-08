package plugin

// Packet plugin, for example debug,gzip,encrypt,decrypt
type PacketPlugin interface {
	Handle(header, body []byte) (h, b []byte)
}
