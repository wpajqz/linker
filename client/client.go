package client

import (
	"net"

	"hash/crc32"

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

func (c *Client) Handle(pattern uint32, handler Handler) {
	if _, ok := c.handlerContainer[pattern]; !ok {
		c.handlerContainer[pattern] = handler
	}
}

func (c *Client) BindRouter(routers []Router) {
	for _, router := range routers {
		data := []byte(router.Operator)
		operator := crc32.ChecksumIEEE(data)

		c.Handle(operator, router.Handler)
	}
}

func (c *Client) Send(operator uint32, pb interface{}) error {
	p, err := c.protocolPacket.Pack(operator, pb)
	if err != nil {
		return err
	}

	c.packet <- p
	return nil
}
