package client

import (
	"hash/crc32"
	"net"

	"github.com/wpajqz/linker"
)

type Handler func(*Context)

type Client struct {
	packet         chan linker.Packet
	receivePackets map[uint32]linker.Packet
	protocolPacket linker.Packet
}

func NewClient() *Client {
	return &Client{
		packet:         make(chan linker.Packet, 100),
		receivePackets: make(map[uint32]linker.Packet, 100),
	}
}

func (c *Client) Run(network, address string) {
	conn, err := net.Dial(network, address)
	if err != nil {
		panic("start client:" + err.Error())
	}
	defer conn.Close()

	c.handleConnection(conn)
}

func (c *Client) SetProtocolPacket(packet linker.Packet) {
	c.protocolPacket = packet
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
		if rp, ok := c.receivePackets[op]; ok {
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
			if rp, ok := c.receivePackets[op]; ok {
				response(&Context{op, rp})
				return
			}

			continue
		}
	}()

	return nil
}
