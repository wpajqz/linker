package client

import (
	"errors"
	"hash/crc32"
	"net"
	"strconv"
	"sync"
	"time"

	. "github.com/golang/protobuf/proto"
	"github.com/wpajqz/linker"
)

const MaxPayload = uint32(2048)

var (
	ErrPacketLength = errors.New("the packet is big than " + strconv.Itoa(int(MaxPayload)))
)

type Handler func(*Context)

type Client struct {
	mutex                   *sync.Mutex
	rwMutex                 *sync.RWMutex
	Context                 *Context
	timeout                 time.Duration
	conn                    net.Conn
	handlerContainer        map[int64]Handler
	packet                  chan linker.Packet
	cancelHeartbeat         chan bool
	closeClient             chan bool
	running                 chan bool
	OnConnectionStateChange func(status bool)
}

func NewClient() *Client {
	c := &Client{
		mutex:            new(sync.Mutex),
		rwMutex:          new(sync.RWMutex),
		Context:          &Context{Request: &request{}, Response: response{}},
		timeout:          30 * time.Second,
		packet:           make(chan linker.Packet, 1024),
		handlerContainer: make(map[int64]Handler),
		cancelHeartbeat:  make(chan bool, 1),
		closeClient:      make(chan bool, 1),
		running:          make(chan bool, 1),
	}

	return c
}

func (c *Client) Connect(network, address string) error {
	conn, err := net.Dial(network, address)
	if err != nil {
		return err
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
					// 把在线状态传递出去,方便调用方给用户提示信息
					if c.OnConnectionStateChange != nil {
						c.OnConnectionStateChange(r)
					}

					for {
						//服务端timeout设置影响链接延时时间
						conn, err := net.Dial(network, address)
						if err == nil {
							c.conn = conn
							if c.OnConnectionStateChange != nil {
								c.OnConnectionStateChange(true)
							}

							break
						}
					}
				}
			case <-c.closeClient:
				return
			}
		}
	}(network, address, c.conn)

	return nil
}

func (c *Client) StartHeartbeat(interval time.Duration, param Message) error {
	pbData, err := Marshal(param)
	if err != nil {
		return err
	}

	sequence := time.Now().UnixNano()
	p := linker.NewPack(linker.OPERATOR_HEARTBEAT, sequence, c.Context.Request.Header, pbData)

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

func (c *Client) StopHeartbeat() {
	c.cancelHeartbeat <- true
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
func (c *Client) SyncCall(operator string, param Message, onSuccess func(*Context), onError func(*Context)) error {
	pbData, err := Marshal(param)
	if err != nil {
		return err
	}

	nType := crc32.ChecksumIEEE([]byte(operator))
	sequence := time.Now().UnixNano()
	listener := int64(nType) + sequence

	// 对数据请求的返回状态进行处理,同步阻塞处理机制
	c.mutex.Lock()
	quit := make(chan bool)
	c.addMessageListener(listener, func(ctx *Context) {
		status := ctx.Response.GetResponseProperty("status")
		if status != "0" {
			onSuccess(ctx)
		} else {
			onError(ctx)
		}

		c.removeMessageListener(listener)
		quit <- true
	})

	p := linker.NewPack(nType, sequence, c.Context.Request.Header, pbData)
	c.packet <- p
	<-quit
	c.mutex.Unlock()

	return nil
}

// 向服务端发送请求，异步处理服务端返回结果
func (c *Client) AsyncCall(operator string, param Message, onSuccess func(*Context), onError func(*Context)) error {
	pbData, err := Marshal(param)
	if err != nil {
		return err
	}

	nType := crc32.ChecksumIEEE([]byte(operator))
	sequence := time.Now().UnixNano()

	listener := int64(nType) + sequence
	c.addMessageListener(listener, func(ctx *Context) {
		status := ctx.Response.GetResponseProperty("status")
		if status != "0" {
			onSuccess(ctx)
		} else {
			onError(ctx)
		}

		c.removeMessageListener(listener)
	})

	p := linker.NewPack(nType, sequence, c.Context.Request.Header, pbData)
	c.packet <- p

	return nil
}

// 添加事件监听器
func (c *Client) AddMessageListener(operator string, callback func(*Context)) {
	c.addMessageListener(int64(crc32.ChecksumIEEE([]byte(operator))), callback)
}

// 移除事件监听器
func (c *Client) RemoveMessageListener(operator string) {
	c.removeMessageListener(int64(crc32.ChecksumIEEE([]byte(operator))))
}

func (c *Client) addMessageListener(operator int64, callback func(*Context)) {
	c.rwMutex.Lock()
	c.handlerContainer[operator] = callback
	c.rwMutex.Unlock()
}

func (c *Client) removeMessageListener(operator int64) {
	c.rwMutex.Lock()
	delete(c.handlerContainer, operator)
	c.rwMutex.Unlock()
}
