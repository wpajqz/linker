package client

import (
	"fmt"
	"io"
	"net"

	"github.com/wpajqz/linker"
	"github.com/wpajqz/linker/utils"
)

func (c *Client) handleConnection(conn net.Conn) {
	quit := make(chan bool)
	defer func() { quit <- true }()

	receivePackets := make(chan linker.Packet, 100)
	go c.handlePackets(conn, receivePackets, quit)
	go c.handleSendPackets(conn, quit)

	var (
		bLen   []byte = make([]byte, 4)
		bType  []byte = make([]byte, 4)
		pacLen int32
	)

	for {

		if n, err := io.ReadFull(conn, bLen); err != nil && n != 4 {
			fmt.Println(err.Error())
			return
		}

		if n, err := io.ReadFull(conn, bType); err != nil && n != 4 {
			fmt.Println(err.Error())
			return
		}

		if pacLen = utils.BytesToInt32(bLen); pacLen > 2048 {
			fmt.Println(pacLen)
			return
		}

		dataLength := pacLen - 8
		data := make([]byte, dataLength)
		if n, err := io.ReadFull(conn, data); err != nil && n != int(dataLength) {
			fmt.Println(err.Error())
			return
		}

		receivePackets <- c.protocolPacket.New(pacLen, utils.BytesToInt32(bType), data)
	}
}

func (c *Client) handlePackets(conn net.Conn, receivePackets <-chan linker.Packet, quit <-chan bool) {
	for {
		select {
		case p := <-receivePackets:
			handler, ok := c.handlerContainer[p.OperateType()]
			if !ok {
				continue
			}

			ctx := &Context{conn, p.OperateType(), p}

			go handler(ctx)
		case <-quit:
			return
		}
	}
}

func (c *Client) handleSendPackets(conn net.Conn, quit <-chan bool) {
	for {
		select {
		case p := <-c.packet:
			conn.Write(p.Bytes())
		case <-quit:
			return
		}
	}
}
