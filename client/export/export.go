package export

import (
	"bytes"
	"errors"
	"hash/crc32"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wpajqz/linker"
	"github.com/wpajqz/linker/plugin"
)

// Connection status
const (
	CONNECTING = 0 // 连接还没开启
	OPEN       = 1 // 连接已开启并准备好进行通信
	CLOSING    = 2 // 连接正在关闭的过程中
	CLOSED     = 3 // 连接已经关闭，或者连接无法建立
)

// Handler handle the connection
type Handler interface {
	Handle(header, body []byte)
}

// RequestStatusCallback 请求状态回调
type RequestStatusCallback interface {
	OnSuccess(header, body []byte)
	OnError(status int, message string)
	OnStart()
	OnEnd()
}

// ReadyStateCallback 链接状态回调
type ReadyStateCallback interface {
	OnOpen()
	OnClose()
	OnError(err string)
}

// Client 客户端结构体
type Client struct {
	conn                    net.Conn
	closed                  bool
	readyStateCallback      ReadyStateCallback
	readyState              int
	mutex                   *sync.Mutex
	rwMutex                 *sync.RWMutex
	timeout                 time.Duration
	handlerContainer        sync.Map
	packet                  chan linker.Packet
	pluginForPacketSender   []plugin.PacketPlugin
	pluginForPacketReceiver []plugin.PacketPlugin
	maxPayload              int
	request, response       struct {
		Header, Body []byte
	}
}

type HandlerFunc func(header, body []byte)

func (f HandlerFunc) Handle(header, body []byte) {
	f(header, body)
}

