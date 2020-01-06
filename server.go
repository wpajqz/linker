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
		options Options
		router  *Router
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

// 绑定路由
func (s *Server) BindRouter(r *Router) {
	s.router = r
}

func (f HandlerFunc) Handle(ctx Context) {
	f(ctx)
}
