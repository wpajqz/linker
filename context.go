package linker

import (
	"context"
	"fmt"
	"net"
	"time"
)

type (
	Request struct {
		net.Conn
		Method uint32
		Params Packet
	}

	Response struct {
		net.Conn
		Method uint32
		Params Packet
	}

	Context struct {
		context.Context
		Request  *Request
		Response Response
	}

	SystemError struct {
		when time.Time
		what string
	}
)

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

func (c *Context) RawParam() []byte {
	return c.Request.Params.Bytes()
}

func (ctx *Context) Success(data interface{}) {
	_, err := ctx.Write(ctx.Request.Method, data)
	if err != nil {
		panic(SystemError{time.Now(), err.Error()})
	}

	panic(nil)
}

func (ctx *Context) Error(data interface{}) {
	_, err := ctx.Write(uint32(0), data)
	if err != nil {
		panic(SystemError{time.Now(), err.Error()})
	}

	panic(nil)
}

func (c *Context) Write(operator uint32, data interface{}) (int, error) {
	p, err := c.Request.Params.Pack(operator, data)
	if err != nil {
		return 0, err
	}

	return c.Response.Write(p.Bytes())
}

func (e SystemError) Error() string {
	return fmt.Sprintf("%v: %v", e.when, e.what)
}
