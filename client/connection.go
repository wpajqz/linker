package client

import (
	"io"
	"net"

	"github.com/wpajqz/linker/utils"
)

func (c *Client) handleConnection(conn net.Conn) error {
	quit := make(chan bool)
	defer func() { quit <- true }()

	go c.handleSendPackets(conn, quit)

	var (
		bLen   []byte = make([]byte, 4)
		bType  []byte = make([]byte, 4)
		pacLen uint32
	)

	for {

		if n, err := io.ReadFull(conn, bLen); err != nil && n != 4 {
			return err
		}

		if n, err := io.ReadFull(conn, bType); err != nil && n != 4 {
			return err
		}

		if pacLen = utils.BytesToUint32(bLen); pacLen > 2048 {
			return ErrPacketLength
		}

		dataLength := pacLen - 8
		data := make([]byte, dataLength)
		if n, err := io.ReadFull(conn, data); err != nil && n != int(dataLength) {
			return err
		}

		c.receivePackets <- c.protocolPacket.New(pacLen, utils.BytesToUint32(bType), data)
	}
}

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
