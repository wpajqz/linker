package client

import (
	"time"

	"github.com/wpajqz/linker/plugin"
)

type (
	options struct {
		network                 string
		udpPayload              int
		dialTimeout             time.Duration
		initialCap              int
		maxCap                  int
		contentType             string
		idleTimeout             time.Duration
		onOpen, onClose         func()
		onError                 func(error)
		pluginForPacketSender   []plugin.PacketPlugin
		pluginForPacketReceiver []plugin.PacketPlugin
	}

	Option func(*options)
)

func Network(n string) Option {
	return Option(func(o *options) {
		o.network = n
	})
}

func UDPPayload(size int) Option {
	return func(o *options) {
		o.udpPayload = size
	}
}

func DialTimeout(n time.Duration) Option {
	return Option(func(o *options) {
		o.dialTimeout = n
	})
}

func ContentType(mime string) Option {
	return func(o *options) {
		o.contentType = mime
	}
}

func InitialCapacity(n int) Option {
	return Option(func(o *options) {
		o.initialCap = n
	})
}

func MaxCapacity(n int) Option {
	return Option(func(o *options) {
		o.maxCap = n
	})
}

func IdleTimeout(timeout time.Duration) Option {
	return Option(func(o *options) {
		o.idleTimeout = timeout
	})
}

func WithOnOpen(fn func()) Option {
	return Option(func(o *options) {
		o.onOpen = fn
	})
}

func WithOnClose(fn func()) Option {
	return Option(func(o *options) {
		o.onClose = fn
	})
}

func WithOnError(fn func(err error)) Option {
	return Option(func(o *options) {
		o.onError = fn
	})
}

func PluginForPacketSender(plugins ...plugin.PacketPlugin) Option {
	return func(o *options) {
		o.pluginForPacketSender = append(o.pluginForPacketSender, plugins...)
	}
}

func PluginForPacketReceiver(plugins ...plugin.PacketPlugin) Option {
	return func(o *options) {
		o.pluginForPacketReceiver = append(o.pluginForPacketReceiver, plugins...)
	}
}
