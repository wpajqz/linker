package linker

import (
	"github.com/wpajqz/linker/codec"
)

const (
	OperatorHeartbeat = iota
	OperatorRegisterListener
	OperatorRemoveListener
	OperatorMax = 1024
)

const (
	errorTag = "error"
	nodeID   = "node_id"
)

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
		debug:       false,
		maxPayload:  1024 * 1024,
		contentType: codec.JSON,
	}

	for _, o := range opts {
		o(&options)
	}

	return &Server{options: options}
}

// 绑定路由
func (s *Server) BindRouter(r *Router) {
	s.registerInternalRouter(r)

	s.router = r
}

// 注册内部路由
func (s *Server) registerInternalRouter(r *Router) *Router {
	r.handlerContainer[OperatorRegisterListener] = HandlerFunc(func(ctx Context) {
		var topic string

		coder, err := codec.NewCoder(codec.String)
		if err != nil {
			ctx.Error(StatusInternalServerError, err.Error())
		}

		if err := coder.Decoder(ctx.RawBody(), &topic); err != nil {
			ctx.Error(StatusInternalServerError, err.Error())
		}

		if err := ctx.Subscribe(topic, func(bytes []byte) {
			if _, err := ctx.Write(topic, bytes); err != nil {
				ctx.Error(StatusInternalServerError, err.Error())
			}
		}); err != nil {
			ctx.Error(StatusInternalServerError, err.Error())
		}
	})

	r.handlerContainer[OperatorRemoveListener] = HandlerFunc(func(ctx Context) {
		var topic string

		coder, err := codec.NewCoder(codec.String)
		if err != nil {
			ctx.Error(StatusInternalServerError, err.Error())
		}

		if err := coder.Decoder(ctx.RawBody(), &topic); err != nil {
			ctx.Error(StatusInternalServerError, err.Error())
		}

		if err := ctx.UnSubscribe(topic); err != nil {
			ctx.Error(StatusInternalServerError, err.Error())
		}
	})

	return r
}

func (f HandlerFunc) Handle(ctx Context) {
	f(ctx)
}
