package linker

import (
	"context"
	"hash/crc32"
	"runtime"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/wpajqz/linker/codec"
)

var _ Context = new(ContextWebsocket)

type ContextWebsocket struct {
	common
	Conn *webSocketConn
}

func NewContextWebsocket(ctx context.Context, conn *webSocketConn, OperateType uint32, Sequence int64, Header, Body []byte, options Options) *ContextWebsocket {
	return &ContextWebsocket{
		common: common{
			Context:     ctx,
			operateType: OperateType,
			sequence:    Sequence,
			options:     options,
			Request:     struct{ Header, Body []byte }{Header: Header, Body: Body},
			body:        Body,
		},
		Conn: conn,
	}
}

// 响应请求成功的数据包
func (c *ContextWebsocket) Success(body interface{}) {
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

	_ = c.Conn.WriteMessage(websocket.BinaryMessage, p.Bytes())

	runtime.Goexit()
}

// 响应请求失败的数据包
func (c *ContextWebsocket) Error(code int, message string) {
	c.SetResponseProperty("code", strconv.Itoa(code))
	c.SetResponseProperty("message", message)

	p, err := NewPacket(c.operateType, c.sequence, c.Response.Header, nil, c.options.pluginForPacketSender)

	if err != nil {
		panic(err)
	}

	_ = c.Conn.WriteMessage(websocket.BinaryMessage, p.Bytes())

	runtime.Goexit()
}

// 向客户端发送数据
func (c *ContextWebsocket) Write(operator string, body []byte) (int, error) {
	p, err := NewPacket(crc32.ChecksumIEEE([]byte(operator)), 0, c.Response.Header, body, c.options.pluginForPacketSender)
	if err != nil {
		return 0, err
	}

	return 0, c.Conn.WriteMessage(websocket.BinaryMessage, p.Bytes())
}

func (c *ContextWebsocket) LocalAddr() string {
	return c.Conn.LocalAddr().String()
}

func (c *ContextWebsocket) RemoteAddr() string {
	return c.Conn.RemoteAddr().String()
}
