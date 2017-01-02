package client

import (
	"io"
	"net"

	"github.com/wpajqz/linker"
)

// 处理客户端连接
func (c *Client) handleConnection(conn net.Conn) error {
	qs := make(chan bool)
	defer func() { qs <- true }()

	go c.handleSendPackets(conn, qs)

	return c.handleReceivedPackets(conn)
}

// 对发送的数据包进行处理
func (c *Client) handleSendPackets(conn net.Conn, quit <-chan bool) {
	for {
		select {
		case p := <-c.packet:
			_, err := conn.Write(p.Bytes())
			if err != nil {
				return
			}
		case <-quit:
			return
		}
	}
}

// 对接收到的数据包进行处理
func (c *Client) handleReceivedPackets(conn net.Conn) error {
	var (
		bType         []byte = make([]byte, 4)
		bHeaderLength []byte = make([]byte, 4)
		bBodyLength   []byte = make([]byte, 4)
		headerLength  uint32
		bodyLength    uint32
		pacLen        uint32
	)

	for {
		if n, err := io.ReadFull(conn, bType); err != nil && n != 4 {
			return err
		}

		if n, err := io.ReadFull(conn, bHeaderLength); err != nil && n != 4 {
			return err
		}

		if n, err := io.ReadFull(conn, bBodyLength); err != nil && n != 4 {
			return err
		}

		headerLength = linker.BytesToUint32(bHeaderLength)
		bodyLength = linker.BytesToUint32(bBodyLength)

		pacLen = headerLength + bodyLength + 12
		if pacLen > MaxPayload {
			return ErrPacketLength
		}

		header := make([]byte, headerLength)
		if n, err := io.ReadFull(conn, header); err != nil && n != int(headerLength) {
			return err
		}

		body := make([]byte, bodyLength)
		if n, err := io.ReadFull(conn, body); err != nil && n != int(bodyLength) {
			return err
		}

		p := linker.NewPack(linker.BytesToUint32(bType), header, body)
		if handler, ok := c.handlerContainer[p.OperateType()]; ok {
			req := &request{Conn: conn, OperateType: p.OperateType()}
			res := response{Conn: conn, OperateType: p.OperateType(), Header: p.Header(), Body: p.Body()}

			ctx := &Context{req, res}
			handler(ctx)
		}
	}
}
