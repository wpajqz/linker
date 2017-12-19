package linker

import (
	"io"
	"log"
	"time"

	"github.com/wpajqz/linker/codec"
)

type (
	Handler interface {
		Handle(Context)
	}
	HandlerFunc  func(Context)
	ErrorHandler func(error)
	Server       struct {
		router           *Router
		debug            bool
		contentType      string
		timeout          time.Duration
		maxPayload       uint32
		errorHandler     ErrorHandler
		constructHandler Handler
		destructHandler  Handler
		pingHandler      Handler
	}
)

func NewServer() *Server {
	return &Server{
		contentType: codec.JSON,
		timeout:     TIMEOUT,
		maxPayload:  MaxPayload,
		errorHandler: func(err error) {
			if err != io.EOF {
				log.Println(err.Error())
			}
		},
	}
}

// 设置所有请求的序列化数据类型
func (s *Server) SetDebug(bool bool) {
	s.debug = bool
}

// 设置所有请求的序列化数据类型
func (s *Server) SetContentType(contentType string) {
	s.contentType = contentType
}

// 设置默认超时时间
func (s *Server) SetTimeout(timeout time.Duration) {
	s.timeout = timeout
}

// 设置可处理的数据包的最大长度
func (s *Server) SetMaxPayload(maxPayload uint32) {
	s.maxPayload = maxPayload
}

// 设置默认错误处理方法
func (s *Server) OnError(errorHandler ErrorHandler) {
	s.errorHandler = errorHandler
}

// 客户端链接断开以后执行回收操作
func (s *Server) OnClose(handler Handler) {
	s.destructHandler = handler
}

// 客户端建立连接以后初始化操作
func (s *Server) OnOpen(handler Handler) {
	s.constructHandler = handler
}

// 设置心跳包的handler,需要客户端发送心跳包才能够触发
// 客户端发送心跳包，服务端未调用此方法时只起到建立长连接的作用
func (s *Server) OnPing(handler Handler) {
	s.pingHandler = handler
}

// 绑定路由
func (s *Server) BindRouter(r *Router) {
	s.router = r
}

func (f HandlerFunc) Handle(ctx Context) {
	f(ctx)
}
