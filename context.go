package linker

import "context"

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

func (c *Context) Write(operator int, data interface{}) (int, error) {
	return c.Response.Write(c.Request.Params.Bytes())
}
