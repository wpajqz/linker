package http

import (
	"time"

	"github.com/wpajqz/linker/client"
)

type Options struct {
	address     string
	timeout     time.Duration
	dialOptions []client.Option
}

type Option func(o *Options)

func Timeout(d time.Duration) Option {
	return func(o *Options) {
		o.timeout = d
	}
}

func DialOptions(opts ...client.Option) Option {
	return func(o *Options) {
		o.dialOptions = opts
	}
}
