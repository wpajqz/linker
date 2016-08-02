package client

import (
	"net"

	"github.com/links123com/linker/library/linker"
)

type Context struct {
	conn   net.Conn
	Method int32
	Params linker.Packet
}

func (c *Context) ParseParam(pb interface{}) error {
	return c.Params.UnPack(pb)
}
