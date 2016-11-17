package linker

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"time"
)

type (
	Context struct {
		context.Context
		Request  *request
		Response response
	}

	SystemError struct {
		when time.Time
		what string
	}
)

func NewContext(ctx context.Context, req *request, res response) *Context {
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

func (c *Context) Success(body interface{}) {
	_, err := c.write(c.Request.Method, body)
	if err != nil {
		panic(SystemError{time.Now(), err.Error()})
	}

	panic(nil)
}

func (c *Context) Error(body interface{}) {
	_, err := c.write(c.Request.Method, body)
	if err != nil {
		panic(SystemError{time.Now(), err.Error()})
	}

	panic(nil)
}

func (c *Context) Write(operator string, body interface{}) (int, error) {
	return c.write(crc32.ChecksumIEEE([]byte(operator)), body)
}

func (c *Context) write(operator uint32, body interface{}) (int, error) {
	header, err := json.Marshal(c.Response.Header)
	if err != nil {
		return 0, err
	}

	p, err := c.Request.Params.Pack(operator, header, body)
	if err != nil {
		return 0, err
	}

	return c.Response.Write(p.Bytes())
}

func (e SystemError) Error() string {
	return fmt.Sprintf("%v: %v", e.when, e.what)
}
