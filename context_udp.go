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

type ContextUdp struct {
	config            Config
	operateType       uint32
	sequence          int64
	body              []byte
	remote            *net.UDPAddr
	Context           context.Context
	Conn              *net.UDPConn
	Request, Response struct {
		Header, Body []byte
	}
}

func NewContextUdp(conn *net.UDPConn, remote *net.UDPAddr, OperateType uint32, Sequence int64, Header, Body []byte, config Config) *ContextUdp {
	return &ContextUdp{
		Context:     context.Background(),
		Conn:        conn,
		remote:      remote,
		operateType: OperateType,
		sequence:    Sequence,
		config:      config,
		Request:     struct{ Header, Body []byte }{Header: Header, Body: Body},
		body:        Body,
	}
}

func (c *ContextUdp) WithValue(key interface{}, value interface{}) Context {
	c.Context = context.WithValue(c.Context, key, value)
	return c
}

func (c *ContextUdp) Value(key interface{}) interface{} {
	return c.Context.Value(key)
}

func (c *ContextUdp) ParseParam(data interface{}) error {
	r, err := codec.NewCoder(c.config.ContentType)
	if err != nil {
		return err
	}

	return r.Decoder(c.body, data)
}

// 响应请求成功的数据包
func (c *ContextUdp) Success(body interface{}) {
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

	c.Conn.WriteToUDP(p.Bytes(), c.remote)

	runtime.Goexit()
}

// 响应请求失败的数据包
func (c *ContextUdp) Error(code int, message string) {
	c.SetResponseProperty("code", strconv.Itoa(code))
	c.SetResponseProperty("message", message)

	p, err := NewPacket(c.operateType, c.sequence, c.Response.Header, nil, c.config.PluginForPacketSender)

	if err != nil {
		panic(err)
	}

	c.Conn.WriteToUDP(p.Bytes(), c.remote)

	runtime.Goexit()
}

// 向客户端发送数据
func (c *ContextUdp) Write(operator string, body interface{}) (int, error) {
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

	return c.Conn.WriteToUDP(p.Bytes(), c.remote)
}

func (c *ContextUdp) SetRequestProperty(key, value string) {
	v := c.GetRequestProperty(key)
	if v != "" {
		c.Request.Header = bytes.Trim(c.Request.Header, key+"="+value+";")
	}

	c.Request.Header = append(c.Request.Header, []byte(key+"="+value+";")...)
}

func (c *ContextUdp) GetRequestProperty(key string) string {
	values := strings.Split(string(c.Request.Header), ";")
	for _, value := range values {
		kv := strings.Split(value, "=")
		if kv[0] == key {
			return kv[1]
		}
	}

	return ""
}

func (c *ContextUdp) SetResponseProperty(key, value string) {
	v := c.GetResponseProperty(key)
	if v != "" {
		c.Response.Header = bytes.Trim(c.Response.Header, key+"="+value+";")
	}

	c.Response.Header = append(c.Response.Header, []byte(key+"="+value+";")...)
}

func (c *ContextUdp) GetResponseProperty(key string) string {
	values := strings.Split(string(c.Response.Header), ";")
	for _, value := range values {
		kv := strings.Split(value, "=")
		if kv[0] == key {
			return kv[1]
		}
	}

	return ""
}

func (c *ContextUdp) LocalAddr() string {
	return c.Conn.LocalAddr().String()
}

func (c *ContextUdp) RemoteAddr() string {
	return c.Conn.RemoteAddr().String()
}
