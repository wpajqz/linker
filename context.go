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

func (c *Context) Write(operator uint32, data interface{}) (int, error) {
	p, err := c.Request.Params.Pack(operator, data)
	if err != nil {
		return 0, err
	}

	return c.Response.Write(p.Bytes())
}
