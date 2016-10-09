package linker

import (
	"fmt"
	"hash/crc32"
	"log"
	"net"
	"time"
)

const (
	MaxPayload = 2048
)

type (
	Handler      func(*Context)
	ErrorHandler func(SystemError)
	Server       struct {
		timeout          time.Duration
		handlerContainer map[uint32]Handler
		middleware       []Middleware
		routeMiddleware  map[string]Middleware
		int32Middleware  map[uint32][]Middleware
		MaxPayload       uint32
		protocolPacket   Packet
		errorHandler     ErrorHandler
	}
)

func NewServer() *Server {
	return &Server{
		MaxPayload:       MaxPayload,
		handlerContainer: make(map[uint32]Handler),
		routeMiddleware:  make(map[string]Middleware),
		int32Middleware:  make(map[uint32][]Middleware),
		errorHandler: func(err SystemError) {
			if err.level == "error" {
				log.Println("[" + err.level + "] " + err.err.Error())
			}
		},
	}
}

// 设置默认超时时间
func (s *Server) SetTimeout(timeout time.Duration) {
	s.timeout = timeout
}

// 设置可处理的数据包的最大长度
func (s *Server) SetMaxPayload(maxPayload uint32) {
	s.MaxPayload = maxPayload
}

// 设置服务端解析协议所使用的协议包规则
func (s *Server) SetProtocolPacket(packet Packet) {
	s.protocolPacket = packet
}

// 开始运行服务
func (s *Server) Run(name, address string) {
	listener, err := net.Listen(name, address)
	if err != nil {
		panic(err.Error())
	}

	defer listener.Close()

	fmt.Printf("%s server running on %s\n", name, address)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go s.handleConnection(conn)
	}
}

// 在服务中注册要处理的handler
func (s *Server) Handle(pattern string, handler Handler) {
	data := []byte(pattern)
	operator := crc32.ChecksumIEEE(data)

	_, ok := s.handlerContainer[operator]
	if !ok {
		s.handlerContainer[operator] = handler
	}
}

// 绑定Server需要处理的router
func (s *Server) BindRouter(routers []Router) {
	for _, router := range routers {
		data := []byte(router.Operator)
		operator := crc32.ChecksumIEEE(data)

		for _, m := range router.Middleware {
			if rm, ok := s.routeMiddleware[m]; ok {
				s.int32Middleware[operator] = append(s.int32Middleware[operator], rm)
			}
		}

		s.Handle(router.Operator, router.Handler)
	}
}

// 添加请求需要进行处理的中间件
func (s *Server) Use(middleware ...Middleware) {
	s.middleware = append(s.middleware, middleware...)
}

// 添加请求需要进行处理的中间件
func (s *Server) RouteMiddleware(routerMiddleware map[string]Middleware) {
	s.routeMiddleware = routerMiddleware
}

// 设置默认错误处理方法
func (s *Server) SetErrorHandler(errorHandler ErrorHandler) {
	s.errorHandler = errorHandler
}
