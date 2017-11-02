package export

import (
	"hash/crc32"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wpajqz/linker"
)

const (
	MaxPayload = 2048
	TIMEOUT    = 30
)

type request struct {
	net.Conn
	OperateType  uint32
	Sequence     int64
	Header, Body []byte
}

type response struct {
	net.Conn
	OperateType  uint32
	Sequence     int64
	Header, Body []byte
}

type Handler interface {
	Handle(header, body []byte)
}

type handlerFunc func(header, body []byte)

func (f handlerFunc) Handle(header, body []byte) {
	f(header, body)
}

type ErrorHandler interface {
	Handle(err string)
}

type RequestStatusCallback interface {
	OnSuccess(header, body []byte)
	OnError(status int, message string)
	OnStart()
	OnEnd()
}

type Client struct {
	mutex            *sync.Mutex
	rwMutex          *sync.RWMutex
	timeout          time.Duration
	conn             net.Conn
	handlerContainer sync.Map
	packet           chan linker.Packet
	constructHandler Handler
	destructHandler  Handler
	errorHandler     ErrorHandler
	request          request
	response         response
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
		timeout:          TIMEOUT * time.Second,
		packet:           make(chan linker.Packet, 1024),
		handlerContainer: sync.Map{},
	}

	go c.handleConnection(c.conn)

	return c
}

func (c *Client) Ping(interval int64, param []byte) {
	sequence := time.Now().UnixNano()
	p := linker.NewPack(linker.OPERATOR_HEARTBEAT, sequence, c.request.Header, param)

	// 建立连接以后就发送心跳包建立会话信息，后面的定期发送
	c.packet <- p
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for {
		select {
		case <-ticker.C:
			c.packet <- p
		}
	}
}

// 向服务端发送请求，同步处理服务端返回结果
func (c *Client) SyncSend(operator string, param []byte, callback RequestStatusCallback) {
	nType := crc32.ChecksumIEEE([]byte(operator))
	sequence := time.Now().UnixNano()
	listener := int64(nType) + sequence

	// 对数据请求的返回状态进行处理,同步阻塞处理机制
	c.mutex.Lock()
	quit := make(chan bool)

	if callback.OnStart != nil {
		callback.OnStart()
	}

	c.handlerContainer.Store(listener, handlerFunc(func(header, body []byte) {
		code := c.GetResponseProperty("code")
		if code != "" {
			message := c.GetResponseProperty("message")
			if callback.OnError != nil {
				v, _ := strconv.Atoi(code)
				callback.OnError(v, message)
			}
		} else {
			if callback.OnSuccess != nil {
				callback.OnSuccess(header, body)
			}

			if callback.OnEnd != nil {
				callback.OnEnd()
			}
		}

		c.handlerContainer.Delete(listener)
		quit <- true
	}))

	p := linker.NewPack(nType, sequence, c.request.Header, param)
	c.packet <- p
	<-quit
	c.mutex.Unlock()
}

// 向服务端发送请求，异步处理服务端返回结果
func (c *Client) AsyncSend(operator string, param []byte, callback RequestStatusCallback) {
	nType := crc32.ChecksumIEEE([]byte(operator))
	sequence := time.Now().UnixNano()

	listener := int64(nType) + sequence
	if callback.OnStart != nil {
		callback.OnStart()
	}

	c.handlerContainer.Store(listener, handlerFunc(func(header, body []byte) {
		code := c.GetResponseProperty("code")
		if code != "" {
			message := c.GetResponseProperty("message")
			if callback.OnError != nil {
				v, _ := strconv.Atoi(code)
				callback.OnError(v, message)
			}
		} else {
			if callback.OnSuccess != nil {
				callback.OnSuccess(header, body)
			}

			if callback.OnEnd != nil {
				callback.OnEnd()
			}
		}

		c.handlerContainer.Delete(listener)
	}))

	p := linker.NewPack(nType, sequence, c.request.Header, param)
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

// 接收服务端返回的对心跳包的响应
func (c *Client) SetRequestProperty(key, value string) {
	c.request.Header = append(c.request.Header, []byte(key+"="+value+";")...)
}

func (c *Client) GetRequestProperty(key string) string {
	values := strings.Split(string(c.request.Header), ";")
	for _, value := range values {
		kv := strings.Split(value, "=")
		if kv[0] == key {
			return kv[1]
		}
	}

	return ""
}

func (c *Client) GetResponseProperty(key string) string {
	values := strings.Split(string(c.response.Header), ";")
	for _, value := range values {
		kv := strings.Split(value, "=")
		if kv[0] == key {
			return kv[1]
		}
	}

	return ""
}
