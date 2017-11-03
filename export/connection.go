package export

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/wpajqz/linker"
)

// 处理客户端连接
func (c *Client) handleConnection(conn net.Conn) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		if err := recover(); err != nil {
			if c.errorHandler != nil {
				c.errorHandler.Handle(err.(error).Error())
			}
		}

		cancel()
	}()

	if c.constructHandler != nil {
		c.constructHandler.Handle(nil, nil)
	}

	go c.handleSendPackets(ctx, conn)

	err := c.handleReceivedPackets(conn)
	if err != nil {
		if err == io.EOF {
			if c.destructHandler != nil {
				c.destructHandler.Handle(nil, nil)
			}
		}
	}

	return err
}

// 对发送的数据包进行处理
func (c *Client) handleSendPackets(ctx context.Context, conn net.Conn) {
	for {
		select {
		case p := <-c.packet:
			_, err := conn.Write(p.Bytes())
			if err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// 对接收到的数据包进行处理
func (c *Client) handleReceivedPackets(conn net.Conn) error {
	var (
		bType         = make([]byte, 4)
		bSequence     = make([]byte, 8)
		bHeaderLength = make([]byte, 4)
		bBodyLength   = make([]byte, 4)
		sequence      int64
		headerLength  uint32
		bodyLength    uint32
		pacLen        uint32
	)

	for {
		if n, err := io.ReadFull(conn, bType); err != nil && n != 4 {
			return err
		}

		if n, err := io.ReadFull(conn, bSequence); err != nil && n != 8 {
			return err
		}

		if n, err := io.ReadFull(conn, bHeaderLength); err != nil && n != 4 {
			return err
		}

		if n, err := io.ReadFull(conn, bBodyLength); err != nil && n != 4 {
			return err
		}

		nType := linker.BytesToUint32(bType)
		sequence = linker.BytesToInt64(bSequence)
		headerLength = linker.BytesToUint32(bHeaderLength)
		bodyLength = linker.BytesToUint32(bBodyLength)

		pacLen = headerLength + bodyLength + 20
		if pacLen > MaxPayload {
			return fmt.Errorf("the packet is big than %v" + strconv.Itoa(MaxPayload))
		}

		header := make([]byte, headerLength)
		if n, err := io.ReadFull(conn, header); err != nil && n != int(headerLength) {
			return err
		}

		body := make([]byte, bodyLength)
		if n, err := io.ReadFull(conn, body); err != nil && n != int(bodyLength) {
			return err
		}

		c.response.Header = header
		c.response.Body = body

		operator := int64(nType) + sequence
		if handler, ok := c.handlerContainer.Load(operator); ok {
			if v, ok := handler.(Handler); ok {
				v.Handle(header, body)
			}
		}
	}
}
