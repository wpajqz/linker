package linker

import (
	"context"
	"hash/crc32"
)

type Context struct {
	context.Context
	Request  *Request
	Response Response
}

func NewContext(ctx context.Context, req *Request, res Response) *Context {
	return &Context{
		ctx,
		req,
		res,
	}
}

func (c *Context) ParseParam(data interface{}) error {
	return c.Request.Params.UnPack(data)
}

func (c *Context) Write(operator string, data interface{}) (int, error) {
	p, err := c.Request.Params.Pack(crc32.ChecksumIEEE([]byte(operator)), data)
	if err != nil {
		return 0, err
	}

	return c.Response.Write(p.Bytes())
}
