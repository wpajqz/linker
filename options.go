package linker

import (
	"time"

	"github.com/wpajqz/linker/broker"
	"github.com/wpajqz/linker/plugin"
)

type (
	Options struct {
		debug                                                        bool
		readBufferSize                                               int
		writeBufferSize                                              int
		timeout                                                      time.Duration
		maxPayload                                                   uint32
		contentType                                                  string
		broker                                                       broker.Broker
		pluginForPacketSender                                        []plugin.PacketPlugin
		pluginForPacketReceiver                                      []plugin.PacketPlugin
		errorHandler, constructHandler, destructHandler, pingHandler Handler
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

func Timeout(d time.Duration) Option {
	return func(o *Options) {
		o.timeout = d
	}
}

func MaxPayload(maxPayload uint32) Option {
	return func(o *Options) {
		o.maxPayload = maxPayload
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
