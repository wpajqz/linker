package linker

import (
	"context"
	"hash/crc32"
	"runtime"
	"strconv"
	"time"

	"github.com/wpajqz/linker/coder"
)

type (
	Context struct {
		context.Context
		Request     *request
		Response    response
		contentType string
	}
)

func NewContext(ctx context.Context, req *request, res response, contentType string) *Context {
	return &Context{Context: ctx, Request: req, Response: res, contentType: contentType}
}

func (c *Context) ParseParam(data interface{}) error {
	r, err := coder.NewCoder(c.contentType)
	if err != nil {
		return err
	}

	return r.Decoder(c.Request.Body, data)
}

// 设置单个请求可以处理的序列化数据类型，还可以在中间件中更改
func (c *Context) SetContentType(contentType string) {
	c.contentType = contentType
}

// 响应请求成功的数据包
func (c *Context) Success(body interface{}) {
	r, err := coder.NewCoder(c.contentType)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		panic(SystemError{time.Now(), file, line, err.Error()})
	}

	data, err := r.Encoder(body)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		panic(SystemError{time.Now(), file, line, err.Error()})
	}

	p := NewPack(c.Request.OperateType, c.Request.Sequence, c.Response.Header, data)
	c.Response.Write(p.Bytes())

	panic(nil)
}

// 响应请求失败的数据包
func (c *Context) Error(code int, message string) {
	c.Response.SetResponseProperty("code", strconv.Itoa(code))
	c.Response.SetResponseProperty("message", message)

	p := NewPack(c.Request.OperateType, c.Request.Sequence, c.Response.Header, nil)
	c.Response.Write(p.Bytes())

	panic(nil)
}

// 向客户端发送数据
func (c *Context) Write(operator string, body interface{}) (int, error) {
	r, err := coder.NewCoder(c.contentType)
	if err != nil {
		return 0, err
	}

	data, err := r.Encoder(body)
	if err != nil {
		return 0, err
	}

	p := NewPack(crc32.ChecksumIEEE([]byte(operator)), 0, c.Response.Header, data)

	return c.Response.Write(p.Bytes())
}

// 向客户端发送原始数据数据
func (c *Context) WriteBinary(operator string, data []byte) (int, error) {
	p := NewPack(crc32.ChecksumIEEE([]byte(operator)), 0, c.Response.Header, data)

	return c.Response.Write(p.Bytes())
}
