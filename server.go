package linker

import (
	"github.com/wpajqz/linker/codec"
)

const (
	OperatorHeartbeat = iota
	OperatorMax       = 1024
)

const errorTag = "error"

type (
	Handler interface {
		Handle(Context)
	}
	HandlerFunc func(Context)
	Server      struct {
		options          Options
		router           *Router
		errorHandler     Handler
		constructHandler Handler
		destructHandler  Handler
		pingHandler      Handler
	}
)

func NewServer(opts ...Option) *Server {
	options := Options{
		Debug:       false,
		MaxPayload:  1024 * 1024,
		ContentType: codec.JSON,
	}

	for _, o := range opts {
		o(&options)
	}

	return &Server{options: options}
}

// 设置默认错误处理方法
func (s *Server) OnError(errorHandler Handler) {
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
