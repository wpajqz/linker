package client

import (
	"net"

	"github.com/wpajqz/linker"
)

type Handler func(*Context)

type Client struct {
	handlerContainer map[int32]Handler
	packet           chan linker.Packet
	protocolPacket   linker.Packet
}

func NewClient() *Client {
	return &Client{
		handlerContainer: make(map[int32]Handler),
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

func (c *Client) Handle(pattern int32, handler Handler) {
	if _, ok := c.handlerContainer[pattern]; !ok {
		c.handlerContainer[pattern] = handler
	}
}

func (c *Client) BindRouter(routers []Router) {
	for _, router := range routers {
		c.Handle(router.Operator, router.Handler)
	}
}

func (c *Client) Send(operator int32, pb interface{}) error {
	p, err := c.protocolPacket.Pack(operator, pb)
	if err != nil {
		return err
	}

	c.packet <- p
	return nil
}
