package client

import (
	"fmt"
	"hash/crc32"
	"net"
	"time"

	"github.com/wpajqz/linker"
)

type Handler func(*Context)

type Client struct {
	timeout                time.Duration
	conn                   net.Conn
	packet, receivePackets chan linker.Packet
	protocolPacket         linker.Packet
}

func NewClient(network, address string) *Client {
	client := &Client{
		timeout:        30 * time.Second,
		packet:         make(chan linker.Packet, 100),
		receivePackets: make(chan linker.Packet, 100),
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
	c.timeout = timeout
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
		select {
		case rp := <-c.receivePackets:
			if rp.OperateType() == op {
				response(&Context{op, rp})
				return nil
			}
		case <-time.After(c.timeout):
			return fmt.Errorf("can't handle %s", operator)
		}
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

	for {
		select {
		case rp := <-c.receivePackets:
			if rp.OperateType() == op {
				go response(&Context{op, rp})
				return nil
			}
		case <-time.After(c.timeout):
			return fmt.Errorf("can't handle %s", operator)
		}
	}

	return nil
}

func (c *Client) Heartbeat(interval time.Duration, pb interface{}) error {
	data := []byte("heartbeat")
	op := crc32.ChecksumIEEE(data)

	p, err := c.protocolPacket.Pack(op, pb)
	if err != nil {
		return err
	}

	for {
		timer := time.NewTimer(interval * time.Second)
		select {
		case <-timer.C:
			c.packet <- p
		case <-time.After(c.timeout):
			return fmt.Errorf("can't send %s", "heartbeat")
		}
	}

	return nil
}
