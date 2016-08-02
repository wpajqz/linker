package linker

import (
	"net"
)

type Response struct {
	net.Conn
	Code    int
	Message string
}
