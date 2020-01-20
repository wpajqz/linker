package linker

import (
	"net/http"
	"time"

	"github.com/wpajqz/linker/api"
	"github.com/wpajqz/linker/broker"
	"github.com/wpajqz/linker/plugin"
)

type (
	Options struct {
		debug                                                        bool
		readBufferSize                                               int
		writeBufferSize                                              int
		udpPayload                                                   int
		timeout                                                      time.Duration
		contentType                                                  string
		broker                                                       broker.Broker
		api                                                          api.API
		pluginForPacketSender                                        []plugin.PacketPlugin
		pluginForPacketReceiver                                      []plugin.PacketPlugin
		errorHandler, constructHandler, destructHandler, pingHandler Handler
		httpEndpoint, tcpEndpoint, udpEndpoint                       *Endpoint
	}

	Endpoint struct {
		Address string
		WSRoute string
		Handler http.Handler
	}

	Option func(o *Options)
)

func Debug() Option {
	return func(o *Options) {
		o.debug = true
	}
}

func ReadBufferSize(size int) Option {
	return func(o *Options) {
		o.readBufferSize = size
	}
}

func WriteBufferSize(size int) Option {
	return func(o *Options) {
		o.writeBufferSize = size
	}
}

func UDPPayload(size int) Option {
	return func(o *Options) {
		o.udpPayload = size
	}
}

func Timeout(d time.Duration) Option {
	return func(o *Options) {
		o.timeout = d
	}
}

func API(api api.API) Option {
	return func(o *Options) {
		o.api = api
	}
}

func ContentType(mime string) Option {
	return func(o *Options) {
		o.contentType = mime
	}
}

func Broker(broker broker.Broker) Option {
	return func(o *Options) {
		o.broker = broker
	}
}

func PluginForPacketSender(plugins ...plugin.PacketPlugin) Option {
	return func(o *Options) {
		o.pluginForPacketSender = append(o.pluginForPacketSender, plugins...)
	}
}

func PluginForPacketReceiver(plugins ...plugin.PacketPlugin) Option {
	return func(o *Options) {
		o.pluginForPacketReceiver = append(o.pluginForPacketReceiver, plugins...)
	}
}

func WithOnError(handler Handler) Option {
	return func(o *Options) {
		o.errorHandler = handler
	}
}

func WithOnClose(handler Handler) Option {
	return func(o *Options) {
		o.destructHandler = handler
	}
}

func WithOnOpen(handler Handler) Option {
	return func(o *Options) {
		o.constructHandler = handler
	}
}

func WithOnPing(handler Handler) Option {
	return func(o *Options) {
		o.pingHandler = handler
	}
}

func WithHTTPEndpoint(e Endpoint) Option {
	return func(o *Options) {
		o.httpEndpoint = &e
	}
}

func WithTCPEndpoint(e Endpoint) Option {
	return func(o *Options) {
		o.tcpEndpoint = &e
	}
}

func WithUDPEndpoint(e Endpoint) Option {
	return func(o *Options) {
		o.udpEndpoint = &e
	}
}
