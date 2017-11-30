package linker

import (
	"bytes"
	"context"
	"hash/crc32"
	"runtime"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"

	"github.com/wpajqz/linker/codec"
)

var _ Context = new(ContextWebsocket)

type (
	ContextWebsocket struct {
		operateType       uint32
		sequence          int64
		body              []byte
		contentType       string
		Context           context.Context
		Conn              *websocket.Conn
		Request, Response struct {
			Header, Body []byte
		}
	}
)

func NewContextWebsocket(conn *websocket.Conn, OperateType uint32, Sequence int64, contentType string, Header, Body []byte) *ContextWebsocket {
	return &ContextWebsocket{
		Context:     context.Background(),
		Conn:        conn,
		operateType: OperateType,
		sequence:    Sequence,
		contentType: contentType,
		Request:     struct{ Header, Body []byte }{Header: Header, Body: Body},
		body:        Body,
	}
}

func (c *ContextWebsocket) WithValue(key interface{}, value interface{}) Context {
	c.Context = context.WithValue(c.Context, key, value)
	return c
}

func (c *ContextWebsocket) Value(key interface{}) interface{} {
	return c.Context.Value(key)
}

// 设置单个请求可以处理的序列化数据类型，还可以在中间件中更改
func (c *ContextWebsocket) SetContentType(contentType string) {
	c.contentType = contentType
}

func (c *ContextWebsocket) ParseParam(data interface{}) error {
	r, err := codec.NewCoder(c.contentType)
	if err != nil {
		return err
	}

	return r.Decoder(c.body, data)
}

// 响应请求成功的数据包
func (c *ContextWebsocket) Success(body interface{}) {
	r, err := codec.NewCoder(c.contentType)
	if err != nil {
		panic(err)
	}

	data, err := r.Encoder(body)
	if err != nil {
		panic(err)
	}

	p := NewPack(c.operateType, c.sequence, c.Response.Header, data)

	c.Conn.WriteMessage(websocket.TextMessage, p.Bytes())

	runtime.Goexit()
}

// 响应请求失败的数据包
func (c *ContextWebsocket) Error(code int, message string) {
	c.SetResponseProperty("code", strconv.Itoa(code))
	c.SetResponseProperty("message", message)

	p := NewPack(c.operateType, c.sequence, c.Response.Header, nil)

	c.Conn.WriteMessage(websocket.TextMessage, p.Bytes())

	runtime.Goexit()
}

// 向客户端发送数据
func (c *ContextWebsocket) Write(operator string, body interface{}) (int, error) {
	r, err := codec.NewCoder(c.contentType)
	if err != nil {
		return 0, err
	}

	data, err := r.Encoder(body)
	if err != nil {
		return 0, err
	}

	p := NewPack(crc32.ChecksumIEEE([]byte(operator)), 0, c.Response.Header, data)

	return 0, c.Conn.WriteMessage(websocket.TextMessage, p.Bytes())
}

// 向客户端发送原始数据数据
func (c *ContextWebsocket) WriteBinary(operator string, data []byte) (int, error) {
	p := NewPack(crc32.ChecksumIEEE([]byte(operator)), 0, c.Response.Header, data)

	return 0, c.Conn.WriteMessage(websocket.TextMessage, p.Bytes())
}

func (c *ContextWebsocket) SetRequestProperty(key, value string) {
	v := c.GetRequestProperty(key)
	if v != "" {
		c.Request.Header = bytes.Trim(c.Request.Header, key+"="+value+";")
	}

	c.Request.Header = append(c.Request.Header, []byte(key+"="+value+";")...)
}

func (c *ContextWebsocket) GetRequestProperty(key string) string {
	values := strings.Split(string(c.Request.Header), ";")
	for _, value := range values {
		kv := strings.Split(value, "=")
		if kv[0] == key {
			return kv[1]
		}
	}

	return ""
}

func (c *ContextWebsocket) SetResponseProperty(key, value string) {
	v := c.GetResponseProperty(key)
	if v != "" {
		c.Response.Header = bytes.Trim(c.Response.Header, key+"="+value+";")
	}

	c.Response.Header = append(c.Response.Header, []byte(key+"="+value+";")...)
}

func (c *ContextWebsocket) GetResponseProperty(key string) string {
	values := strings.Split(string(c.Response.Header), ";")
	for _, value := range values {
		kv := strings.Split(value, "=")
		if kv[0] == key {
			return kv[1]
		}
	}

	return ""
}
