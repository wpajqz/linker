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
	running                bool
	timeout                time.Duration
	conn                   net.Conn
	packet, receivePackets chan linker.Packet
	protocolPacket         linker.Packet
	quit                   chan bool
	mutex                  sync.RWMutex
}

func NewClient(network, address string) *Client {
	client := &Client{
		timeout:        30 * time.Second,
		packet:         make(chan linker.Packet, 100),
		receivePackets: make(chan linker.Packet, 100),
		quit:           make(chan bool),
	}

	go func(string, string) {
		for {
			if client.running {
				err := client.handleConnection(client.conn)
				if err != nil {
					client.setRunningStatus(false)
				}
			} else {
				for {
					println(1)
					conn, err := net.Dial(network, address)
					println(2)
					if err == nil {
						client.setRunningStatus(true)
						client.conn = conn

						break
					}

					time.Sleep(client.timeout)
				}
			}
		}
	}(network, address)

	return client
}

func (c *Client) SetProtocolPacket(packet linker.Packet) {
	c.protocolPacket = packet
}

func (c *Client) Close() error {
	c.quit <- true
	return c.conn.Close()
}

func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

func (c *Client) SyncCall(operator string, pb interface{}, response func(*Context)) error {
	if c.RunningStatus() {
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
	}

	return ErrClosed
}

func (c *Client) AsyncCall(operator string, pb interface{}, response func(*Context)) error {
	if c.RunningStatus() {
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
	}

	return ErrClosed
}

func (c *Client) Heartbeat(interval time.Duration, pb interface{}) error {
	if c.RunningStatus() {
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
			case <-c.quit:
				return nil
			}
		}
	}

	return ErrClosed
}

func (c *Client) RunningStatus() bool {
	c.mutex.RLock()
	r := c.running
	c.mutex.RUnlock()
	return r
}

func (c *Client) setRunningStatus(status bool) {
	c.mutex.Lock()
	c.running = status
	c.mutex.Unlock()
}
