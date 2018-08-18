package linker

import (
	"bytes"
	"context"
	"hash/crc32"
	"net"
	"runtime"
	"strconv"
	"strings"

	"github.com/wpajqz/linker/codec"
)

type ContextTcp struct {
	config            Config
	operateType       uint32
	sequence          int64
	body              []byte
	Context           context.Context
	Conn              net.Conn
	Request, Response struct {
		Header, Body []byte
	}
}

func NewContextTcp(conn net.Conn, OperateType uint32, Sequence int64, Header, Body []byte, config Config) *ContextTcp {
	return &ContextTcp{
		Context:     context.Background(),
		Conn:        conn,
		operateType: OperateType,
		sequence:    Sequence,
		config:      config,
		Request:     struct{ Header, Body []byte }{Header: Header, Body: Body},
		body:        Body,
	}
}

func (c *ContextTcp) WithValue(key interface{}, value interface{}) Context {
	c.Context = context.WithValue(c.Context, key, value)
	return c
}

func (c *ContextTcp) Value(key interface{}) interface{} {
	return c.Context.Value(key)
}

func (c *ContextTcp) ParseParam(data interface{}) error {
	r, err := codec.NewCoder(c.config.ContentType)
	if err != nil {
		return err
	}

	return r.Decoder(c.body, data)
}

// 响应请求成功的数据包
func (c *ContextTcp) Success(body interface{}) {
	r, err := codec.NewCoder(c.config.ContentType)
	if err != nil {
		panic(err)
	}

	data, err := r.Encoder(body)
	if err != nil {
		panic(err)
	}

	p, err := NewPacket(c.operateType, c.sequence, c.Response.Header, data, c.config.PluginForPacketSender)

	if err != nil {
		panic(err)
	}

	c.Conn.Write(p.Bytes())

	runtime.Goexit()
}

// 响应请求失败的数据包
func (c *ContextTcp) Error(code int, message string) {
	c.SetResponseProperty("code", strconv.Itoa(code))
	c.SetResponseProperty("message", message)

	p, err := NewPacket(c.operateType, c.sequence, c.Response.Header, nil, c.config.PluginForPacketSender)

	if err != nil {
		panic(err)
	}

	c.Conn.Write(p.Bytes())

	runtime.Goexit()
}

// 向客户端发送数据
func (c *ContextTcp) Write(operator string, body interface{}) (int, error) {
	r, err := codec.NewCoder(c.config.ContentType)
	if err != nil {
		return 0, err
	}

	data, err := r.Encoder(body)
	if err != nil {
		return 0, err
	}

	p, err := NewPacket(crc32.ChecksumIEEE([]byte(operator)), 0, c.Response.Header, data, c.config.PluginForPacketSender)

	if err != nil {
		panic(err)
	}

	return c.Conn.Write(p.Bytes())
}

func (c *ContextTcp) SetRequestProperty(key, value string) {
	v := c.GetRequestProperty(key)
	if v != "" {
		c.Request.Header = bytes.Trim(c.Request.Header, key+"="+value+";")
	}

	c.Request.Header = append(c.Request.Header, []byte(key+"="+value+";")...)
}

func (c *ContextTcp) GetRequestProperty(key string) string {
	values := strings.Split(string(c.Request.Header), ";")
	for _, value := range values {
		kv := strings.Split(value, "=")
		if kv[0] == key {
			return kv[1]
		}
	}

	return ""
}

func (c *ContextTcp) SetResponseProperty(key, value string) {
	v := c.GetResponseProperty(key)
	if v != "" {
		c.Response.Header = bytes.Trim(c.Response.Header, key+"="+value+";")
	}

	c.Response.Header = append(c.Response.Header, []byte(key+"="+value+";")...)
}

func (c *ContextTcp) GetResponseProperty(key string) string {
	values := strings.Split(string(c.Response.Header), ";")
	for _, value := range values {
		kv := strings.Split(value, "=")
		if kv[0] == key {
			return kv[1]
		}
	}

	return ""
}

func (c *ContextTcp) LocalAddr() string {
	return c.Conn.LocalAddr().String()
}

func (c *ContextTcp) RemoteAddr() string {
	return c.Conn.RemoteAddr().String()
}
