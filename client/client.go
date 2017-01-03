package client

import (
	"errors"
	"hash/crc32"
	"net"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/wpajqz/linker"
)

const MaxPayload = uint32(2048)

var (
	ErrClosed       = errors.New("connection is closed.")
	ErrPacketLength = errors.New("the packet is big than " + strconv.Itoa(int(MaxPayload)))
)

type Handler func(*Context)

type Client struct {
	Context          *Context
	timeout          time.Duration
	conn             net.Conn
	protocolPacket   linker.Packet
	handlerContainer map[uint32]Handler
	packet           chan linker.Packet
	cancelHeartbeat  chan bool
	closeClient      chan bool
	running          chan bool
}

func NewClient(network, address string) *Client {
	c := &Client{
		Context:          &Context{Request: &request{}, Response: response{}},
		timeout:          30 * time.Second,
		packet:           make(chan linker.Packet, 1024),
		handlerContainer: make(map[uint32]Handler),
		cancelHeartbeat:  make(chan bool, 1),
		closeClient:      make(chan bool, 1),
		running:          make(chan bool, 1),
	}

	conn, err := net.Dial(network, address)
	if err != nil {
		panic(err.Error())
	}

	c.conn = conn

	go func(string, string, net.Conn) {
		for {
			err := c.handleConnection(c.conn)
			if err != nil {
				c.running <- false
			}

			select {
			case r := <-c.running:
				if !r {
					for {
						//服务端timeout设置影响链接延时时间
						conn, err := net.Dial(network, address)
						if err == nil {
							c.conn = conn
							break
						}
					}
				}
			case <-c.closeClient:
				return
			}
		}
	}(network, address, c.conn)

	return c
}

func (c *Client) Heartbeat(interval time.Duration, param interface{}) error {
	pbData, err := proto.Marshal(param.(proto.Message))
	if err != nil {
		return err
	}

	p := linker.NewPack(linker.OPERATOR_HEARTBEAT, c.Context.Request.Header, pbData)

	// 建立连接以后就发送心跳包建立会话信息，后面的定期发送
	c.packet <- p
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

func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

func (c *Client) Close() {
	c.cancelHeartbeat <- true
	c.closeClient <- true
	close(c.cancelHeartbeat)
	close(c.closeClient)
}

// 向服务端发送请求，同步处理服务端返回结果
func (c *Client) SyncCall(operator string, param interface{}, callback func(*Context)) error {
	data := []byte(operator)
	op := crc32.ChecksumIEEE(data)

	pbData, err := proto.Marshal(param.(proto.Message))
	if err != nil {
		return err
	}
	p := linker.NewPack(op, c.Context.Request.Header, pbData)

	c.packet <- p

	// 对数据请求的返回状态进行处理,同步阻塞处理机制
	ch := make(chan bool)
	c.AddMessageListener(operator, func(ctx *Context) {
		callback(ctx)
		c.RemoveMessageListener(operator)
		ch <- true
	})

	<-ch

	return nil
}

// 向服务端发送请求，异步处理服务端返回结果
func (c *Client) AsyncCall(operator string, param interface{}, callback func(*Context)) error {
	data := []byte(operator)
	op := crc32.ChecksumIEEE(data)

	pbData, err := proto.Marshal(param.(proto.Message))
	if err != nil {
		return err
	}
	p := linker.NewPack(op, c.Context.Request.Header, pbData)

	c.packet <- p

	c.AddMessageListener(operator, func(ctx *Context) {
		callback(ctx)
		c.RemoveMessageListener(operator)
	})

	return nil
}

// 添加事件监听器
func (c *Client) AddMessageListener(listener string, callback func(*Context)) {
	operator := crc32.ChecksumIEEE([]byte(listener))
	c.handlerContainer[operator] = callback
}

// 移除事件监听器
func (c *Client) RemoveMessageListener(listener string) {
	operator := crc32.ChecksumIEEE([]byte(listener))
	delete(c.handlerContainer, operator)
}
