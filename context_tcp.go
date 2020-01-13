package linker

import (
	"context"
	"hash/crc32"
	"net"
	"runtime"
	"strconv"

	"github.com/wpajqz/linker/codec"
)

var _ Context = new(ContextTcp)

type ContextTcp struct {
	common
	Conn net.Conn
}

func NewContextTcp(ctx context.Context, conn net.Conn, OperateType uint32, Sequence int64, Header, Body []byte, options Options) *ContextTcp {
	return &ContextTcp{
		common: common{
			options:     options,
			operateType: OperateType,
			sequence:    Sequence,
			Context:     ctx,
			Request:     struct{ Header, Body []byte }{Header: Header, Body: Body},
			body:        Body,
		},
		Conn: conn,
	}
}

// 响应请求成功的数据包
func (c *ContextTcp) Success(body interface{}) {
	r, err := codec.NewCoder(c.options.contentType)
	if err != nil {
		panic(err)
	}

	data, err := r.Encoder(body)
	if err != nil {
		panic(err)
	}

	p, err := NewPacket(c.operateType, c.sequence, c.Response.Header, data, c.options.pluginForPacketSender)
	if err != nil {
		panic(err)
	}

	_, _ = c.Conn.Write(p.Bytes())

	runtime.Goexit()
}

// 响应请求失败的数据包
func (c *ContextTcp) Error(code int, message string) {
	c.SetResponseProperty("code", strconv.Itoa(code))
	c.SetResponseProperty("message", message)

	p, err := NewPacket(c.operateType, c.sequence, c.Response.Header, nil, c.options.pluginForPacketSender)

	if err != nil {
		panic(err)
	}

	_, _ = c.Conn.Write(p.Bytes())

	runtime.Goexit()
}

// 向客户端发送数据
func (c *ContextTcp) Write(operator string, body []byte) (int, error) {
	p, err := NewPacket(crc32.ChecksumIEEE([]byte(operator)), 0, c.Response.Header, body, c.options.pluginForPacketSender)
	if err != nil {
		return 0, err
	}

	return c.Conn.Write(p.Bytes())
}

func (c *ContextTcp) LocalAddr() string {
	return c.Conn.LocalAddr().String()
}

func (c *ContextTcp) RemoteAddr() string {
	return c.Conn.RemoteAddr().String()
}
