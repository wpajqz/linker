package client

import (
	"errors"
	"hash/crc32"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wpajqz/linker"
	"github.com/wpajqz/linker/coder"
)

const (
	MaxPayload = 2048
	TIMEOUT    = 30
)

var (
	ErrPacketLength = errors.New("the packet is big than " + strconv.Itoa(MaxPayload))
)

type Handler func(*Context)
type ErrorHandler func(error)

type RequestStatusCallback struct {
	OnSuccess      func(ctx *Context)
	OnError        func(status int, message string)
	OnStart, OnEnd func()
}

type Client struct {
	mutex                   *sync.Mutex
	rwMutex                 *sync.RWMutex
	Context                 *Context
	timeout                 time.Duration
	conn                    net.Conn
	handlerContainer        sync.Map
	packet                  chan linker.Packet
	OnConnectionStateChange func(status bool)
	constructHandler        Handler
	destructHandler         Handler
	errorHandler            ErrorHandler
}

func NewClient(server string, port int) *Client {
	address := strings.Join([]string{server, strconv.Itoa(port)}, ":")
	conn, err := net.Dial("tcp", address)
	if err != nil {
		panic(err)
	}

	c := &Client{
		conn:             conn,
		mutex:            new(sync.Mutex),
		rwMutex:          new(sync.RWMutex),
		Context:          &Context{Request: &Request{}, Response: Response{}},
		timeout:          TIMEOUT * time.Second,
		packet:           make(chan linker.Packet, 1024),
		handlerContainer: sync.Map{},
	}

	go c.handleConnection(c.conn)

	return c
}

func (c *Client) Ping(interval time.Duration, param interface{}) {
	t := c.Context.Request.GetRequestProperty("Content-Type")
	r, err := coder.NewCoder(t)
	if err != nil {
		panic(err)
	}

	pbData, err := r.Encoder(param)
	if err != nil {
		panic(err)
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
		}
	}
}

// 向服务端发送请求，同步处理服务端返回结果
func (c *Client) SyncSend(operator string, param interface{}, callback RequestStatusCallback) {
	t := c.Context.Request.GetRequestProperty("Content-Type")
	r, err := coder.NewCoder(t)
	if err != nil {
		panic(err)
	}

	pbData, err := r.Encoder(param)
	if err != nil {
		panic(err)
	}

	nType := crc32.ChecksumIEEE([]byte(operator))
	sequence := time.Now().UnixNano()
	listener := int64(nType) + sequence

	// 对数据请求的返回状态进行处理,同步阻塞处理机制
	c.mutex.Lock()
	quit := make(chan bool)

	if callback.OnStart != nil {
		callback.OnStart()
	}

	var handler Handler = func(ctx *Context) {
		code := ctx.Response.GetResponseProperty("code")
		if code != "" {
			message := ctx.Response.GetResponseProperty("message")
			if callback.OnError != nil {
				v, _ := strconv.Atoi(code)
				callback.OnError(v, message)
			}
		} else {
			if callback.OnSuccess != nil {
				callback.OnSuccess(ctx)
			}

			if callback.OnEnd != nil {
				callback.OnEnd()
			}
		}

		c.handlerContainer.Delete(listener)
		quit <- true
	}

	c.handlerContainer.Store(listener, handler)

	p := linker.NewPack(nType, sequence, c.Context.Request.Header, pbData)
	c.packet <- p
	<-quit
	c.mutex.Unlock()
}

// 向服务端发送请求，异步处理服务端返回结果
func (c *Client) AsyncSend(operator string, param interface{}, callback RequestStatusCallback) {
	t := c.Context.Request.GetRequestProperty("Content-Type")
	r, err := coder.NewCoder(t)
	if err != nil {
		panic(err)
	}

	pbData, err := r.Encoder(param)
	if err != nil {
		panic(err)
	}

	nType := crc32.ChecksumIEEE([]byte(operator))
	sequence := time.Now().UnixNano()

	listener := int64(nType) + sequence
	if callback.OnStart != nil {
		callback.OnStart()
	}

	var handler Handler = func(ctx *Context) {
		code := ctx.Response.GetResponseProperty("code")
		if code != "" {
			message := ctx.Response.GetResponseProperty("message")
			if callback.OnError != nil {
				v, _ := strconv.Atoi(code)
				callback.OnError(v, message)
			}
		} else {
			if callback.OnSuccess != nil {
				callback.OnSuccess(ctx)
			}

			if callback.OnEnd != nil {
				callback.OnEnd()
			}
		}

		c.handlerContainer.Delete(listener)
	}
	c.handlerContainer.Store(listener, handler)

	p := linker.NewPack(nType, sequence, c.Context.Request.Header, pbData)
	c.packet <- p
}

// 添加事件监听器
func (c *Client) AddMessageListener(listener string, callback Handler) {
	c.handlerContainer.Store(int64(crc32.ChecksumIEEE([]byte(listener))), callback)
}

// 移除事件监听器
func (c *Client) RemoveMessageListener(listener string) {
	c.handlerContainer.Delete(int64(crc32.ChecksumIEEE([]byte(listener))))
}

// 链接建立以后执行的操作
func (c *Client) OnOpen(handler Handler) {
	c.constructHandler = handler
}

// 链接断开以后执行回收操作
func (c *Client) OnClose(handler Handler) {
	c.destructHandler = handler
}

// 设置默认错误处理方法
func (c *Client) OnError(errorHandler ErrorHandler) {
	c.errorHandler = errorHandler
}

// 接收服务端返回的对心跳包的响应
func (c *Client) OnPong(pongHandler Handler) {
	c.handlerContainer.Store(linker.OPERATOR_HEARTBEAT, pongHandler)
}