// NewClient 初始化客户端链接
func NewClient(address string, readyStateCallback ReadyStateCallback) (*Client, error) {
	c := &Client{
		readyState:       CONNECTING,
		mutex:            new(sync.Mutex),
		rwMutex:          new(sync.RWMutex),
		packet:           make(chan linker.Packet, 1024),
		handlerContainer: sync.Map{},
	}

	if readyStateCallback != nil {
		c.readyStateCallback = readyStateCallback
	}

	err := c.connect("tcp", address)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// NewUDPClient 初始化UDP客户端链接
func NewUDPClient(address string, readyStateCallback ReadyStateCallback) (*Client, error) {
	c := &Client{
		readyState:       CONNECTING,
		mutex:            new(sync.Mutex),
		rwMutex:          new(sync.RWMutex),
		packet:           make(chan linker.Packet, 1024),
		handlerContainer: sync.Map{},
	}

	if readyStateCallback != nil {
		c.readyStateCallback = readyStateCallback
	}

	err := c.connect("udp", address)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// GetReadyState 获取链接运行状态
func (c *Client) GetReadyState() int {
	return c.readyState
}

// Ping 心跳处理，客户端与服务端保持长连接
func (c *Client) Ping(param []byte, callback RequestStatusCallback) error {
	if callback == nil {
		return errors.New("callback can't be nil")
	}

	if c.readyState != OPEN {
		return errors.New("ping getsockopt: connection refuse")
	}

	sequence := time.Now().UnixNano()
	listener := int64(linker.OperatorHeartbeat) + sequence

	c.handlerContainer.Store(listener, HandlerFunc(func(header, body []byte) {
		code := c.GetResponseProperty("code")
		if code != "" {
			message := c.GetResponseProperty("message")
			v, _ := strconv.Atoi(code)
			callback.OnError(v, message)
		} else {
			callback.OnSuccess(header, body)
		}

		c.handlerContainer.Delete(listener)
	}))

	// 建立连接以后就发送心跳包建立会话信息，后面的定期发送
	p, err := linker.NewPacket(linker.OperatorHeartbeat, sequence, c.request.Header, param, c.pluginForPacketSender)
	if err != nil {
		return err
	}

	_, err = c.conn.Write(p.Bytes())
	if err != nil {
		return err
	}

	if c.timeout != 0 {
		err := c.conn.SetWriteDeadline(time.Now().Add(c.timeout))
		if err != nil {
			return err
		}
	}

	return nil
}

// SyncSend 向服务端发送请求，同步处理服务端返回结果
func (c *Client) SyncSend(operator string, param []byte, callback RequestStatusCallback) error {
	if callback == nil {
		return errors.New("callback can't be nil")
	}

	if c.readyState != OPEN {
		return errors.New("SyncSend getsockopt: connection refuse")
	}

	nType := crc32.ChecksumIEEE([]byte(operator))
	sequence := time.Now().UnixNano()
	listener := int64(nType) + sequence

	// 对数据请求的返回状态进行处理,同步阻塞处理机制
	c.mutex.Lock()
	quit := make(chan bool)

	callback.OnStart()

	c.handlerContainer.Store(listener, HandlerFunc(func(header, body []byte) {
		code := c.GetResponseProperty("code")
		if code != "" {
			message := c.GetResponseProperty("message")
			v, _ := strconv.Atoi(code)
			callback.OnError(v, message)
		} else {
			callback.OnSuccess(header, body)
		}

		callback.OnEnd()

		c.handlerContainer.Delete(listener)
		quit <- true
	}))

	p, err := linker.NewPacket(nType, sequence, c.request.Header, param, c.pluginForPacketSender)

	if err != nil {
		return err
	}

	c.packet <- p
	<-quit
	c.mutex.Unlock()

	return nil
}

// AsyncSend 向服务端发送请求，异步处理服务端返回结果
func (c *Client) AsyncSend(operator string, param []byte, callback RequestStatusCallback) error {
	if callback == nil {
		return errors.New("callback can't be nil")
	}

	if c.readyState != OPEN {
		return errors.New("AsyncSend getsockopt: connection refuse")
	}

	nType := crc32.ChecksumIEEE([]byte(operator))
	sequence := time.Now().UnixNano()

	listener := int64(nType) + sequence
	callback.OnStart()

	c.handlerContainer.Store(listener, HandlerFunc(func(header, body []byte) {
		code := c.GetResponseProperty("code")
		if code != "" {
			message := c.GetResponseProperty("message")
			v, _ := strconv.Atoi(code)
			callback.OnError(v, message)
		} else {
			callback.OnSuccess(header, body)
		}

		callback.OnEnd()

		c.handlerContainer.Delete(listener)
	}))

	p, err := linker.NewPacket(nType, sequence, c.request.Header, param, c.pluginForPacketSender)

	if err != nil {
		return err
	}

	c.packet <- p

	return nil
}

// AddMessageListener 添加事件监听器
func (c *Client) AddMessageListener(topic string, callback Handler) error {
	if callback == nil {
		return errors.New("callback can't be nil")
	}

	if c.readyState != OPEN {
		return errors.New("ping getsockopt: connection refuse")
	}

	sequence := time.Now().UnixNano()
	listener := int64(linker.OperatorRegisterListener) + sequence

	// 对数据请求的返回状态进行处理,同步阻塞处理机制
	c.mutex.Lock()
	quit := make(chan bool)

	var errRequest error
	c.handlerContainer.Store(listener, HandlerFunc(func(header, body []byte) {
		code := c.GetResponseProperty("code")
		if code != "" {
			message := c.GetResponseProperty("message")
			errRequest = errors.New(message)
		} else {
			c.handlerContainer.Store(int64(crc32.ChecksumIEEE([]byte(topic))), callback)
		}

		c.handlerContainer.Delete(listener)
		quit <- true
	}))

	p, err := linker.NewPacket(linker.OperatorRegisterListener, sequence, c.request.Header, []byte(topic), c.pluginForPacketSender)
	if err != nil {
		return err
	}

	c.packet <- p
	<-quit

	if errRequest != nil {
		return errRequest
	}

	c.mutex.Unlock()

	return nil
}

// SetMaxPayload 设置可处理的数据包的最大长度
func (c *Client) SetMaxPayload(maxPayload int) {
	c.maxPayload = maxPayload
}

// SetPluginForPacketSender 设置发送包需要的插件
func (c *Client) SetPluginForPacketSender(plugins ...plugin.PacketPlugin) {
	c.pluginForPacketSender = plugins
}

// pluginForPacketReceiver 设置接收包需要的插件
func (c *Client) SetPluginForPacketReceiver(plugins ...plugin.PacketPlugin) {
	c.pluginForPacketReceiver = plugins
}

// RemoveMessageListener 移除事件监听器
func (c *Client) RemoveMessageListener(listener string) {
	c.handlerContainer.Delete(int64(crc32.ChecksumIEEE([]byte(listener))))
}

// SetRequestProperty 设置请求属性
func (c *Client) SetRequestProperty(key, value string) {
	v := c.GetRequestProperty(key)
	old := []byte(key + "=" + v + ";")
	new := []byte("")

	c.request.Header = bytes.ReplaceAll(c.request.Header, old, new)
	c.request.Header = append(c.request.Header, []byte(key+"="+value+";")...)
}

// GetRequestProperty 获取请求属性
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

// GetResponseProperty 获取响应属性
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

// SetResponseProperty 设置响应属性
func (c *Client) SetResponseProperty(key, value string) {
	v := c.GetResponseProperty(key)
	old := []byte(key + "=" + v + ";")
	new := []byte("")

	c.response.Header = bytes.ReplaceAll(c.response.Header, old, new)
	c.response.Header = append(c.response.Header, []byte(key+"="+value+";")...)
}

// SetTimeout 设置服务端默认超时时间, 单位s
func (c *Client) SetTimeout(timeout int) {
	c.timeout = time.Duration(timeout) * time.Second
}

// Close 关闭链接
func (c *Client) Close() error {
	c.closed = true
	return c.conn.Close()
}

func (c *Client) connect(network, address string) error {
	var err error

	c.conn, err = net.Dial(network, address)
	if err != nil {
		return err
	}

	c.readyState = OPEN

	go c.handleConnection(network, c.conn)

	return nil
}
