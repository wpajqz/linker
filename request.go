package linker

import "net"

type Request struct {
	net.Conn
	Method int32
	Params Packet
}
