package export

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/wpajqz/linker"
	"github.com/wpajqz/linker/utils/convert"
	"golang.org/x/sync/errgroup"
)

const (
	networkTCP = "tcp"
	networkUDP = "udp"
)

// handleConnection 处理客户端连接
func (c *Client) handleConnection(network string, conn net.Conn) {
	eg, ctx := errgroup.WithContext(context.Background())

	eg.Go(func() error {
		var err error

		switch network {
		case networkTCP:
			err = c.handleReceivedTCPPackets(conn)
		case networkUDP:
			err = c.handleReceivedUDPPackets(conn)
		default:
			panic("unsupported network, must be tcp or udp")
		}

		return err
	})

	eg.Go(func() error {
		return c.handleSendPackets(ctx, conn)
	})

	// wait one second for receive and send routine loaded
	time.Sleep(time.Second)
	go c.readyStateCallback.OnOpen()

	err := eg.Wait()
	if err != nil {
		c.readyState = CLOSED
		if err == io.EOF {
			c.readyStateCallback.OnClose()
		} else {
			c.readyStateCallback.OnError(err.Error())
		}
		_ = c.Close()
	}
}

// handleSendPackets 对发送的数据包进行处理
func (c *Client) handleSendPackets(ctx context.Context, conn net.Conn) error {
	for {
		select {
		case p := <-c.packet:
			_, err := conn.Write(p.Bytes())
			if err != nil {
				return err
			}

			if c.timeout != 0 {
				err := conn.SetWriteDeadline(time.Now().Add(c.timeout))
				if err != nil {
					return err
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// handleReceivedUDPPackets 对接收到的数据包进行处理
func (c *Client) handleReceivedUDPPackets(conn net.Conn) error {
	udpConn := conn.(*net.UDPConn)
	for {
		if c.timeout != 0 {
			err := conn.SetReadDeadline(time.Now().Add(c.timeout))
			if err != nil {
				return err
			}
		}

		data := make([]byte, 0, c.maxPayload)
		n, _, err := udpConn.ReadFromUDP(data)
		if err != nil {
			continue
		}

		bType := data[0:4]
		bSequence := data[4:12]
		bHeaderLength := data[12:16]

		sequence := convert.BytesToInt64(bSequence)
		headerLength := convert.BytesToUint32(bHeaderLength)

		header := data[20 : 20+headerLength]
		body := data[20+headerLength : n]

		receive, err := linker.NewPacket(convert.BytesToUint32(bType), sequence, header, body, c.pluginForPacketReceiver)
		if err != nil {
			return err
		}

		c.response.Header = receive.Header
		c.response.Body = receive.Body

		operator := int64(convert.BytesToUint32(bType)) + sequence
		if handler, ok := c.handlerContainer.Load(operator); ok {
			if v, ok := handler.(Handler); ok {
				v.Handle(receive.Header, receive.Body)
			}
		}
	}
}

// handleReceivedTCPPackets 对接收到的数据包进行处理
func (c *Client) handleReceivedTCPPackets(conn net.Conn) error {
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
		if c.timeout != 0 {
			err := conn.SetReadDeadline(time.Now().Add(c.timeout))
			if err != nil {
				return err
			}
		}

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
		if pacLen > uint32(c.maxPayload) {
			return fmt.Errorf("the packet is big than %v\n", strconv.Itoa(c.maxPayload))
		}

		header := make([]byte, headerLength)
		if n, err := io.ReadFull(conn, header); err != nil && n != int(headerLength) {
			return err
		}

		body := make([]byte, bodyLength)
		if n, err := io.ReadFull(conn, body); err != nil && n != int(bodyLength) {
			return err
		}

		receive, err := linker.NewPacket(nType, sequence, header, body, c.pluginForPacketReceiver)
		if err != nil {
			return err
		}

		c.response.Header = receive.Header
		c.response.Body = receive.Body

		operator := int64(nType) + sequence
		if handler, ok := c.handlerContainer.Load(operator); ok {
			if v, ok := handler.(Handler); ok {
				v.Handle(receive.Header, receive.Body)
			}
		}
	}
}
