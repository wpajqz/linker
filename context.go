package linker

import (
	"context"
	"hash/crc32"
)

type Context struct {
	context.Context
	request  *Request
	response Response
}

func NewContext(ctx context.Context, req *Request, res Response) *Context {
	return &Context{
		ctx,
		req,
		res,
	}
}

func (c *Context) ParseParam(data interface{}) error {
	return c.request.Params.UnPack(data)
}

func (ctx *Context) Success(data interface{}) {
	ctx.write(ctx.request.Method, data)
}

func (ctx *Context) Error(data interface{}) {
	ctx.write(uint32(0), data)
}

func (c *Context) Write(operator string, data interface{}) (int, error) {
	return c.write(crc32.ChecksumIEEE([]byte(operator)), data)
}

func (c *Context) write(operator uint32, data interface{}) (int, error) {
	p, err := c.request.Params.Pack(operator, data)
	if err != nil {
		return 0, err
	}

	return c.response.Write(p.Bytes())
}
