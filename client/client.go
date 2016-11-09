package client

import (
	"fmt"
	"hash/crc32"
	"net"
	"sync"
	"time"

	"github.com/wpajqz/linker"
)

const MaxPayload = 2048

type Handler func(*Context)

type Client struct {
	mutex                  sync.RWMutex
	timeout                time.Duration
	conn                   net.Conn
	protocolPacket         linker.Packet
	packet, receivePackets chan linker.Packet
	listenerPackets        chan linker.Packet
	cancelHeartbeat        chan bool
	closeClient            chan bool
	running                chan bool
	removeMessageListener  chan bool
}

func NewClient(network, address string) *Client {
	client := &Client{
		timeout:               30 * time.Second,
		packet:                make(chan linker.Packet, 100),
		receivePackets:        make(chan linker.Packet, 100),
		listenerPackets:       make(chan linker.Packet, 100),
		removeMessageListener: make(chan bool, 1),
		cancelHeartbeat:       make(chan bool, 1),
		closeClient:           make(chan bool, 1),
		running:               make(chan bool, 1),
	}

	conn, err := net.Dial(network, address)
	if err != nil {
		panic(err.Error())
	}

	client.conn = conn

	go func(string, string, net.Conn) {
		for {
			err := client.handleConnection(client.conn)
			if err != nil {
				client.running <- false
			}

			select {
			case r := <-client.running:
				if !r {
					for {
						//服务端timeout设置影响链接延时时间
						conn, err := net.Dial(network, address)
						if err == nil {
							client.conn = conn
							break
						}
					}
				}
			case <-client.closeClient:
				return
			}
		}
	}(network, address, client.conn)

	return client
}

func (c *Client) Heartbeat(interval time.Duration, pb interface{}) error {
	data := []byte("heartbeat")
	op := crc32.ChecksumIEEE(data)

	p, err := c.protocolPacket.Pack(op, pb)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(interval * time.Second)
	for {
		select {
		case <-ticker.C:
			c.packet <- p
		case <-c.cancelHeartbeat:
			return nil
		}
	}
}

func (c *Client) SetProtocolPacket(packet linker.Packet) {
	c.protocolPacket = packet
}

func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

func (c *Client) Close() {
	c.cancelHeartbeat <- true
	c.closeClient <- true
	c.removeMessageListener <- true
	close(c.cancelHeartbeat)
	close(c.closeClient)
	close(c.cancelHeartbeat)
}

func (c *Client) SyncCall(operator string, pb interface{}, success func(*Context), error func(*Context)) error {
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
			ctx := &Context{rp.OperateType(), rp}
			if rp.OperateType() == op {
				success(ctx)
				return nil
			}

			if rp.OperateType() == uint32(0) {
				error(ctx)
				return nil
			}

			c.listenerPackets <- rp
		case <-time.After(c.timeout):
			return fmt.Errorf("can't handle %s", operator)
		}
	}
}

func (c *Client) AsyncCall(operator string, pb interface{}, success func(*Context), error func(*Context)) error {
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
			ctx := &Context{rp.OperateType(), rp}
			if rp.OperateType() == op {
				go success(ctx)
				return nil
			}

			if rp.OperateType() == uint32(0) {
				go error(ctx)
				return nil
			}

			c.listenerPackets <- rp
		case <-time.After(c.timeout):
			return fmt.Errorf("can't handle %s", operator)
		}
	}
}

func (c *Client) AddMessageListener(callback func(*Context)) error {
	for {
		select {
		case rp := <-c.listenerPackets:
			callback(&Context{rp.OperateType(), rp})
			return nil
		case <-c.removeMessageListener:
			return nil
		}
	}
}
