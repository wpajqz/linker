package linker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"runtime"
	"strconv"
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
	switch c.Request.GetRequestProperty("Content-Type") {
	case "text/json":
		err := json.Unmarshal(c.Request.Body, data)
		if err != nil {
			return fmt.Errorf("Unpack error: %v", err.Error())
		}
	case "text/protobuf":
		err := proto.Unmarshal(c.Request.Body, data.(proto.Message))
		if err != nil {
			return fmt.Errorf("Unpack error: %v", err.Error())
		}

	default:
		return errors.New("Unsupported data type")
	}

	return nil
}

// 响应请求成功的数据包
func (c *Context) Success(body interface{}) {
	var (
		data []byte
		err  error
	)

	switch c.Request.GetRequestProperty("Content-Type") {
	case "text/json":
		data, err = json.Marshal(body)
	case "text/protobuf":
		data, err = proto.Marshal(body.(proto.Message))

	default:
		err = errors.New("Unsupported data type")
	}

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
	var (
		data []byte
		err  error
	)

	switch c.Request.GetRequestProperty("Content-Type") {
	case "text/protobuf":
		data, err = proto.Marshal(body.(proto.Message))

	case "text/json":
		data, err = json.Marshal(body)
	default:
		err = errors.New("Unsupported data type")
	}

	if err != nil {
		return 0, err
	}

	p := NewPack(crc32.ChecksumIEEE([]byte(operator)), c.Request.Sequence, c.Response.Header, data)

	return c.Response.Write(p.Bytes())
}
