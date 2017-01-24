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

func (c *Context) ParseParam(data proto.Message) error {
	err := proto.Unmarshal(c.Request.Body, data)
	if err != nil {
		return fmt.Errorf("Unpack error: %v", err.Error())
	}

	return nil
}

// 响应请求成功的数据包
func (c *Context) Success(body proto.Message) {
	_, err := c.write(body)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		panic(SystemError{time.Now(), file, line, err.Error()})
	}

	panic(nil)
}

// 响应请求失败的数据包
func (c *Context) Error(body proto.Message) {
	c.Response.SetResponseProperty("status", "0")
	_, err := c.write(body)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		panic(SystemError{time.Now(), file, line, err.Error()})
	}

	panic(nil)
}

// 向客户端发送数据
func (c *Context) Write(operator string, body proto.Message) (int, error) {
	pbData, err := proto.Marshal(body)
	if err != nil {
		return 0, err
	}

	p := NewPack(crc32.ChecksumIEEE([]byte(operator)), c.Request.Sequence, c.Response.Header, pbData)
	return c.Response.Write(p.Bytes())
}

// 对客户端请求进行响应
func (c *Context) write(body proto.Message) (int, error) {
	pbData, err := proto.Marshal(body)
	if err != nil {
		return 0, err
	}

	p := NewPack(c.Request.OperateType, c.Request.Sequence, c.Response.Header, pbData)

	return c.Response.Write(p.Bytes())
}
