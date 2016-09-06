package linker

import "net"

type Request struct {
	net.Conn
	Method uint32
	Params Packet
}
