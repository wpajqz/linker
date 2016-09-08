package client

import (
	"net"

	"github.com/wpajqz/linker"
)

type Context struct {
	conn   net.Conn
	Method uint32
	Params linker.Packet
}

func (c *Context) ParseParam(pb interface{}) error {
	return c.Params.UnPack(pb)
}
