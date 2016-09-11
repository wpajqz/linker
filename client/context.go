package client

import (
	"github.com/wpajqz/linker"
)

type Context struct {
	Method uint32
	Params linker.Packet
}

func (c *Context) ParseParam(pb interface{}) error {
	return c.Params.UnPack(pb)
}
