package client

import (
	"hash/crc32"
	"net"

	"github.com/wpajqz/linker"
)

type Handler func(*Context)

type Client struct {
	handlerContainer map[uint32]Handler
	packet           chan linker.Packet
	protocolPacket   linker.Packet
}

func NewClient() *Client {
	return &Client{
		handlerContainer: make(map[uint32]Handler),
		packet:           make(chan linker.Packet, 100),
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

func (c *Client) Handle(pattern string, handler Handler) {
	data := []byte(pattern)
	operator := crc32.ChecksumIEEE(data)

	if _, ok := c.handlerContainer[operator]; !ok {
		c.handlerContainer[operator] = handler
	}
}

func (c *Client) BindRouter(routers []Router) {
	for _, router := range routers {
		c.Handle(router.Operator, router.Handler)
	}
}

func (c *Client) SetProtocolPacket(packet linker.Packet) {
	c.protocolPacket = packet
}

func (c *Client) Send(operator string, pb interface{}) error {
	data := []byte(operator)
	op := crc32.ChecksumIEEE(data)

	p, err := c.protocolPacket.Pack(op, pb)
	if err != nil {
		return err
	}

	c.packet <- p
	return nil
}
