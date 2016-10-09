package linker

import (
	"net"
)

type Response struct {
	net.Conn
	Method uint32
	Params Packet
}
