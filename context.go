package linker

import (
	"context"
	"fmt"
	"hash/crc32"
	"runtime"
	"time"

	"github.com/golang/protobuf/proto"
)

type (
	Context struct {
		context.Context
		Request  *request
		Response response
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
	err := proto.Unmarshal(c.Request.Body(), data.(proto.Message))
	if err != nil {
		return fmt.Errorf("Unpack error: %v", err.Error())
	}

	return nil
}

func (c *Context) RawParam() []byte {
	return c.Request.Bytes()
}

func (c *Context) Success(body interface{}) {
	_, err := c.write(c.Request.OperateType(), body)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		panic(SystemError{time.Now(), file, line, err.Error()})
	}

	panic(nil)
}

func (c *Context) Error(body interface{}) {
	c.Response.SetResponseProperty("status", "0")
	_, err := c.write(c.Request.OperateType(), body)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		panic(SystemError{time.Now(), file, line, err.Error()})
	}

	panic(nil)
}

func (c *Context) Write(operator string, body interface{}) (int, error) {
	return c.write(crc32.ChecksumIEEE([]byte(operator)), body)
}

func (c *Context) write(operator uint32, body interface{}) (int, error) {
	pbData, err := proto.Marshal(body.(proto.Message))
	if err != nil {
		return 0, err
	}

	p := NewPack(operator, c.Response.Header(), pbData)

	return c.Response.Write(p.Bytes())
}
