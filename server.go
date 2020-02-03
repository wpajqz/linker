package linker

import (
	"github.com/wpajqz/linker/broker/memory"
	"github.com/wpajqz/linker/codec"
	"golang.org/x/sync/errgroup"
)

const (
	OperatorHeartbeat = iota
	OperatorRegisterListener
	OperatorRemoveListener
	OperatorMax = 1024
)

const (
	NetworkTCP = "tcp"
	NetworkUDP = "udp"
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

	Server struct {
		options Options
		router  *Router
	}
)

func NewServer(opts ...Option) *Server {
	options := Options{
		debug:       false,
		udpPayload:  4096,
		contentType: codec.JSON,
		broker:      memory.NewBroker(),
		tcpEndpoint: &Endpoint{Address: "localhost:8080"},
	}

	for _, o := range opts {
		o(&options)
	}

	return &Server{options: options}
}

func (s *Server) Run() error {
	var eg errgroup.Group

	if s.options.tcpEndpoint != nil {
		eg.Go(func() error {
			return s.runTCP(s.options.tcpEndpoint.Address)
		})
	}

	if s.options.httpEndpoint != nil {
		eg.Go(func() error {
			return s.runHTTP(s.options.httpEndpoint.Address, s.options.httpEndpoint.WSRoute, s.options.httpEndpoint.Handler)
		})
	}

	if s.options.udpEndpoint != nil {
		eg.Go(func() error {
			return s.runUDP(s.options.udpEndpoint.Address)
		})
	}

	return eg.Wait()
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
