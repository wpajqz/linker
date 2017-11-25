package export

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
	"strconv"
	"time"

	"github.com/wpajqz/linker/utils/convert"
	"github.com/wpajqz/linker/utils/encrypt"
)

// 处理客户端连接
func (c *Client) handleConnection(conn net.Conn) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer func(cancel context.CancelFunc) { cancel() }(cancel)

	go c.handleSendPackets(ctx, conn)

	return c.handleReceivedPackets(conn)
}

// 对发送的数据包进行处理
func (c *Client) handleSendPackets(ctx context.Context, conn net.Conn) {
	for {
		select {
		case p := <-c.packet:
			if c.debug {
				log.Println("[send packet]", "operator:", p.OperateType(), "header:", string(p.Header()), "body:", string(p.Body()))
			}
			_, err := conn.Write(p.Bytes())
			if err != nil {
				if c.errorHandler != nil {
					_, file, line, _ := runtime.Caller(0)
					s := fmt.Sprintf("[datetime]:%v [file]:%v [line]:%v [message]:%v", time.Now(), file, line, err.Error())
					c.errorHandler.Handle(s)
				}
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

		nType := convert.BytesToUint32(bType)
		sequence = convert.BytesToInt64(bSequence)
		headerLength = convert.BytesToUint32(bHeaderLength)
		bodyLength = convert.BytesToUint32(bBodyLength)

		pacLen = headerLength + bodyLength + 20
		if pacLen > MaxPayload {
			return fmt.Errorf("the packet is big than %v", strconv.Itoa(MaxPayload))
		}

		header := make([]byte, headerLength)
		if n, err := io.ReadFull(conn, header); err != nil && n != int(headerLength) {
			return err
		}

		body := make([]byte, bodyLength)
		if n, err := io.ReadFull(conn, body); err != nil && n != int(bodyLength) {
			return err
		}

		header, err := encrypt.Decrypt(header)
		if err != nil {
			return err
		}

		body, err = encrypt.Decrypt(body)
		if err != nil {
			return err
		}

		c.response.Header = header
		c.response.Body = body

		if c.debug {
			log.Println("[receive packet]", "operator:", nType, "header:", string(header), "body:", string(body))
		}

		operator := int64(nType) + sequence
		if handler, ok := c.handlerContainer.Load(operator); ok {
			if v, ok := handler.(Handler); ok {
				v.Handle(header, body)
			}
		}
	}
}
