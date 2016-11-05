package linker

import (
	"context"
	"fmt"
	"hash/crc32"
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
		request  *Request
		response Response
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
	return c.request.Params.UnPack(data)
}

func (c *Context) RawParam() []byte {
	return c.request.Params.Bytes()
}

func (ctx *Context) Success(data interface{}) {
	_, err := ctx.write(ctx.request.Method, data)
	if err != nil {
		panic(SystemError{time.Now(), err.Error()})
	}

	panic(SystemError{time.Now(), "user stop run"})
}

func (ctx *Context) Error(data interface{}) {
	_, err := ctx.write(uint32(0), data)
	if err != nil {
		panic(SystemError{time.Now(), err.Error()})
	}

	panic(SystemError{time.Now(), "user stop run"})
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

func (e SystemError) Error() string {
	return fmt.Sprintf("%v: %v", e.when, e.what)
}
