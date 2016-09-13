package client

import (
	"hash/crc32"
	"net"
	"sync"
	"time"

	"github.com/wpajqz/linker"
)

type Handler func(*Context)

type packetContainer struct {
	sync.RWMutex
	m map[uint32]linker.Packet
}

type Client struct {
	readTimeout, writeTimeout time.Duration
	conn                      net.Conn
	packet                    chan linker.Packet
	receivePackets            packetContainer
	protocolPacket            linker.Packet
}

func NewClient(network, address string) *Client {
	client := &Client{
		packet:         make(chan linker.Packet, 100),
		receivePackets: packetContainer{m: make(map[uint32]linker.Packet)},
	}

	conn, err := net.Dial(network, address)
	if err != nil {
		panic("start client:" + err.Error())
	}

	client.conn = conn
	go client.handleConnection(client.conn)

	return client
}

func (c *Client) SetProtocolPacket(packet linker.Packet) {
	c.protocolPacket = packet
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) SetTimeout(timeout time.Duration) {
	c.readTimeout = timeout
	c.writeTimeout = timeout
}

func (c *Client) SetReadTimeout(readTimeout time.Duration) {
	c.readTimeout = readTimeout
}

func (c *Client) SetWriteTimeout(writeTimeout time.Duration) {
	c.writeTimeout = writeTimeout
}

func (c *Client) SyncCall(operator string, pb interface{}, response func(*Context)) error {
	data := []byte(operator)
	op := crc32.ChecksumIEEE(data)

	p, err := c.protocolPacket.Pack(op, pb)
	if err != nil {
		return err
	}

	c.packet <- p

	for {
		if rp, ok := c.receivePackets.m[op]; ok {
			response(&Context{op, rp})
			return nil
		}

		continue
	}

	return nil
}

func (c *Client) AsyncCall(operator string, pb interface{}, response func(*Context)) error {
	data := []byte(operator)
	op := crc32.ChecksumIEEE(data)

	p, err := c.protocolPacket.Pack(op, pb)
	if err != nil {
		return err
	}

	c.packet <- p

	go func() {
		for {
			if rp, ok := c.receivePackets.m[op]; ok {
				response(&Context{op, rp})
				return
			}

			continue
		}
	}()

	return nil
}
